//go:build mage

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/spirit-labs/tektite/kafkagen"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

const (
	GotestsumUrl    = "gotest.tools/gotestsum"
	GolangciLintUrl = "github.com/golangci/golangci-lint/cmd/golangci-lint"
	AddLicenseUrl   = "github.com/google/addlicense"
)

var (
	goexec = mg.GoCmd()
	g0     = sh.RunCmd(goexec)
)

// Build builds the binary
func Build() error {
	fmt.Println("Building the binary...")
	return g0("build", "-o", "bin", "./...")
}

func mustRun(cmd string, args ...string) {
	out := lipgloss.NewStyle().Bold(true).Render(
		fmt.Sprintf("\n> %s %s\n", cmd, strings.Join(args, " ")),
	)

	fmt.Println(out)
	if err := sh.RunV(cmd, args...); err != nil {
		panic(err)
	}
}

func checkTools() error {
	if _, err := exec.LookPath("gotestsum"); err != nil {
		fmt.Println("gotestsum is not installed. Installing...")
		fmt.Printf("Installing gotestsum from %s\n", GotestsumUrl)
		mustRun(goexec, "install", GotestsumUrl)
	}

	if _, err := exec.LookPath("golangci-lint"); err != nil {
		fmt.Println("golangci-lint not found, installing...")
		fmt.Printf("Installing golangci-lint from %s\n", GolangciLintUrl)
		mustRun(goexec, "install", GolangciLintUrl)
	}

	if _, err := exec.LookPath("addlicense"); err != nil {
		fmt.Println("addlicense not found, installing...")
		fmt.Printf("Installing addlicense from %s\n", AddLicenseUrl)
		mustRun(goexec, "install", AddLicenseUrl)
	}
	return nil
}

// Lint runs the linter
func Lint() error {
	mg.Deps(checkTools)
	fmt.Println("Running golanci-lint linter...")
	return sh.RunV("golangci-lint", "run")
}

// Test runs only the unit tests
func Test() error {
	mg.Deps(checkTools)
	fmt.Println("Running unit tests...")
	return sh.RunV("gotestsum", "-f", "standard-verbose", "--", "-race", "-failfast", "-count", "1", "-timeout", "10m", "./...")
}

// Test runs only the integration tests
func Integration() error {
	mg.Deps(checkTools)
	fmt.Println("Running integration tests...")
	return sh.RunV("gotestsum", "-f", "standard-verbose", "--", "-tags", "integration", "-race", "-failfast", "-count", "1", "-timeout", "10m", "./integration/...")
}

// LicenseCheck fixes any missing license header in the source code
func LicenseCheck() error {
	mg.Deps(checkTools)
	fmt.Println("Running license check...")
	return sh.RunV("addlicense", "-c", "The Tektite Authors", "-ignore", "**/*.yml", "-ignore", "**/*.xml", ".")
}

// Presubmit is intended to be run by contributors before pushing the code and creating a PR.
// It depends on LicenseCheck, Build, Lint, Test and Integration in order
func Presubmit() error {
	mg.Deps(LicenseCheck, Build, Lint, Test)
	return Integration()
}

// Run tests in a continuous loop
func Loop() error {
	iteration := 0
	logFile, err := os.OpenFile("test-results.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %v", err)
	}
	defer logFile.Close()

	logWriter := bufio.NewWriter(logFile)
	defer logWriter.Flush()

	// Write the current date and time to the log file
	date := time.Now().Format(time.RFC1123)
	logWriter.WriteString(date + "\n")

	for {
		iteration++
		iter := fmt.Sprintf("Running loop iteration %d", iteration)
		fmt.Println(iter)
		logWriter.WriteString(iter + "\n")

		ran, err := sh.Exec(nil, logWriter, logWriter, "gotestsum", "-f", "standard-verbose", "--", "-tags", "integration", "-race", "-failfast", "-count", "1", "-timeout", "7m", "./integration/...")

		if !ran {
			errString := fmt.Sprintf("Go test failed. Exiting loop. Error: %v\n", err)
			fmt.Println(errString)
			logWriter.WriteString(errString)
			break
		}
	}
	return nil
}

// Run tektited in a standalone setup
func Run() error {
	fmt.Println("Running tektited in a standalone setup...")
	return g0("run", "cmd/tektited/main.go", "--config", "cfg/standalone.conf")
}

