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
	"github.com/asteris-llc/converge/resource/file/content"
	"github.com/asteris-llc/converge/resource/systemd"
)

// Preparer for Systemd Unit
type Preparer struct {
	// The name of this unit as in "foo.service"
	Name string `hcl:"name"`

	// Active determines whether this unit should be active or inactive
	Active string `hcl:"active"`

	/* State is the UnitFileState of the unit.
	UnitFileState encodes the install state of the unit file of FragmentPath.
	It currently knows the following states: enabled, enabled-runtime, linked,
	linked-runtime, masked, masked-runtime, static, disabled, invalid. enabled
	indicates that a unit file is permanently enabled. enable-runtime indicates
	the unit file is only temporarily enabled, and will no longer be enabled
	after a reboot (that means, it is enabled via /run symlinks, rather than /etc).
	linked indicates that a unit is linked into /etc permanently, linked indicates
	that a unit is linked into /run temporarily (until the next reboot). masked
	indicates that the unit file is masked permanently, masked-runtime indicates
	that it is only temporarily masked in /run, until the next reboot.
	static indicates that the unit is statically enabled, i.e. always enabled and
	doesn't need to be enabled explicitly. invalid indicates that it could not
	be determined whether the unit file is enabled.
	*/
	UnitFileState string `hcl:"state"`

	/* Mode for the call to StartUnit()
	StartUnit() enqeues a start job, and possibly depending jobs.
	Takes the unit to activate, plus a mode string. The mode needs to be one of
	replace, fail, isolate, ignore-dependencies, ignore-requirements.
	If "replace" the call will start the unit and its dependencies,
	possibly replacing already queued jobs that conflict with this. If "fail" the
	call will start the unit and its dependencies, but will fail if this would
	change an already queued job. If "isolate" the call will start the unit in
	question and terminate all units that aren't dependencies of it. If
	"ignore-dependencies" it will start a unit but ignore all its dependencies.
	If "ignore-requirements" it will start a unit but only ignore the requirement
	dependencies. It is not recommended to make use of the latter two options.
	Returns the newly created job object.
	*/
	StartMode string `hcl:"mode"`

	// the amount of time the command will wait for configuration to load
	// before halting forcefully. The
	// format is Go's duraction string. A duration string is a possibly signed
	// sequence of decimal numbers, each with optional fraction and a unit
	// suffix, such as "300ms", "-1.5h" or "2h45m". Valid time units are "ns",
	// "us" (or "µs"), "ms", "s", "m", "h".
	Timeout string `hcl:"timeout" doc_type:"duration string"`

	// The content of the unit file
	Content string `hcl:"content"`
}

func (p *Preparer) Prepare(render resource.Renderer) (resource.Task, error) {
	name, err := render.Render("name", p.Name)
	if err != nil {
		return nil, err
	}
	ufs, err := render.Render("state", p.UnitFileState)
	if err != nil {
		return nil, err
	}
	sm, err := render.Render("mode", p.StartMode)
	if err != nil {
		return nil, err
	}
	active, err := render.RenderBool("active", p.Active)
	if err != nil {
		active = true
	}
	timeout, err := render.Render("timeout", p.Timeout)
	if err != nil {
		return nil, err
	}
	var t time.Duration
	if timeout == "" {
		t = systemd.DefaultTimeout
	} else {
		t, err = time.ParseDuration(timeout)
		if err != nil {
			return nil, err
		}
	}
	contentStr, err := render.Render("content", p.Content)
	if err != nil {
		return nil, err
	}

	// Handle Defaults
	if ufs == "invalid" {
		return nil, fmt.Errorf("task %q parameter cannot be %q", "UnitFileState", "invalid")
	}
	if sm == "" {
		sm = string(systemd.SMReplace)
	}
	var contentTask *content.Content
	if contentStr != "" {
		contentTask = &content.Content{
			Content: contentStr,
		}
	}

	unit := &Unit{
		Name:          name,
		Active:        active,
		UnitFileState: systemd.UnitFileState(ufs),
		StartMode:     systemd.StartMode(sm),
		Timeout:       t,

		Content: contentTask,
	}
	return unit, unit.Validate()
}

func init() {
	registry.Register("systemd.unit", (*Preparer)(nil), (*Unit)(nil))
}
