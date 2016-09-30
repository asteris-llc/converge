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

package port

import (
	"fmt"
	"net"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/wait"
	"github.com/pkg/errors"
)

const (
	defaultInterval = 5 * time.Second
	defaultTimeout  = 10 * time.Second
	defaultRetries  = 5
	defaultHost     = "localhost"
)

// Port represents a port check
type Port struct {
	*resource.Status
	*wait.Retrier
	Host string
	Port int
}

// Check if the port is open
func (p *Port) Check(resource.Renderer) (resource.TaskStatus, error) {
	p.Status = resource.NewStatus()

	alive, err := p.checkConnection()
	if err != nil {
		return p, errors.Wrapf(err, "failed to check connection")
	}

	if !alive {
		p.RaiseLevel(resource.StatusWillChange)
	}

	return p, nil
}

// Apply retries the check until it passes or returns max failure threshold
func (p *Port) Apply() (resource.TaskStatus, error) {
	p.Status = resource.NewStatus()

	_, err := p.RetryUntil(p.checkConnection)
	if err != nil {
		return p, errors.Wrapf(err, "failed to check connection")
	}

	return p, nil
}

func (p *Port) checkConnection() (connected bool, err error) {
	logger := log.WithField("module", "wait.port")

	if p.Host == "" {
		p.Host = defaultHost
	}

	addr := fmt.Sprintf("%s:%d", p.Host, p.Port)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		if opErr, ok := err.(*net.OpError); ok {
			logger.WithError(opErr).WithField("addr", addr).Debug("connection failed")
			return false, nil
		}
		return false, errors.Wrapf(err, "dial failed")
	}
	defer conn.Close()

	return true, nil
}
