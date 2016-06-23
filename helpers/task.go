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

package helpers

import (
	"fmt"

	"github.com/asteris-llc/converge/resource"
)

// DummyTask is a task that can be configured to always/never change (when
// planning), always/never fail (when applying), and always/never throw an error
// (when planning or applying). It can also act as a monitor.
type DummyTask struct {
	Name string
	Deps []string
	// configuration options
	Monitor          bool
	Change           bool
	ChangeAfterApply bool
	PrepareError     error
	CheckError       error
	ApplyError       error
}

// Check satisfies the Monitor interface
func (d *DummyTask) Check() (status string, willChange bool, err error) {
	return fmt.Sprintf("will change: %v", d.Change), d.Change, d.CheckError
}

// Apply satisfies the Task interface
func (d *DummyTask) Apply() error {
	if !d.ChangeAfterApply {
		d.Change = false
	}
	return d.ApplyError
}

// String satisfies the Resource, fmt.Stringer interfaces
func (d *DummyTask) String() string {
	if d.Monitor {
		return "monitor." + d.Name
	}
	return "task." + d.Name
}

// Prepare satisfies the Resource interface
func (d *DummyTask) Prepare(m *resource.Module) error { return d.PrepareError }

// Depends satisfies the Resource interface
func (d *DummyTask) Depends() []string { return d.Deps }

// SetDepends satisfies the Resource interface
func (d *DummyTask) SetDepends(deps []string) { d.Deps = deps }

// HasBaseDependencies satisfies the Resource interface
func (d *DummyTask) HasBaseDependencies() bool { return false }

// SetName satisfies the Resource interface
func (d *DummyTask) SetName(name string) { d.Name = name }
