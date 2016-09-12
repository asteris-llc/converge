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
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/file/content"
	"github.com/asteris-llc/converge/resource/systemd"
	"github.com/coreos/go-systemd/dbus"
)

// Unit represents a systemd units
// TODO Enable Parallelization of this Unit
type Unit struct {
	// The name of this unit as in "foo.service"
	Name string // (either just file names or full absolute paths if the unit files are residing outside the usual unit search paths

	// Active determines whether this unit should be active or inactive
	Active bool

	/*
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
	UnitFileState systemd.UnitFileState

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
	StartMode systemd.StartMode // how to start the unit

	// the amount of time the command will wait for configuration to load
	// before halting forcefully. The
	// format is Go's duraction string. A duration string is a possibly signed
	// sequence of decimal numbers, each with optional fraction and a unit
	// suffix, such as "300ms", "-1.5h" or "2h45m". Valid time units are "ns",
	// "us" (or "µs"), "ms", "s", "m", "h".
	Timeout time.Duration // how long to wait for the load state to resolve

	// Composed Tasks
	Content *content.Content
}

// Check if all the properties for the unit are correct
// 1. Checks if the unit is currently loading, if so waits Default 5 seconds
// 2. Checks if the unit is in the active state
// 3. Check if the unit is in the unit file state
func (t *Unit) Check(r resource.Renderer) (resource.TaskStatus, error) {
	systemd.ApplyDaemonReload()
	// Get the connection from the pool
	conn, err := systemd.GetDbusConnection()
	if err != nil {
		status := &resource.Status{
			WarningLevel: resource.StatusFatal,
			Status:       err.Error(),
		}
		return status, err
	}
	defer conn.Return()
	dbusConn := conn.Connection

	//Create context
	var (
		ctx    context.Context
		cancel context.CancelFunc
	)
	if t.Timeout == 0 {
		t.Timeout = systemd.DefaultTimeout
	}
	ctx, cancel = context.WithTimeout(context.Background(), t.Timeout)
	defer cancel()

	err = systemd.WaitToLoad(ctx, dbusConn, t.Name)
	if err != nil {
		status := &resource.Status{
			WarningLevel: resource.StatusFatal,
			Status:       err.Error(),
		}
		return status, err
	}

	// First check if the content matches
	var contentStatus *resource.Status
	var contentError error
	if t.Content != nil {
		err := t.loadContent(dbusConn)
		if err != nil {
			status := &resource.Status{
				WarningLevel: resource.StatusFatal,
				Status:       err.Error(),
			}
			return status, err
		}
		// Don't return immediately if there was a content error.
		taskStatus, err := t.Content.Check(r)
		contentError = err
		contentStatus = taskStatus.(*content.Content).Status
	}

	// Next check that the ActiveState property matches
	activeState := systemd.ASActive
	if !t.Active {
		activeState = systemd.ASInactive
	}
	validActiveStates := []*dbus.Property{
		systemd.PropActiveState(activeState),
	}
	asStatus, asErr := systemd.CheckProperty(dbusConn, t.Name, "ActiveState", validActiveStates)
	asStatus = systemd.AppendStatus(asStatus, contentStatus)
	asErr = systemd.MultiErrorAppend(contentError, asErr)

	// Then check that there is a valid ufs state
	validUnitFileStates := []*dbus.Property{
		systemd.PropUnitFileState(t.UnitFileState),
	}
	// If the user wants unit file to be linked or linked-runtime, then it may also be in an enabled state
	if t.UnitFileState.IsLinked() {
		if t.UnitFileState.IsRuntimeState() {
			validUnitFileStates = append(validUnitFileStates, systemd.PropUnitFileState(systemd.UFSEnabledRuntime))
		} else {
			validUnitFileStates = append(validUnitFileStates, systemd.PropUnitFileState(systemd.UFSEnabled))
		}
	}
	// if a unit is static, you CANNOT change the ufs state except to mask it
	if t.UnitFileState != systemd.UFSStatic && t.UnitFileState != systemd.UFSMasked && t.UnitFileState != systemd.UFSMaskedRuntime {
		validUnitFileStates = append(validUnitFileStates, systemd.PropUnitFileState(systemd.UFSStatic))
	}

	// a masked unit may have the "bad" ufs
	if t.UnitFileState.IsMaskedState() {
		validUnitFileStates = append(validUnitFileStates, systemd.PropUnitFileState(systemd.UFSBad))
	}
	ufsStatus, ufsErr := systemd.CheckProperty(dbusConn, t.Name, "UnitFileState", validUnitFileStates)
	ufsStatus = systemd.AppendStatus(ufsStatus, asStatus)

	err = systemd.MultiErrorAppend(asErr, ufsErr)
	return ufsStatus, err
}

/* Apply sets the properties
1. Apply active state
2. Apply UFS
TODO linking and masking units
*/
func (t *Unit) Apply(r resource.Renderer) (resource.TaskStatus, error) {
	// Get the connection from the pool
	conn, err := systemd.GetDbusConnection()
	if err != nil {
		status := &resource.Status{
			WarningLevel: resource.StatusFatal,
			Status:       err.Error(),
		}
		return status, err
	}
	defer conn.Return()
	dbusConn := conn.Connection

	//Create context
	var (
		ctx    context.Context
		cancel context.CancelFunc
	)
	if t.Timeout == 0 {
		t.Timeout = systemd.DefaultTimeout
	}
	ctx, cancel = context.WithTimeout(context.Background(), t.Timeout)
	defer cancel()

	err = systemd.WaitToLoad(ctx, dbusConn, t.Name)
	if err != nil {
		status := &resource.Status{
			WarningLevel: resource.StatusFatal,
			Status:       err.Error(),
		}
		return status, err
	}

	// First Apply Content
	var contentStatus *resource.Status
	if t.Content != nil {
		err := t.loadContent(dbusConn)
		if err != nil {
			status := &resource.Status{
				WarningLevel: resource.StatusFatal,
				Status:       err.Error(),
			}
			return status, err
		}
		taskStatus, err := t.Content.Apply(r)
		contentStatus = taskStatus.(*content.Content).Status

		if err != nil {
			return contentStatus, err
		}
		// reload daemon as file changed on disk
		err = systemd.ApplyDaemonReload()
		if err != nil {
			status := &resource.Status{
				WarningLevel: resource.StatusFatal,
				Status:       err.Error(),
			}
			return systemd.AppendStatus(status, contentStatus), err
		}
	}

	// Determine if unit was enabled-runtime or not
	prop, err := dbusConn.GetUnitProperty(t.Name, "UnitFileState")
	if err != nil {
		status := &resource.Status{
			WarningLevel: resource.StatusFatal,
			Status:       err.Error(),
		}
		return systemd.AppendStatus(status, contentStatus), err
	}
	state := systemd.UnitFileState(prop.Value.String())

	// Apply the activeState
	job := make(chan string)
	if t.Active {
		_, err = dbusConn.ReloadOrRestartUnit(t.Name, string(t.StartMode), job)
	} else {
		_, err = dbusConn.StopUnit(t.Name, string(t.StartMode), job)
	}
	if err != nil {
		status := &resource.Status{
			WarningLevel: resource.StatusFatal,
			Status:       err.Error(),
		}
		return systemd.AppendStatus(status, contentStatus), err
	}
	<-job

	//////////////////////////
	// Apply the UnitFileState
	//////////////////////////

	// Before manking further changes, unmask unit file
	status, err := systemd.CheckProperty(
		dbusConn,
		t.Name,
		"UnitFileState",
		[]*dbus.Property{
			systemd.PropUnitFileState(systemd.UFSMasked),
			systemd.PropUnitFileState(systemd.UFSMaskedRuntime),
		},
	)
	if err != nil {
		status := &resource.Status{
			WarningLevel: resource.StatusFatal,
			Status:       err.Error(),
		}
		return systemd.AppendStatus(status, contentStatus), err
	}
	// If unit is maksed, and user does not want it masked, unmask it.
	if status.StatusCode() == resource.StatusNoChange && !(t.UnitFileState.Equal(systemd.UFSMasked) || t.UnitFileState.Equal(systemd.UFSMaskedRuntime)) {
		_, err := dbusConn.UnmaskUnitFiles([]string{t.Name}, state.IsRuntimeState())
		if err != nil {
			status := &resource.Status{
				WarningLevel: resource.StatusFatal,
				Status:       err.Error(),
			}
			return systemd.AppendStatus(status, contentStatus), err
		}
	}

	// Now that the unit is unmasked continue.

	////////////////////////////
	//Disabling
	////////////////////////////
	if t.UnitFileState.Equal(systemd.UFSDisabled) {
		_, err = dbusConn.DisableUnitFiles([]string{t.Name}, state.IsRuntimeState())
		if err != nil {
			status := &resource.Status{
				WarningLevel: resource.StatusFatal,
				Status:       err.Error(),
			}
			return status, err
		}
		return t.Check(r)
	}
	// If unit file should be enabled or not
	if t.UnitFileState.IsEnabled() {
		_, _, err := dbusConn.EnableUnitFiles([]string{t.Name}, t.UnitFileState.IsRuntimeState(), true)
		if err != nil {
			status := &resource.Status{
				WarningLevel: resource.StatusFatal,
				Status:       err.Error(),
			}
			return systemd.AppendStatus(status, contentStatus), err
		}
	}
	return t.Check(r)
}

func (t *Unit) Validate() error {
	if t.Name == "" {
		return fmt.Errorf("task requires a %q parameter", "name")
	}
	if !systemd.IsValidUnitFileState(t.UnitFileState) {
		return fmt.Errorf("invalid %q parameter. can be one of [%s]", "state", systemd.ValidUnitFileStatesWithoutInvalid)
	}
	if !systemd.IsValidStartMode(t.StartMode) {
		return fmt.Errorf("task's parameter %q is not one of %s, is %q", "mode", systemd.ValidStartModes, t.StartMode)
	}
	return nil
}

//loadContent fills in the Destination parameter of the Content parameter.
func (t *Unit) loadContent(dbusConn *dbus.Conn) error {
	if t.Content.Destination != "" {
		return nil
	}
	prop, err := dbusConn.GetUnitProperty(t.Name, "FragmentPath")
	if err != nil {
		return err
	}
	str := strings.Replace(prop.Value.String(), "\"", "", -1)
	if str == "" {
		return fmt.Errorf("Could not find unit %q", t.Name)
	}
	t.Content.Destination = str
	return nil
}
