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
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/asteris-llc/converge/helpers"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/systemd"
	"github.com/coreos/go-systemd/dbus"
)

// Unit represents a systemd units
type Unit struct {
	Name   string // (either just file names or full absolute paths if the unit files are residing outside the usual unit search paths
	Active bool

	UnitFileState systemd.UnitFileState
	Timeout       time.Duration     // how long to wait for the load state to resolve
	StartMode     systemd.StartMode // how to start the unit
}

// Check if all the properties for the unit are correct
// 1. Checks if the unit is currently loading, if so waits Default 5 seconds
// 2. Checks if the unit is in the active state
// 3. Check if the unit is in the unit file state
func (t *Unit) Check() (resource.TaskStatus, error) {
	// Get the connection from the pool
	conn, err := systemd.GetDbusConnection()
	if err != nil {
		return nil, err
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
		return nil, err
	}

	// First check that the ActiveState property matches
	activeState := systemd.ASActive
	if !t.Active {
		activeState = systemd.ASInactive
	}
	validActiveStates := []*dbus.Property{
		systemd.PropActiveState(activeState),
	}
	asStatus, asErr := systemd.CheckProperty(dbusConn, t.Name, "ActiveState", validActiveStates)

	// Then check that there is a valid ufs state
	validUnitFileStates := []*dbus.Property{
		systemd.PropUnitFileState(t.UnitFileState),
	}
	// If the user wants unit file to be linked or linked-runtime, then it may also be in an inabled state
	if t.UnitFileState.IsLinked() {
		if t.UnitFileState.IsRuntimeState() {
			validUnitFileStates = append(validUnitFileStates, systemd.PropUnitFileState(systemd.UFSEnabledRuntime))
		} else {
			validUnitFileStates = append(validUnitFileStates, systemd.PropUnitFileState(systemd.UFSEnabled))
		}
	}
	// if a unit is static, you cannot change the ufs state except to mask it
	if t.UnitFileState != systemd.UFSMasked && t.UnitFileState != systemd.UFSMaskedRuntime {
		validUnitFileStates = append(validUnitFileStates, systemd.PropUnitFileState(systemd.UFSStatic))
	}
	ufsStatus, ufsErr := systemd.CheckProperty(dbusConn, t.Name, "UnitFileState", validUnitFileStates)
	ufsStatus.Merge(asStatus)

	err = helpers.MultiErrorAppend(asErr, ufsErr)
	return ufsStatus, err
}

/* Apply sets the properties
1. Apply active state
2. Apply UFS
*/
func (t *Unit) Apply() error {
	// Get the connection from the pool
	conn, err := systemd.GetDbusConnection()
	if err != nil {
		return err
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
		return err
	}

	// Defer daemon reload
	defer func() {
		prop, err := dbusConn.GetUnitProperty(t.Name, "NeedDaemonReload")
		if err != nil {
			return
		}
		if shouldReload, ok := prop.Value.Value().(bool); shouldReload && ok {
			systemctl, lookErr := exec.LookPath("systemctl")
			if lookErr != nil {
				return
			}
			args := []string{"daemon-reload"}
			env := os.Environ()
			syscall.Exec(systemctl, args, env)
		}
	}()
	// Apply the activeState
	job := make(chan string)
	// TODO check code
	if t.Active {
		_, err = dbusConn.ReloadOrRestartUnit(t.Name, string(t.StartMode), job)
	} else {
		_, err = dbusConn.StopUnit(t.Name, string(t.StartMode), job)
	}
	if err != nil {
		return err
	}
	<-job
	//////////////////////////
	// Apply the UnitFileState
	//////////////////////////

	if t.UnitFileState == systemd.UFSDisabled {
		// Get the current state
		prop, err := dbusConn.GetUnitProperty(t.Name, "UnitFileState")
		if err != nil {
			return err
		}
		state := systemd.UnitFileState(prop.Value.String())
		_, err = dbusConn.DisableUnitFiles([]string{t.Name}, state.IsRuntimeState())
		return err
	}
	// Cover the masked state after disable
	if t.UnitFileState.IsMaskedState() {
		_, err = dbusConn.MaskUnitFiles([]string{t.Name}, t.UnitFileState.IsRuntimeState(), true)
		return err
	}

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
		return err
	}
	if status.StatusCode() == resource.StatusNoChange {
		_, err := dbusConn.UnmaskUnitFiles([]string{t.Name}, t.UnitFileState.IsRuntimeState())
		if err != nil {
			return err
		}
	}

	// If unit file should be linked or not
	if t.UnitFileState.IsLinked() {
		_, err := dbusConn.LinkUnitFiles([]string{t.Name}, t.UnitFileState.IsRuntimeState(), true)
		return err
	}

	// If unit file should be enabled or not
	if t.UnitFileState.IsEnabled() {
		_, _, err := dbusConn.EnableUnitFiles([]string{t.Name}, t.UnitFileState.IsRuntimeState(), true)
		return err
	}
	return nil
}

func (t *Unit) Validate() error {
	if t.Name == "" {
		return fmt.Errorf("task requires a %q parameter", "name")
	}
	if !systemd.IsValidUnitFileState(t.UnitFileState) {
		return fmt.Errorf("invalid %q parameter. can be one of [%s]", systemd.ValidUnitFileStatesWithoutInvalid)
	}
	if !systemd.IsValidStartMode(t.StartMode) {
		return fmt.Errorf("task's parameter %q is not one of %s, is %q", "mode", systemd.ValidStartModes, t.StartMode)
	}
	return nil
}
