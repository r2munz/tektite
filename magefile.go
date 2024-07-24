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
	// Unmarshal the JSON data into a map
	var certsConfig map[string]interface{}
	err = json.Unmarshal(bytes, &certsConfig)
	if err != nil {
		return fmt.Errorf("could not unmarshal JSON: %w", err)
	}
	var caEnv string
	var caSignedServerEnv string
	var caSignedClientEnv string
	var selfSignedServerEnv string
	var selfSignedClientEnv string
	var selfSignedClient2Env string
	var adminCertEnv string
	var apiCertEnv string
	var integrationCertEnv string
	var remotingCertEnv string
	var shutdownCertEnv string
	var tektclientCertEnv string
	for k, v := range certsConfig {
		val, _ := v.(map[string]interface{})
		switch k {
		case "CA":
			var Env []string
			for _, i := range val["paths"].(map[string]interface{}) {
				Env = append(Env, i.(string))
			}
			Env = append(Env, val["crt"].(string))
			caEnv = strings.Join(Env, "/")
		case "CA_SIGNED_SERVER":
			var Env []string
			for _, i := range val["paths"].(map[string]interface{}) {
				Env = append(Env, i.(string))
			}
			Env = append(Env, val["crt"].(string))
			caSignedServerEnv = strings.Join(Env, "/")
		case "CA_SIGNED_CLIENT":
			var Env []string
			for _, i := range val["paths"].(map[string]interface{}) {
				Env = append(Env, i.(string))
			}
			Env = append(Env, val["crt"].(string))
			caSignedClientEnv = strings.Join(Env, "/")
		case "SELF_SIGNED_SERVER":
			var Env []string
			for _, i := range val["paths"].(map[string]interface{}) {
				Env = append(Env, i.(string))
			}
			Env = append(Env, val["crt"].(string))
			selfSignedServerEnv = strings.Join(Env, "/")
		case "SELF_SIGNED_CLIENT":
			var Env []string
			for _, i := range val["paths"].(map[string]interface{}) {
				Env = append(Env, i.(string))
			}
			Env = append(Env, val["crt"].(string))
			selfSignedClientEnv = strings.Join(Env, "/")
		case "SELF_SIGNED_CLIENT2":
			var Env []string
			for _, i := range val["paths"].(map[string]interface{}) {
				Env = append(Env, i.(string))
			}
			Env = append(Env, val["crt"].(string))
			selfSignedClient2Env = strings.Join(Env, "/")
		case "SERVER_UTILS":
			var adminEnv []string
			var apiEnv []string
			var integrationEnv []string
			var remotingEnv []string
			var shutdownEnv []string
			var tektclientEnv []string
			for p, i := range val["paths"].(map[string]interface{}) {
				switch p {
				case "admin":
					adminEnv = append(adminEnv, i.(string))
				case "api":
					apiEnv = append(apiEnv, i.(string))
				case "integration":
					integrationEnv = append(integrationEnv, i.(string))
				case "remoting":
					remotingEnv = append(remotingEnv, i.(string))
				case "shutdown":
					shutdownEnv = append(shutdownEnv, i.(string))
				case "tektclient":
					tektclientEnv = append(tektclientEnv, i.(string))
				}
			}
			adminEnv = append(adminEnv, val["crt"].(string))
			apiEnv = append(apiEnv, val["crt"].(string))
			integrationEnv = append(integrationEnv, val["crt"].(string))
			remotingEnv = append(remotingEnv, val["crt"].(string))
			shutdownEnv = append(shutdownEnv, val["crt"].(string))
			tektclientEnv = append(tektclientEnv, val["crt"].(string))

			adminCertEnv = strings.Join(adminEnv, "/")
			apiCertEnv = strings.Join(apiEnv, "/")
			integrationCertEnv = strings.Join(integrationEnv, "/")
			remotingCertEnv = strings.Join(remotingEnv, "/")
			shutdownCertEnv = strings.Join(shutdownEnv, "/")
			tektclientCertEnv = strings.Join(tektclientEnv, "/")
		}
	}
	return sh.RunV("./certsCheck.sh", caEnv, caSignedServerEnv, caSignedClientEnv, selfSignedServerEnv, selfSignedClientEnv, selfSignedClient2Env, adminCertEnv, apiCertEnv, integrationCertEnv, remotingCertEnv, shutdownCertEnv, tektclientCertEnv)
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
