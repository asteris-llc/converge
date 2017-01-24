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

package unit

import (
	"github.com/asteris-llc/converge/load/registry"
	"github.com/asteris-llc/converge/resource"
	"golang.org/x/net/context"
)

// Preparer for UnitState
//
// UnitState configures loaded systemd units, allowing you to start, stop, or
// restart them, reload configuration files, and send unix signals.
type Preparer struct {
	// The name of the unit.  This may optionally include the unit type,
	// e.g. "foo.service" and "foo" are both valid.
	Name string `hcl:"unit" required:"true"`

	// The desired state of the unit.  This will affect the current unit job.  Use
	// `systemd.unit.file` to enable and disable jobs, or `systemd.unit.config` to
	// set options.
	State string `hcl:"state" valid_values:"running,stopped,restarted"`

	// If reload is true then the service will be instructed to reload it's
	// configuration as if the user had run `systemctl reload`.  This will reload
	// the actual confguration file for the service, not the systemd unit file
	// configuration. See `systemctl(1)` for more information.
	Reload bool `hcl:"reload"`

	// Sends a signal to the process, using it's name.  The signal may be in upper
	// or lower case (the `SIG` prefix) is optional.  For example, to send user
	// defined signal 1 to the process you may write any of: usr1, USR1, SIGUSR1,
	// or sigusr1
	//
	// see `signal(3)` on BSD/Darwin, or `signal(7)` on GNU Linux systems for more
	// information on signals
	SignalName string `hcl:"signal_name" mutually_exclusive:"signal_name,signal_num"`

	// Sends a signal to the process, using it's signal number.  The value must be
	// an unsigned integer value between 1 and 31 inclusive.
	SignalNumber uint `hcl:"signal_number" mutually_exclusive:"signal_name,signal_num"`
}

// Prepare a new task
func (p *Preparer) Prepare(ctx context.Context, render resource.Renderer) (resource.Task, error) {
	var signal *Signal
	if p.SignalName != "" {
		num, err := ParseSignalByName(p.SignalName)
		if err != nil {
			return nil, err
		}
		signal = &num
	} else if p.SignalNumber != 0 {
		name, err := ParseSignalByNumber(p.SignalNumber)
		if err != nil {
			return nil, err
		}
		signal = &name
	}

	executor, err := realExecutor()

	if err != nil {
		return nil, err
	}

	r := &Resource{
		Reload:          p.Reload,
		Name:            p.Name,
		State:           p.State,
		systemdExecutor: executor,
	}

	if signal != nil {
		r.SignalName = signal.String()
		r.SignalNumber = uint(*signal)
		r.sendSignal = true
	}

	return r, nil
}

func init() {
	registry.Register("systemd.unit.state", (*Preparer)(nil), (*Resource)(nil))
}
