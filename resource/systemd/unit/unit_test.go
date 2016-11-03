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

package unit_test

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/user"
	"path/filepath"
	"testing"
	"time"

	"github.com/asteris-llc/converge/helpers/fakerenderer"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/shell"
	"github.com/asteris-llc/converge/resource/systemd"
	"github.com/asteris-llc/converge/resource/systemd/unit"
	"github.com/coreos/go-systemd/dbus"
	"github.com/stretchr/testify/assert"
)

// TestTaskInterface test  that unit taks implements Task
func TestTaskInterface(t *testing.T) {
	t.Parallel()

	assert.Implements(t, (*resource.Task)(nil), new(unit.Unit))
}

// TestCheckingActiveStates checks if unit resource can monitor the active state
// of some units.
func TestCheckingActiveStates(t *testing.T) {
	t.Parallel()

	if !HasSystemd() {
		return
	}
	fr := fakerenderer.FakeRenderer{}
	found, svc := FindUnitWith(
		[]UnitCheck{
			HasActiveStateCheck(systemd.ASActive),
		})
	if !found {
		t.Log("Could not find an active unit on this system")
	} else {

		task := unit.Unit{Name: svc.Name, Active: true}
		status, err := task.Check(&fr)
		assert.NoError(t, err)
		assert.Equal(t, resource.StatusNoChange, status.StatusCode())
		assert.False(t, status.HasChanges())
	}

	found, svc = FindUnitWith(
		[]UnitCheck{
			HasActiveStateCheck(systemd.ASInactive),
		})
	if !found {
		t.Log("Could not find an inactive unit on this system")
	} else {

		task := unit.Unit{Name: svc.Name, Active: false}
		status, err := task.Check(&fr)
		assert.NoError(t, err)
		assert.Equal(t, resource.StatusNoChange, status.StatusCode())
		assert.False(t, status.HasChanges())
	}
}

// TestCheckingUnitFileStates checks if the unit resource can monitor the
// ufs of selected units.
func TestCheckingUnitFileStates(t *testing.T) {
	t.Parallel()

	if !HasSystemd() {
		return
	}
	fr := fakerenderer.FakeRenderer{}
	found, svc := FindUnitWith(
		[]UnitCheck{
			HasActiveStateCheck(systemd.ASActive),
			HasUnitFileStateCheck(systemd.UFSMasked),
		})
	if !found {
		t.Log("Could not find a masked unit on this system")
	} else {

		task := unit.Unit{Name: svc.Name, Active: true, UnitFileState: systemd.UFSMasked}
		status, err := task.Check(&fr)
		assert.NoError(t, err)
		assert.Equal(t, resource.StatusNoChange, status.StatusCode())
		assert.False(t, status.HasChanges())
	}

	found, svc = FindUnitWith(
		[]UnitCheck{
			HasActiveStateCheck(systemd.ASInactive),
			HasUnitFileStateCheck(systemd.UFSEnabled),
		})
	if !found {
		t.Log("Could not find an enabled unit on this system")
	} else {

		task := unit.Unit{Name: svc.Name, Active: false}
		status, err := task.Check(&fr)
		assert.NoError(t, err)
		assert.Equal(t, resource.StatusNoChange, status.StatusCode())
		assert.False(t, status.HasChanges())
	}

	found, svc = FindUnitWith(
		[]UnitCheck{
			HasActiveStateCheck(systemd.ASInactive),
			HasUnitFileStateCheck(systemd.UFSDisabled),
		})
	if !found {
		t.Log("Could not find a disabled unit on this system")
	} else {

		task := unit.Unit{Name: svc.Name, Active: false}
		status, err := task.Check(&fr)
		assert.NoError(t, err)
		assert.Equal(t, resource.StatusNoChange, status.StatusCode())
		assert.False(t, status.HasChanges())
	}
}

// TestInactiveToActiveUnit test  if unit can activate a unit
func TestInactiveToActiveUnit(t *testing.T) {
	t.Parallel()
	fr := fakerenderer.FakeRenderer{}

	if !IsRoot() || !HasSystemd() {
		return
	}
	svc, err := NewTmpService("/run", false)
	assert.NoError(t, err)
	defer svc.Remove()

	foo := unit.Unit{Name: svc.Name, Active: true, UnitFileState: "enabled-runtime", StartMode: "replace"}
	foo.Apply()
	status, err := foo.Check(&fr)
	assert.NoError(t, err)
	assert.Equal(t, resource.StatusNoChange, status.StatusCode())
	assert.Contains(t, status.Messages(), fmt.Sprintf("property \"UnitFileState\" of unit %q is \"enabled-runtime\", expected one of [\"enabled-runtime, static\"]", svc.Name))
	assert.False(t, status.HasChanges())
}

