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
	"fmt"
	"time"

	"github.com/asteris-llc/converge/load/registry"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/systemd"
)

// Preparer for Systemd Unit
type Preparer struct {
	// The name of this unit as in `foo.service`
	// (either just file names or full absolute paths if the unit files are
	// residing outside the usual unit search paths
	Name string `hcl:"name" required:"true"`

	// Active determines whether this unit should be in an active or inactive state
	Active bool `hcl:"active" required:"true"`

	/* state is the `UnitFileState` of the unit.
	`UnitFileState` encodes the install state of the unit file of FragmentPath.
	It currently knows the following states: `enabled`, `enabled-runtime`, `linked`,
	`linked-runtime`, `masked`, `masked-runtime`, `static`, `bad`, `disabled`, `invalid`.

	Of these states only `enabled`, `enabled-runtime`, `linked`,
	`linked-runtime`, `masked`, `masked-runtime`, `static`, and `disabled`
	are permitted in the preparer. If you would like your unit to be
	run at runtime, put the unit file in the /run/systemd/system folder, and
	only use states with the `-runtime` suffix. Once a unit has
	a runtime state, converge cannot change it to be used through /etc/. Static
	units, units that have no `[Install]` section, should be disabled with `masked` or
	`masked-runtime`. Converge will not throw a warning if you want your unit to be
	`enabled`, when it is infact `static`, and technically cannot be `enabled`,
	since the unit will be "enabled" either way.
	Likewise it will not throw a warning if you want your unit disabled, but it is
	actually masked.

	See [dbus](https://godoc.org/github.com/coreos/go-systemd/dbus) for more info
	*/
	UnitFileState string `hcl:"state" required:"false" valid_values:"enabled,enabled-runtime,linked,linked-runtime,masked,masked-runtime,disabled"`

	/* Mode for the call to StartUnit()
	StartUnit() enqeues a start job, and possibly depending jobs.
	Takes the unit to activate, plus a mode string. The mode needs to be one of
	replace, fail, isolate, ignore-dependencies, ignore-requirements.

	See [dbus](https://godoc.org/github.com/coreos/go-systemd/dbus) for more info
	*/
	StartMode string `hcl:"mode" required:"false" valid_values:"replace,fail,isolate,ignore-dependencies,ignore-requirements"`

	// the amount of time the command will wait for configuration to load
	// before halting forcefully. The
	// format is Go's duraction string. A duration string is a possibly signed
	// sequence of decimal numbers, each with optional fraction and a unit
	// suffix, such as "300ms", "-1.5h" or "2h45m". Valid time units are "ns",
	// "us" (or "µs"), "ms", "s", "m", "h".
	Timeout string `hcl:"timeout" doc_type:"duration string" required:"false"`
}

//Prepare returns a new systemd.unit task
func (p *Preparer) Prepare(render resource.Renderer) (resource.Task, error) {
	var t time.Duration
	var err error
	if p.Timeout == "" {
		t = systemd.DefaultTimeout
	} else {
		t, err = time.ParseDuration(p.Timeout)
		if err != nil {
			return nil, err
		}
	}
	// Handle Defaults
	if p.StartMode == "" {
		p.StartMode = string(systemd.SMReplace)
	}

	unit := &Unit{
		Name:          p.Name,
		Active:        p.Active,
		UnitFileState: systemd.UnitFileState(p.UnitFileState),
		StartMode:     systemd.StartMode(p.StartMode),
		Timeout:       t,
	}
	if unit.Name == "" {
		return unit, fmt.Errorf("task requires a %q parameter", "name")
	}
	return unit, nil
}

func init() {
	registry.Register("systemd.unit", (*Preparer)(nil), (*Unit)(nil))
}
