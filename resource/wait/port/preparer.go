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
	"time"

	"github.com/asteris-llc/converge/load/registry"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/wait"
)

// Preparer handles wait.query tasks
type Preparer struct {
	// a host name or ip address. A TCP connection will be attempted at this host
	// and the specified Port.
	Host string `hcl:"host"`

	// the TCP port to attempt to connect to.
	Port int `hcl:"port" required:"true"`

	// the amount of time to wait in between checks. The format is Go's duraction
	// string. A duration string is a possibly signed sequence of decimal numbers,
	// each with optional fraction and a unit suffix, such as "300ms", "-1.5h" or
	// "2h45m". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
	Interval string `hcl:"interval" doc_type:"duration string"`

	// the amount of time to wait before running the first check and after a
	// successful check. The format is Go's duraction string. A duration string is
	// a possibly signed sequence of decimal numbers, each with optional fraction
	// and a unit suffix, such as "300ms", "-1.5h" or "2h45m". Valid time units
	// are "ns", "us" (or "µs"), "ms", "s", "m", "h".
	GracePeriod string `hcl:"grace_period" doc_type:"duration string"`

	// the maximum number of attempts before the wait fails.
	MaxRetry int `hcl:"max_retry"`
}

// Prepare creates a new wait.port type
func (p *Preparer) Prepare(render resource.Renderer) (resource.Task, error) {
	host, err := render.Render("host", p.Host)
	if err != nil {
		return nil, err
	}

	port := &Port{
		Host: host,
		Port: p.Port,
		Retrier: &wait.Retrier{
			MaxRetry: p.MaxRetry,
		},
	}

	if intervalDuration, perr := time.ParseDuration(p.Interval); perr == nil {
		port.Interval = intervalDuration
	}

	if gracePeriodDuration, perr := time.ParseDuration(p.GracePeriod); perr == nil {
		port.GracePeriod = gracePeriodDuration
	}

	return port, nil
}

func init() {
	registry.Register("wait.port", (*Preparer)(nil), (*Port)(nil))
}