// TestDisabledtoEnabledUnit test  if unit can enable a unit
func TestDisabledtoEnabledUnit(t *testing.T) {
	t.Parallel()
	fr := fakerenderer.FakeRenderer{}

	if !IsRoot() || !HasSystemd() {
		return
	}
	svc, err := NewTmpService("/etc", false)
	assert.NoError(t, err)
	defer svc.Remove()

	disabled := unit.Unit{Name: svc.Name, Active: false, UnitFileState: "disabled", StartMode: "replace"}
	_, err = disabled.Apply()
	assert.NoError(t, err)
	enabled := unit.Unit{Name: svc.Name, Active: true, UnitFileState: "enabled", StartMode: "replace"}
	enabled.Apply()
	status, err := enabled.Check(&fr)
	assert.NoError(t, err)
	assert.Equal(t, resource.StatusNoChange, status.StatusCode())
	assert.Contains(t, status.Messages(), fmt.Sprintf("property \"UnitFileState\" of unit %q is \"enabled\", expected one of [\"enabled, static\"]", svc.Name))
	assert.False(t, status.HasChanges())
}

// TestDisabledtoEnabledRuntimeUnit test  if unit can make a unitfile enabled at runtime
func TestDisabledtoEnabledRuntimeUnit(t *testing.T) {
	t.Parallel()
	fr := fakerenderer.FakeRenderer{}

	if !IsRoot() || !HasSystemd() {
		return
	}
	svc, err := NewTmpService("/run", false)
	assert.NoError(t, err)
	defer svc.Remove()

	disabled := unit.Unit{Name: svc.Name, Active: false, UnitFileState: "disabled", StartMode: "replace"}
	_, err = disabled.Apply()
	assert.NoError(t, err)
	enabled := unit.Unit{Name: svc.Name, Active: true, UnitFileState: systemd.UFSEnabledRuntime, StartMode: "replace"}
	enabled.Apply()
	status, err := enabled.Check(&fr)
	assert.NoError(t, err)
	assert.Equal(t, resource.StatusNoChange, status.StatusCode())
	assert.Contains(t, status.Messages(), fmt.Sprintf("property \"UnitFileState\" of unit %q is \"enabled-runtime\", expected one of [\"enabled-runtime, static\"]", svc.Name))
	assert.False(t, status.HasChanges())
}

// TestEnabledToDisabledUnit test if unit can disable a unit
func TestEnabledToDisabledUnit(t *testing.T) {
	t.Parallel()
	fr := fakerenderer.FakeRenderer{}

	if !IsRoot() || !HasSystemd() {
		return
	}
	svc, err := NewTmpService("/run", false)
	assert.NoError(t, err)
	defer svc.Remove()

	enabled := unit.Unit{Name: svc.Name, Active: true, UnitFileState: "enabled", StartMode: "replace"}
	enabled.Apply()
	disabled := unit.Unit{Name: svc.Name, Active: false, UnitFileState: "disabled", StartMode: "replace"}
	_, err = disabled.Apply()
	assert.NoError(t, err)
	status, err := disabled.Check(&fr)
	assert.NoError(t, err)
	assert.Equal(t, resource.StatusNoChange, status.StatusCode())
	assert.Contains(t, status.Messages(), fmt.Sprintf("property \"UnitFileState\" of unit %q is \"disabled\", expected one of [\"disabled, bad\"]", svc.Name))
	assert.False(t, status.HasChanges())
}

// TestStaticToDisabledUnit tests if unit can disable a static unit
func TestStaticToDisabledUnit(t *testing.T) {
	t.Parallel()
	fr := fakerenderer.FakeRenderer{}

	if !IsRoot() || !HasSystemd() {
		return
	}
	svc, err := NewTmpService("/run", true)
	assert.NoError(t, err)
	defer svc.Remove()

	static := unit.Unit{Name: svc.Name, Active: true, UnitFileState: "static", StartMode: "replace"}
	_, err = static.Apply()
	assert.NoError(t, err)
	disabled := unit.Unit{Name: svc.Name, Active: false, UnitFileState: "disabled", StartMode: "replace"}
	_, err = disabled.Apply()
	assert.NoError(t, err)
	status, err := disabled.Check(&fr)
	assert.NoError(t, err)
	assert.Equal(t, resource.StatusNoChange, status.StatusCode())
	assert.Contains(t, status.Messages(), fmt.Sprintf("property \"UnitFileState\" of unit %q is \"bad\", expected one of [\"disabled, bad\"]", svc.Name))
	assert.False(t, status.HasChanges())
}

// TestLinkedUnit test if unit can link a unit file
func TestLinkedUnit(t *testing.T) {
	t.Parallel()
	fr := fakerenderer.FakeRenderer{}
	if !IsRoot() || !HasSystemd() {
		return
	}
	svc, err := NewTmpService("/tmp", true)
	assert.NoError(t, err)
	defer svc.Remove()

	linked := unit.Unit{Name: svc.Path, Active: true, UnitFileState: systemd.UFSLinked, StartMode: "replace"}
	_, err = linked.Apply()
	assert.NoError(t, err)
	status, err := linked.Check(&fr)
	assert.NoError(t, err)
	assert.Equal(t, resource.StatusNoChange, status.StatusCode())
	assert.Contains(t, status.Messages(), fmt.Sprintf("property \"UnitFileState\" of unit %q is \"linked\", expected one of [\"linked\"]", svc.Name))
	assert.False(t, status.HasChanges())

	// Testing disabling
	disabled := unit.Unit{Name: svc.Name, Active: false, UnitFileState: "disabled", StartMode: "replace"}
	_, err = disabled.Apply()
	assert.NoError(t, err)
	status, err = disabled.Check(&fr)
	assert.NoError(t, err)
	assert.Equal(t, resource.StatusWontChange, status.StatusCode())
	assert.Contains(t, status.Messages(), fmt.Sprintf("unit %q does not exist, considered disabled", svc.Name))
	assert.False(t, status.HasChanges())
}

