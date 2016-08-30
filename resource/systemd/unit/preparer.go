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

	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/systemd"
)

// Preparer for Content
type Preparer struct {
	Name          string `hcl:"name"`
	Active        bool   `hcl:"active"`
	UnitFileState string `hcl:"state"`
	StartMode     string `hcl:"mode"`
	Timeout       string `hcl:"timeout"`
}

// Prepare a new task
// If Mode is the empty string assumes mode should be replace.
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

	// Handle Defaults
	if ufs == "invalid" {
		return nil, fmt.Errorf("task %q parameter cannot be %q", "UnitFileState", "invalid")
	}
	if sm == "" {
		sm = string(systemd.SMReplace)
	}
	unit := &Unit{
		Name:          name,
		Active:        p.Active,
		UnitFileState: systemd.UnitFileState(ufs),
		StartMode:     systemd.StartMode(sm),
		Timeout:       t,
	}
	return unit, unit.Validate()
}
