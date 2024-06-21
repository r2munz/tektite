// Copyright 2024 The Tektite Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package testutils

import (
	"context"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"sync"
)

type EtcdHolder struct {
	lock      sync.Mutex
	started   bool
	container testcontainers.Container
	address   string
}

func (e *EtcdHolder) Stop() {
	e.lock.Lock()
	defer e.lock.Unlock()
	if !e.started {
		return
	}
	if err := e.container.Stop(context.Background(), nil); err != nil {
		panic(err)
	}
	e.started = false
}

func (e *EtcdHolder) Address() string {
	return e.address
}

func CreateEtcdContainer() (*EtcdHolder, error) {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "gcr.io/etcd-development/etcd:v3.5.10",
		WaitingFor:   wait.ForListeningPort("2379"),
		ExposedPorts: []string{"2379/tcp"},
		Env: map[string]string{
			"ETCD_LOG_LEVEL": "debug",
		},
		Cmd: []string{"etcd", "--advertise-client-urls", "http://0.0.0.0:2379", "--listen-client-urls", "http://0.0.0.0:2379",
			"--data-dir", "/tmp/tektite-test-etcd-data"},
	}
	etcdContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
	})
	if err != nil {
		return nil, err
	}
	if err := etcdContainer.Start(ctx); err != nil {
		return nil, err
	}
	host, err := etcdContainer.Host(ctx)
	if err != nil {
		return nil, err
	}
	np := nat.Port("2379/tcp")
	port, err := etcdContainer.MappedPort(ctx, np)
	if err != nil {
		return nil, err
	}
	return &EtcdHolder{
		started:   true,
		container: etcdContainer,
		address:   fmt.Sprintf("%s:%d", host, port.Int()),
	}, nil
}
