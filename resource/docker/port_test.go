// Copyright © 2016 Asteris, LLC
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

// +build !solaris

package docker_test

import (
	"testing"

	"github.com/asteris-llc/converge/resource/docker"
	"github.com/stretchr/testify/assert"
)

func TestDockerPortPort(t *testing.T) {
	port := "80"
	dockerPort := docker.NewPort(port)
	assert.Equal(t, "80", dockerPort.PortNum())
}

func TestDockerPortDefaultProtocol(t *testing.T) {
	port := "80"
	dockerPort := docker.NewPort(port)
	assert.Equal(t, "tcp", dockerPort.Proto())
}

func TestDockerPortProtocol(t *testing.T) {
	port := "53/udp"
	dockerPort := docker.NewPort(port)
	assert.Equal(t, "udp", dockerPort.Proto())
}

func TestDockerPortSring(t *testing.T) {
	port := "80"
	dockerPort := docker.NewPort(port)
	assert.Equal(t, "80/tcp", dockerPort.String())
}
