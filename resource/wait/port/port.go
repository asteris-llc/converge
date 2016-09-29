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
	Host        string
	Port        int
	GracePeriod time.Duration
	Interval    time.Duration
	MaxRetry    int
	RetryCount  int
	Duration    time.Duration
}

// Check if the port is open
func (p *Port) Check(resource.Renderer) (resource.TaskStatus, error) {
	p.Status = resource.NewStatus()

	alive, err := p.checkConnection()
	if err != nil {
		p.Status.RaiseLevel(resource.StatusFatal)
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
	startTime := time.Now()

	retries := p.MaxRetry
	if retries <= 0 {
		retries = defaultRetries
	}

	interval := p.Interval
	if interval <= 0 {
		interval = defaultInterval
	}

	after := p.GracePeriod
	alive := false
waitLoop:
	for {
		select {
		case <-time.After(after):
			if alive {
				break waitLoop
			}

			p.RetryCount++
			after = interval

			var err error
			alive, err = p.checkConnection()
			if err != nil {
				return p, err
			}

			if alive {
				after = p.GracePeriod
				continue
			}

			if p.RetryCount >= retries {
				break waitLoop
			}
		}
	}

	p.Duration = time.Since(startTime)
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
			logger.WithError(opErr).WithField("addr", addr).Info("connection failed")
			return false, nil
		}
		return false, errors.Wrapf(err, "dial failed")
	}
	defer conn.Close()

	return true, nil
}