var units []dbus.UnitStatus

// GetUnits loads the units availible on the system
func GetUnits() error {
	conn, err := systemd.GetDbusConnection()
	if err != nil {
		return err
	}
	defer conn.Return()
	units, err = conn.Connection.ListUnits()
	return err
}

// FilterLoaded should be caled after GetUnits. Removes all unloaded units.
func FilterLoaded() {
	newUnits := []dbus.UnitStatus{}
	for i := range units {
		if systemd.LoadState(units[i].LoadState).Equal(systemd.LSLoaded) {
			newUnits = append(newUnits, units[i])
		}
	}
	units = newUnits
}

// UnitCheck function that checks the UnitStatus for a property.
type UnitCheck func(dbus.UnitStatus) bool

// HasActiveStateCheck returns a UnitCheck that determines if the Unit has the
// expected active state
func HasActiveStateCheck(state systemd.ActiveState) UnitCheck {
	return func(svc dbus.UnitStatus) bool {
		return systemd.ActiveState(svc.ActiveState).Equal(state)
	}
}

// HasUnitFileStateCheck returns a UnitCheck that determines if the Unit has the
// expected unit file state
func HasUnitFileStateCheck(state systemd.UnitFileState) UnitCheck {
	return func(svc dbus.UnitStatus) bool {
		return systemd.UnitFileState(svc.SubState).Equal(state)
	}
}

// FindUnitWith returns the first unit that passes all the UnitChecks
func FindUnitWith(checks []UnitCheck) (bool, dbus.UnitStatus) {
	for _, unit := range units {
		passedChecks := true
		for _, check := range checks {
			if !check(unit) {
				passedChecks = false
				break
			}
		}
		if passedChecks {
			return true, unit
		}
	}
	return false, dbus.UnitStatus{}
}

// HelloUnit is a simple service file that is not static
const HelloUnit = `
[Unit]
Description=Foo hello world
[Service]
ExecStart=/bin/bash -c "while true; do /bin/echo HELLO WORLD; sleep 5; done;"

[Install]
WantedBy=multi-user.target
`

// HelloUnitStatic is a simple service file that is  static
const HelloUnitStatic = `
[Unit]
Description=Foo hello world
[Service]
ExecStart=/bin/bash -c "while true; do /bin/echo HELLO WORLD; sleep 5; done;"
`

// TmpService is the information for a temporary service file
type TmpService struct {
	Path string
	Name string
}

// Remove removes this service file from the system entirely
func (t *TmpService) Remove() {
	base := filepath.Base(t.Path)

	kill := "systemctl kill " + base
	disable := "systemctl disable " + base
	daemonReload := "systemctl daemon-reload"
	resetFailed := "systemctl reset-failed"

	generator := &shell.CommandGenerator{Interpreter: "/bin/bash"}
	generator.Run(disable)
	generator.Run(kill)

	os.Remove(t.Path)
	locations := []string{"/run/systemd/system/", "/etc/systemd/system/", "/usr/lib/systemd/system/"}
	for _, l := range locations {
		os.Remove(filepath.Join(l, base))
	}

	generator.Run(resetFailed)
	generator.Run(daemonReload)
}

var randGen *rand.Rand

// NewTmpService creates a service file on the system either in
// "/tmp", "/etc/systemd/system", or "/run/systemd/system".
func NewTmpService(prefix string, static bool) (svc *TmpService, err error) {
	number := randGen.Uint32()
	name := fmt.Sprintf("foo%d.service", number)
	var path string
	if prefix == "/tmp" {
		path = filepath.Join(prefix, name)
	} else {
		path = filepath.Join(prefix, "systemd/system", name)
	}

	if static {
		err = ioutil.WriteFile(path, []byte(HelloUnitStatic), 0777)
	} else {
		err = ioutil.WriteFile(path, []byte(HelloUnit), 0777)
	}
	if err != nil {
		return nil, err
	}
	return &TmpService{Path: path, Name: name}, systemd.ApplyDaemonReload()
}

// IsRoot checks if the current user is Root
func IsRoot() bool {
	currentUser, _ := user.Current()
	return currentUser.Uid == "0"
}

// HasSystemd checks if dbus is available
func HasSystemd() bool {
	conn, err := systemd.GetDbusConnection()
	if err != nil {
		return false
	}
	defer conn.Return()
	return true
}

func init() {
	randGen = rand.New(rand.NewSource(time.Now().UnixNano()))
	if HasSystemd() {
		GetUnits()
		FilterLoaded()
	}
}
