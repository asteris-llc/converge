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
	"path/filepath"
	"time"

	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/systemd"
	"github.com/coreos/go-systemd/dbus"
	"github.com/hashicorp/go-multierror"
)

// Unit represents a systemd units
// NOTE: This unit does not parallelize with itself
type Unit struct {
	Name          string
	Active        bool
	UnitFileState systemd.UnitFileState
	StartMode     systemd.StartMode
	Timeout       time.Duration

	*resource.Status
}

var loadError = `LoadError: "[\"org.freedesktop.DBus.Error.FileNotFound\", \"No such file or directory\"]"`

// Check if all the properties for the unit are correct
// 1. Checks if the unit is currently loading, if so waits Default 5 seconds
// 2. Checks if the unit is in the active state
// 3. Check if the unit is in the unit file state
func (t *Unit) Check(r resource.Renderer) (resource.TaskStatus, error) {
	systemd.ApplyDaemonReload()
	/*First thing to do is to check if the Name given is outside
	the normal search path of systemd. If so it should be linked
	*/
	_, unitName := filepath.Split(t.Name)

	// Get the connection from the pool
	conn, err := systemd.GetDbusConnection()
	if err != nil {
		t.Status = &resource.Status{
			Level:  resource.StatusFatal,
			Output: []string{err.Error()},
		}
		return t, err
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

	// Waits for the daemon to finish loading.
	// After check for unit linking
	err = systemd.WaitToLoad(ctx, dbusConn, unitName)

	var linkStatus *resource.Status
	if err != nil {
		linkStatus, err = t.CheckLinkage(err)
		t.Status = linkStatus
		return linkStatus, err
	}

	// Next check that the ActiveState property matches
	activeState := systemd.ASActive
	if !t.Active {
		activeState = systemd.ASInactive
	}
	validActiveStates := []*dbus.Property{
		systemd.PropActiveState(activeState),
	}
	asStatus, asErr := systemd.CheckProperty(dbusConn, unitName, "ActiveState", validActiveStates)
	asStatus = systemd.AppendStatus(asStatus, linkStatus)

	if t.UnitFileState == "" {
		t.Status = asStatus
		return asStatus, asErr
	}

	// Then check that there is a valid ufs state
	validUnitFileStates := []*dbus.Property{
		systemd.PropUnitFileState(t.UnitFileState),
	}

	// if a unit is static, consider it enabled
	if t.UnitFileState.IsEnabled() {
		validUnitFileStates = append(validUnitFileStates, systemd.PropUnitFileState(systemd.UFSStatic))
	}

	// a masked unit may have the "bad" ufs
	if t.UnitFileState.Equal(systemd.UFSDisabled) {
		validUnitFileStates = append(validUnitFileStates, systemd.PropUnitFileState(systemd.UFSBad))
	}
	ufsStatus, ufsErr := systemd.CheckProperty(dbusConn, unitName, "UnitFileState", validUnitFileStates)
	ufsStatus = systemd.AppendStatus(ufsStatus, asStatus)

	if ufsErr != nil {
		if asErr != nil {
			err = multierror.Append(asErr, asErr)
		}
		err = ufsErr
	}

	t.Status = ufsStatus
	return t, err
}

// CheckLinkage handles the "LoadError" systemd returns when a unit file exist
// outside the normal search path.
func (t *Unit) CheckLinkage(err error) (*resource.Status, error) {
	dir, unitName := filepath.Split(t.Name)
	shouldBeLinked := false
	if dir != "" && t.UnitFileState.IsLinked() {
		// The fullpath to a unit was given. Assume that this isn't in the
		// normal search paths
		shouldBeLinked = true
	}

	// Unit doesn't exist should be linked
	if err.Error() == loadError {
		if shouldBeLinked {
			// Check if file to be linked exist on disk
			_, statErr := os.Stat(t.Name)
			if statErr != nil {
				status := &resource.Status{
					Level:  resource.StatusFatal,
					Output: []string{statErr.Error()},
				}
				return status, statErr
			}

			statusMsg := fmt.Sprintf("unit %q does not exist, will be linked", unitName)
			diffMsg := fmt.Sprintf("unit %q does not exist", unitName)
			diff := resource.TextDiff{Default: diffMsg, Values: [2]string{diffMsg, fmt.Sprintf("unit %q is linked", unitName)}}
			status := &resource.Status{
				Level:       resource.StatusWillChange,
				Differences: map[string]resource.Diff{unitName: diff},
				Output:      []string{statusMsg},
			}
			return status, nil
		} else if t.UnitFileState.Equal(systemd.UFSDisabled) {
			// Consider it the same as it being disabled.
			statusMsg := fmt.Sprintf("unit %q does not exist, considered disabled", unitName)
			status := &resource.Status{
				Level:  resource.StatusWontChange,
				Output: []string{statusMsg},
			}
			return status, nil

		}
	}
	status := &resource.Status{
		Level:  resource.StatusFatal,
		Output: []string{err.Error()},
	}
	return status, err

}

// Apply sets the properties
// Apply active state
// Apply UFS
func (t *Unit) Apply() (resource.TaskStatus, error) {
	//First thing to do is to check if the Name given is outside
	//the normal search path of systemd. If so it should be linked
	_, unitName := filepath.Split(t.Name)

	// Get the connection from the pool
	conn, err := systemd.GetDbusConnection()
	if err != nil {
		t.Status = &resource.Status{
			Level:  resource.StatusFatal,
			Output: []string{err.Error()},
		}
		return t, err
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

	err = systemd.WaitToLoad(ctx, dbusConn, unitName)
	var linkStatus *resource.Status
	if err != nil {
		linkStatus, err = t.ApplyLinkage(err, dbusConn)
		if err != nil {
			return linkStatus, err
		}
		systemd.ApplyDaemonReload()
	}
	// Start the unit
	startStatus, err := t.startOrStop(dbusConn)
	startedLinkedStatus := systemd.AppendStatus(startStatus, linkStatus)
	if err != nil {
		t.Status = startedLinkedStatus
		return t.Status, err
	}
	systemd.ApplyDaemonReload()

	if t.UnitFileState == "" {
		t.Status = startedLinkedStatus
		return t.Status, err
	}
	// Get the current UnitFileState
	prop, err := dbusConn.GetUnitProperty(unitName, "UnitFileState")
	if err != nil {
		t.Status = systemd.AppendStatus(&resource.Status{
			Level:  resource.StatusFatal,
			Output: []string{err.Error()},
		}, startedLinkedStatus)
		return t, err
	}
	state := systemd.UnitFileState(prop.Value.Value().(string))

	// Apply the UnitFileState

	// First thing first, if the state you want is the state you have, quit.
	if t.UnitFileState.Equal(state) {
		statusMsg := fmt.Sprintf("unit %q is %s", unitName, state)
		t.Status = systemd.AppendStatus(&resource.Status{
			Level:  resource.StatusNoChange,
			Output: []string{statusMsg},
		}, startedLinkedStatus)
		return t, nil
	}

	switch t.UnitFileState {
	//ENABLING
	case systemd.UFSEnabled:
		fallthrough
	case systemd.UFSEnabledRuntime:
		// treate enabled and enabled runtime as the same thing
		enabledStatus, err := t.ApplyEnabledState(state, dbusConn)
		t.Status = systemd.AppendStatus(enabledStatus, startedLinkedStatus)
		return t, err

	//Disabling
	case systemd.UFSDisabled:
		disabledStatus, err := t.ApplyDisabledState(state, dbusConn)
		t.Status = systemd.AppendStatus(disabledStatus, startedLinkedStatus)
		return t, err
	//Linking
	case systemd.UFSLinked:
		fallthrough
	case systemd.UFSLinkedRuntime:
		// must change from linked to linked-runtime or vice versa
		if state.IsLinked() {
			_, err := dbusConn.DisableUnitFiles([]string{unitName}, state.IsRuntimeState())
			if err != nil {
				t.Status = systemd.AppendStatus(&resource.Status{
					Level:  resource.StatusFatal,
					Output: []string{err.Error()},
				}, startedLinkedStatus)
				return t, err
			}
			// Must relink file
			reLinkedStatus, err := t.ApplyLinkage(fmt.Errorf(loadError), dbusConn)
			t.Status = systemd.AppendStatus(reLinkedStatus, startedLinkedStatus)
			return t.Status, err
		}
	case systemd.UFSStatic:
		statusMsg := "cannot force unit to be static"
		t.Status = systemd.AppendStatus(&resource.Status{
			Level:  resource.StatusCantChange,
			Output: []string{statusMsg},
		}, startedLinkedStatus)
		return t, nil
	}

	statusMsg := fmt.Errorf("unit %q is in unknown state, %s", unitName, t.UnitFileState)
	t.Status = systemd.AppendStatus(&resource.Status{
		Level:  resource.StatusFatal,
		Output: []string{statusMsg.Error()},
	}, startedLinkedStatus)
	return t, statusMsg
}

// ApplyLinkage links a file, if an error in dbus is a loadError due to a unit file
// residing outside the normal search paths.
func (t *Unit) ApplyLinkage(err error, dbusConn *dbus.Conn) (*resource.Status, error) {
	//First thing to do is to check if the Name given is outside
	//the normal search path of systemd. If so it should be linked
	dir, unitName := filepath.Split(t.Name)
	var shouldBeLinked = false
	if dir != "" && t.UnitFileState.IsLinked() {
		// The fullpath to a unit was given. Assume that this isn't in the
		// normal search paths
		shouldBeLinked = true
	}

	// Unit doesn't exist should be linked
	if err.Error() == loadError {
		if shouldBeLinked {
			// Check if file to be linked exist on disk
			_, statErr := os.Stat(t.Name)
			if statErr != nil {
				status := &resource.Status{
					Level:  resource.StatusFatal,
					Output: []string{err.Error()},
				}
				return status, statErr
			}
			// Link the Unit
			changes, e := dbusConn.LinkUnitFiles([]string{t.Name}, t.UnitFileState.IsRuntimeState(), true)
			if e != nil {
				status := &resource.Status{
					Level:  resource.StatusFatal,
					Output: []string{err.Error()},
				}
				return status, err
			}
			statusMsg := fmt.Sprintf("unit %q has been linked", unitName)
			diffMsg := fmt.Sprintf("unit %q does not exist", unitName)
			diff := resource.TextDiff{Default: diffMsg, Values: [2]string{diffMsg, fmt.Sprintf("unit %q is linked", unitName)}}
			status := &resource.Status{
				Level:       resource.StatusNoChange,
				Differences: map[string]resource.Diff{unitName: diff},
				Output: []string{
					statusMsg,
					fmt.Sprintf("file %q has been linkded to %q", changes[0].Filename, changes[0].Destination),
				},
			}
			return status, nil
		}
	}
	status := &resource.Status{
		Level:  resource.StatusFatal,
		Output: []string{err.Error()},
	}
	return status, err

}

// ApplyEnabledState Attempts to enable the file. Cannot switch between an enabled
// and an enabled-runtime state.
func (t *Unit) ApplyEnabledState(currentState systemd.UnitFileState, dbusConn *dbus.Conn) (*resource.Status, error) {
	_, unitName := filepath.Split(t.Name)

	// treate enabled and enabled runtime as the same thing
	if currentState.IsEnabled() {
		statusMsg := fmt.Sprintf("unit %q is enabled, cannot switch between enabled-runtime and enabled", unitName)
		status := &resource.Status{
			Level:  resource.StatusWontChange,
			Output: []string{statusMsg},
		}
		return status, nil
	}
	// check if unit is in a state that cannot be enabled
	if currentState.IsLinked() || currentState.Equal(systemd.UFSInvalid) || currentState.Equal(systemd.UFSBad) {
		statusMsg := fmt.Errorf("unit %q cannot be enabled. state is %s", unitName, currentState)
		status := &resource.Status{
			Level: resource.StatusCantChange,
			Output: []string{
				statusMsg.Error(),
			},
		}
		return status, statusMsg
	}
	// At this point state must be disabled, static, or masked
	if currentState.Equal(systemd.UFSStatic) { // If static do nothing
		statusMsg := fmt.Sprintf("uint %q does not need to be enabled. state is static", unitName)
		status := &resource.Status{
			Level: resource.StatusNoChange,
			Output: []string{
				statusMsg,
			},
		}
		return status, nil
	}
	if currentState.IsMaskedState() { // If masked unmask
		changes, err := dbusConn.UnmaskUnitFiles([]string{unitName}, currentState.IsRuntimeState())
		if err != nil {
			status := &resource.Status{
				Level:  resource.StatusFatal,
				Output: []string{err.Error()},
			}
			return status, err
		}
		statusMsg := fmt.Sprintf("unit %q is unmasked", unitName)
		status := &resource.Status{
			Level: resource.StatusNoChange,
			Output: []string{
				statusMsg,
				fmt.Sprintf("symlink for %q was %q. link between %q and %q", unitName, changes[0].Type, changes[0].Filename, changes[0].Destination),
			},
		}
		return status, nil
	}
	// if hasEnablementInfo is false then the unit will become static
	hasEnablementInfo, changes, err := dbusConn.EnableUnitFiles([]string{unitName}, t.UnitFileState.IsRuntimeState(), true)
	if err != nil {
		status := &resource.Status{
			Level:  resource.StatusFatal,
			Output: []string{err.Error()},
		}
		return status, err
	}
	ufs := t.UnitFileState
	if hasEnablementInfo {
		ufs = "static"
	}
	statusMsg := fmt.Sprintf("unit %q is %s", unitName, ufs)
	status := &resource.Status{
		Level: resource.StatusNoChange,
		Output: []string{
			statusMsg,
			fmt.Sprintf("symlink for %q was %q. link between %q and %q", unitName, changes[0].Type, changes[0].Filename, changes[0].Destination),
		},
	}
	return status, nil

}

// ApplyDisabledState attempts to disable a unit file
func (t *Unit) ApplyDisabledState(currentState systemd.UnitFileState, dbusConn *dbus.Conn) (*resource.Status, error) {
	_, unitName := filepath.Split(t.Name)

	if currentState.IsMaskedState() {
		statusMsg := fmt.Sprintf("unit %q is maked", unitName)
		status := &resource.Status{
			Level:  resource.StatusNoChange,
			Output: []string{statusMsg},
		}
		return status, nil
	}
	if currentState.Equal(systemd.UFSStatic) {
		changes, err := dbusConn.MaskUnitFiles([]string{unitName}, false, true)
		if err != nil {
			status := &resource.Status{
				Level:  resource.StatusFatal,
				Output: []string{err.Error()},
			}
			return status, err
		}
		statusMsg := fmt.Sprintf("unit %q is maksed", unitName)
		status := &resource.Status{
			Level: resource.StatusNoChange,
			Output: []string{
				statusMsg,
				fmt.Sprintf("symlink for %q was %q. link between %q and %q", unitName, changes[0].Type, changes[0].Filename, changes[0].Destination),
			},
		}
		return status, err
	}

	_, err := dbusConn.DisableUnitFiles([]string{unitName}, currentState.IsRuntimeState())
	if err != nil {
		status := &resource.Status{
			Level:  resource.StatusFatal,
			Output: []string{err.Error()},
		}
		return status, err
	}
	statusMsg := fmt.Sprintf("unit %q has been disabled", unitName)
	status := &resource.Status{
		Level: resource.StatusNoChange,
		Output: []string{
			statusMsg,
		},
	}
	return status, nil

}

func (t *Unit) resetFailed(dbusConn *dbus.Conn) (*resource.Status, error) {
	var err error
	_, unitName := filepath.Split(t.Name)

	// First check if the unit is failed so that we can call reset-failed
	// Get the current UnitFileState
	shouldReset, err := systemd.CheckResetFailed(unitName)
	if err != nil {
		t.Status = &resource.Status{
			Level:  resource.StatusFatal,
			Output: []string{err.Error()},
		}
		return t.Status, err
	}
	if shouldReset {
		err = systemd.ApplyResetFailed(unitName)
		if err != nil {
			t.Status = &resource.Status{
				Level:  resource.StatusFatal,
				Output: []string{err.Error()},
			}
			return t.Status, err
		}
		t.Status = &resource.Status{
			Level:  resource.StatusNoChange,
			Output: []string{fmt.Sprintf("unit %q was failed, reset", unitName)},
		}
		return t.Status, err
	}
	t.Status = &resource.Status{
		Level:  resource.StatusNoChange,
		Output: []string{fmt.Sprintf("unit %q is not failed, no need to reset", unitName)},
	}
	return t.Status, nil

}

func (t *Unit) restartUnit(dbusConn *dbus.Conn) (*resource.Status, error) {
	var err error
	_, unitName := filepath.Split(t.Name)

	status, err := t.resetFailed(dbusConn)
	if err != nil {
		t.Status = &resource.Status{
			Level:  resource.StatusFatal,
			Output: []string{fmt.Sprintf("cannot restart unit %q", unitName)},
		}
		t.Status = systemd.AppendStatus(t.Status, status)
		return t.Status, err
	}

	// Apply the activeState
	job := make(chan string)
	_, err = dbusConn.ReloadOrRestartUnit(unitName, string(t.StartMode), job)
	if err != nil {
		t.Status = &resource.Status{
			Level:  resource.StatusFatal,
			Output: []string{err.Error()},
		}
		return t.Status, err
	}
	<-job
	statusMsg := fmt.Sprintf("unit %q is %s", unitName, "active")
	t.Status = &resource.Status{
		Level:  resource.StatusNoChange,
		Output: []string{statusMsg},
	}
	return t.Status, nil
}

func (t *Unit) stopUnit(dbusConn *dbus.Conn) (*resource.Status, error) {
	var err error
	_, unitName := filepath.Split(t.Name)

	status, err := t.resetFailed(dbusConn)
	if err != nil {
		t.Status = &resource.Status{
			Level:  resource.StatusFatal,
			Output: []string{fmt.Sprintf("cannot restart unit %q", unitName)},
		}
		t.Status = systemd.AppendStatus(t.Status, status)
		return t.Status, err
	}

	// Apply the activeState
	job := make(chan string)
	_, err = dbusConn.StopUnit(unitName, string(t.StartMode), job)
	if err != nil {
		t.Status = &resource.Status{
			Level:  resource.StatusFatal,
			Output: []string{err.Error()},
		}
		return t.Status, err
	}
	<-job
	statusMsg := fmt.Sprintf("unit %q is %s", unitName, "inactive")
	t.Status = &resource.Status{
		Level:  resource.StatusNoChange,
		Output: []string{statusMsg},
	}
	return t.Status, nil
}

func (t *Unit) startOrStop(dbusConn *dbus.Conn) (*resource.Status, error) {
	if t.Active {
		return t.restartUnit(dbusConn)
	}
	return t.stopUnit(dbusConn)

}
