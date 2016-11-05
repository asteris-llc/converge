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
	"errors"
	"time"

	"github.com/asteris-llc/converge/load/registry"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/wait"
	"golang.org/x/net/context"
)

// Preparer handles wait.query tasks
type Preparer struct {
	// a host name or ip address. A TCP connection will be attempted at this host
	// and the specified Port.
	Host string `hcl:"host"`

	// the TCP port to attempt to connect to.
	Port int `hcl:"port" required:"true"`

	// the amount of time to wait in between checks. The format is Go's duration
	// string. A duration string is a possibly signed sequence of decimal numbers,
	// each with optional fraction and a unit suffix, such as "300ms", "-1.5h" or
	// "2h45m". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h". If
	// the interval is not specified, it will default to 5 seconds.
	Interval *time.Duration `hcl:"interval"`

	// the amount of time to wait before running the first check and after a
	// successful check. The format is Go's duration string. A duration string is
	// a possibly signed sequence of decimal numbers, each with optional fraction
	// and a unit suffix, such as "300ms", "-1.5h" or "2h45m". Valid time units
	// are "ns", "us" (or "µs"), "ms", "s", "m", "h". If no grace period is
	// specified, no grace period will be taken into account.
	GracePeriod *time.Duration `hcl:"grace_period"`

	// the maximum number of attempts before the wait fails. If the maximum number
	// of retries is not set, it will default to 5.
	MaxRetry *int `hcl:"max_retry"`
}

// Prepare creates a new wait.port type
func (p *Preparer) Prepare(ctx context.Context, render resource.Renderer) (resource.Task, error) {
	if p.Port <= 0 {
		return nil, errors.New("port is required and must be greater than zero")
	}
	port := &Port{
		Host:    p.Host,
		Port:    p.Port,
		Retrier: wait.PrepareRetrier(p.Interval, p.GracePeriod, p.MaxRetry),
	}
	return port, nil
}

func init() {
	registry.Register("wait.port", (*Preparer)(nil), (*Port)(nil))
}
