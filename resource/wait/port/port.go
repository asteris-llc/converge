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

	log "github.com/Sirupsen/logrus"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/wait"
)

// Port represents a port check
type Port struct {
	*resource.Status
	*wait.Retrier
	Host string
	Port int
	ConnectionCheck
}

// Check if the port is open
func (p *Port) Check(resource.Renderer) (resource.TaskStatus, error) {
	p.Status = resource.NewStatus()

	err := p.CheckConnection()
	if err == nil {
		if p.RetryCount > 0 {
			p.Status.AddMessage(fmt.Sprintf("Passed after %d retries (%v)", p.RetryCount, p.Duration))
		}
	} else {
		p.RaiseLevel(resource.StatusWillChange)
		p.Status.AddMessage(fmt.Sprintf("Failed to connect to %s:%d: %s", p.Host, p.Port, err.Error()))
		if p.RetryCount > 0 { // only add retry messages after an apply attempt
			p.Status.AddMessage(fmt.Sprintf("Failed after %d retries (%v)", p.RetryCount, p.Duration))
		}
	}

	return p, nil
}

// Apply retries the check until it passes or returns max failure threshold
func (p *Port) Apply() (resource.TaskStatus, error) {
	p.Status = resource.NewStatus()

	_, err := p.RetryUntil(func() (bool, error) {
		checkErr := p.CheckConnection()
		return checkErr == nil, checkErr
	})

	return p, err
}

// CheckConnection attempts to see if a tcp port is open
func (p *Port) CheckConnection() error {
	if p.ConnectionCheck == nil {
		p.ConnectionCheck = &TCPConnectionCheck{}
	}
	return p.ConnectionCheck.CheckConnection(p.Host, p.Port)
}

// ConnectionCheck represents a connection checker
type ConnectionCheck interface {
	CheckConnection(host string, port int) error
}

// TCPConnectionCheck impelements a ConnectionCheck over TCP
type TCPConnectionCheck struct{}

// CheckConnection attempts to see if a tcp port is open
func (t *TCPConnectionCheck) CheckConnection(host string, port int) error {
	logger := log.WithField("module", "wait.port")

	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		logger.WithError(err).WithField("addr", addr).Debug("connection failed")
		return err
	}
	defer conn.Close()

	return nil
}