// Check internal certificates
func CheckCerts() error {
	fmt.Println("Checking internal certificates expiration dates")
	certsConfigPath := "cli/testdata/certsConfig.json"
	// Open the JSON file
	file, err := os.Open(certsConfigPath)
	if err != nil {
		return fmt.Errorf("could not open file: %w", err)
	}
	defer file.Close()
	// Read the file's content
	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return fmt.Errorf("could not read file: %w", err)
	}

	type Ca struct {
		Bits      int      `json:"bits"`
		Crt       string   `json:"crt"`
		Days      int      `json:"days"`
		Paths     []string `json:"paths"`
		Pubkey    string   `json:"pubkey"`
		SignedCrt string   `json:"signedCrt"`
		Subject   string   `json:"subject"`
	}
	type CaSigned struct {
		Crt     string   `json:"crt"`
		Days    int      `json:"days"`
		ExtFile string   `json:"extFile"`
		Key     string   `json:"key"`
		Paths   []string `json:"paths"`
		Req     string   `json:"req"`
		Subject string   `json:"subject"`
	}
	type SelfSigned struct {
		ConfigFile string   `json:"configFile"`
		Crt        string   `json:"crt"`
		Days       int      `json:"days"`
		Key        string   `json:"key"`
		Paths      []string `json:"paths"`
		PkeyOpt    string   `json:"pkeyOpt"`
		Subject    string   `json:"subject"`
	}
	type CertsConfig struct {
		Ca                *Ca         `json:"ca"`
		CaSignedServer    *CaSigned   `json:"caSignedServer"`
		CaSignedClient    *CaSigned   `json:"caSignedClient"`
		SelfSignedServer  *SelfSigned `json:"selfSignedServer"`
		SelfSignedClient  *SelfSigned `json:"selfSignedClient"`
		SelfSignedClient2 *SelfSigned `json:"selfSignedClient2"`
		ServerUtils       *SelfSigned `json:"serverUtils"`
	}
	var config CertsConfig
	err = json.Unmarshal(bytes, &config)
	if err != nil {
		return fmt.Errorf("could not unmarshal JSON: %s\n", err)
	}
	caEnv := strings.Join([]string{config.Ca.Paths[0], config.Ca.Crt}, "/")
	caSignedServerEnv := strings.Join([]string{config.CaSignedServer.Paths[0], config.CaSignedServer.Crt}, "/")
	caSignedClientEnv := strings.Join([]string{config.CaSignedClient.Paths[0], config.CaSignedClient.Crt}, "/")
	selfSignedServerEnv := strings.Join([]string{config.SelfSignedServer.Paths[0], config.SelfSignedServer.Crt}, "/")
	selfSignedClientEnv := strings.Join([]string{config.SelfSignedClient.Paths[0], config.SelfSignedClient.Crt}, "/")
	selfSignedClient2Env := strings.Join([]string{config.SelfSignedClient2.Paths[0], config.SelfSignedClient2.Crt}, "/")
	adminEnv := strings.Join([]string{config.ServerUtils.Paths[0], config.ServerUtils.Crt}, "/")
	apiEnv := strings.Join([]string{config.ServerUtils.Paths[1], config.ServerUtils.Crt}, "/")
	integrationEnv := strings.Join([]string{config.ServerUtils.Paths[2], config.ServerUtils.Crt}, "/")
	remotingEnv := strings.Join([]string{config.ServerUtils.Paths[3], config.ServerUtils.Crt}, "/")
	shutdownEnv := strings.Join([]string{config.ServerUtils.Paths[4], config.ServerUtils.Crt}, "/")
	tektclientEnv := strings.Join([]string{config.ServerUtils.Paths[5], config.ServerUtils.Crt}, "/")

	Envs := []string{caEnv, caSignedServerEnv, caSignedClientEnv, selfSignedServerEnv, selfSignedClientEnv, selfSignedClient2Env, adminEnv, apiEnv, integrationEnv, remotingEnv, shutdownEnv, tektclientEnv}

	for e := range Envs {
		env := Envs[e]
		opensslCmd := fmt.Sprintf(`openssl x509 -enddate -noout -in "%s"|cut -d= -f 2`, env)
		output, err := sh.Output("sh", "-c", opensslCmd)
		if err != nil {
			return fmt.Errorf("could not execute command")
		}
		fmt.Println(output, ": ", env)
	}
	return nil
}

// Renew internal certificates
/*
func RenewCerts() error {
	fmt.Println("Renewing internal certificates")
	return
}
*/

// GenKafkaProtocol generates the Kafka protocol code from the protocol JSON descriptors
func GenKafkaProtocol() error {
	return kafkagen.Generate("asl/kafka/spec", "kafkaserver/protocol")
}
