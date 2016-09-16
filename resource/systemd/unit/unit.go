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
}

// Check if all the properties for the unit are correct
// 1. Checks if the unit is currently loading, if so waits Default 5 seconds
// 2. Checks if the unit is in the active state
// 3. Check if the unit is in the unit file state
func (t *Unit) Check(r resource.Renderer) (resource.TaskStatus, error) {
	systemd.ApplyDaemonReload()
	/*////////////////////////////////////////////////////
	First thing to do is to check if the Name given is outside
	the normal search path of systemd. If so it should be linked
	*/ ////////////////////////////////////////////////////
	dir, unitName := filepath.Split(t.Name)
	var shouldBeLinked = false
	if dir != "" && t.UnitFileState.IsLinked() {
		// The fullpath to a unit was given. Assume that this isn't in the
		// normal search paths
		shouldBeLinked = true
	}

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

	// Waits for the daemon to finish loading.
	// After check for unit linking
	err = systemd.WaitToLoad(ctx, dbusConn, unitName)

	if err != nil {
		// Unit doesn't exist should be linked
		if err.Error() == "LoadError: \"[\\\"org.freedesktop.DBus.Error.FileNotFound\\\", \\\"No such file or directory\\\"]\"" {
			if shouldBeLinked {
				// Check if file to be linked exist on disk
				_, statErr := os.Stat(t.Name)
				if statErr != nil {
					return &resource.Status{
						Status:       err.Error(),
						WarningLevel: resource.StatusFatal,
						Output:       []string{err.Error()},
					}, err
				}
				statusMsg := fmt.Sprintf("unit %q does not exist, will be linked", unitName)
				diffMsg := fmt.Sprintf("unit %q does not exist", unitName)
				diff := resource.TextDiff{Default: diffMsg, Values: [2]string{diffMsg, fmt.Sprintf("unit %q is linked", unitName)}}
				return &resource.Status{
					Status:       statusMsg,
					WarningLevel: resource.StatusWillChange,
					WillChange:   true,
					Differences:  map[string]resource.Diff{unitName: diff},
					Output:       []string{statusMsg, err.Error()},
				}, nil
			} else if t.UnitFileState.Equal(systemd.UFSDisabled) {
				// Consider it the same as it being disabled.
				statusMsg := fmt.Sprintf("unit %q does not exist, considered disabled", unitName)
				return &resource.Status{
					Status:       statusMsg,
					WarningLevel: resource.StatusWontChange,
					Output:       []string{statusMsg},
				}, nil
			}
		}
		status := &resource.Status{
			WarningLevel: resource.StatusFatal,
			Status:       err.Error(),
		}
		return status, err

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

	err = systemd.MultiErrorAppend(asErr, ufsErr)
	return ufsStatus, err
}

/* Apply sets the properties
1. Apply active state
2. Apply UFS
*/
func (t *Unit) Apply(r resource.Renderer) (resource.TaskStatus, error) {
	/*////////////////////////////////////////////////////
	First thing to do is to check if the Name given is outside
	the normal search path of systemd. If so it should be linked
	*/ ////////////////////////////////////////////////////
	dir, unitName := filepath.Split(t.Name)
	var shouldBeLinked = false
	if dir != "" && t.UnitFileState.IsLinked() {
		// The fullpath to a unit was given. Assume that this isn't in the
		// normal search paths
		shouldBeLinked = true
	}

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

	err = systemd.WaitToLoad(ctx, dbusConn, unitName)
	if err != nil {
		// Unit doesn't exist should be linked
		if err.Error() == "LoadError: \"[\\\"org.freedesktop.DBus.Error.FileNotFound\\\", \\\"No such file or directory\\\"]\"" {
			if shouldBeLinked {
				// Check if file to be linked exist on disk
				_, statErr := os.Stat(t.Name)
				if statErr != nil {
					return &resource.Status{
						Status:       err.Error(),
						WarningLevel: resource.StatusFatal,
						Output:       []string{err.Error()},
					}, err
				}
				// Link the Unit
				changes, e := dbusConn.LinkUnitFiles([]string{t.Name}, t.UnitFileState.IsRuntimeState(), true)
				if e != nil {
					status := &resource.Status{
						WarningLevel: resource.StatusFatal,
						Status:       e.Error(),
					}
					return status, e
				}
				startStatus, e := t.StartOrStop(dbusConn)
				if e != nil {
					return startStatus, e
				}
				statusMsg := fmt.Sprintf("unit %q has been linked", unitName)
				diffMsg := fmt.Sprintf("unit %q does not exist", unitName)
				diff := resource.TextDiff{Default: diffMsg, Values: [2]string{diffMsg, fmt.Sprintf("unit %q is linked", unitName)}}
				return &resource.Status{
					Status:       statusMsg,
					WarningLevel: resource.StatusNoChange,
					WillChange:   false,
					Differences:  map[string]resource.Diff{unitName: diff},
					Output: []string{
						statusMsg,
						fmt.Sprintf("file %q has been linkded to %q", changes[0].Filename, changes[0].Destination),
					},
				}, nil
			}
		}
		status := &resource.Status{
			WarningLevel: resource.StatusFatal,
			Status:       err.Error(),
		}
		return status, err
	}

	// Start the unit
	startStatus, err := t.StartOrStop(dbusConn)
	systemd.ApplyDaemonReload()
	if err != nil {
		return startStatus, err
	}
	// Get the current UnitFileState
	prop, err := dbusConn.GetUnitProperty(unitName, "UnitFileState")
	if err != nil {
		status := &resource.Status{
			WarningLevel: resource.StatusFatal,
			Status:       err.Error(),
		}
		return status, err
	}
	state := systemd.UnitFileState(prop.Value.String())

	//////////////////////////
	// Apply the UnitFileState
	//////////////////////////

	// First thing first, if the state you want is the state you have, quit.
	if t.UnitFileState.Equal(state) {
		statusMsg := fmt.Sprintf("unit %q is %s", unitName, state)
		status := &resource.Status{
			WarningLevel: resource.StatusNoChange,
			Status:       statusMsg,
			Output:       []string{statusMsg},
		}
		return status, err
	}

	switch t.UnitFileState {
	/*///////////////////////////////////
	*ENABLING
	 */ ///////////////////////////////////
	case systemd.UFSEnabled:
		fallthrough
	case systemd.UFSEnabledRuntime:
		// treate enabled and enabled runtime as the same thing
		if state.IsEnabled() {
			statusMsg := fmt.Sprintf("unit %q is enabled, cannot switch between enabled-runtime and enabled", unitName)
			status := &resource.Status{
				Status: statusMsg,
				Output: []string{statusMsg},
			}
			return status, nil
		}
		// check if unit is in a state that cannot be enabled
		if state.IsLinked() || state.Equal(systemd.UFSInvalid) || state.Equal(systemd.UFSBad) {
			statusMsg := fmt.Sprintf("unit %q cannot be enabled. state is %s", unitName, state)
			status := &resource.Status{
				Status:       statusMsg,
				WarningLevel: resource.StatusFatal,
				WillChange:   false,
				Output: []string{
					statusMsg,
				},
			}
			return status, fmt.Errorf(statusMsg)
		}
		// At this point state must be disabled, static, or masked
		if state.Equal(systemd.UFSStatic) { // If static do nothing
			statusMsg := fmt.Sprintf("uint %q does not need to be enabled. state is static", unitName)
			status := &resource.Status{
				Status: statusMsg,
				Output: []string{
					statusMsg,
				},
			}
			return status, err
		}
		if state.IsMaskedState() { // If masked unmask
			changes, err := dbusConn.UnmaskUnitFiles([]string{unitName}, state.IsRuntimeState())
			if err != nil {
				status := &resource.Status{
					WarningLevel: resource.StatusFatal,
					Status:       err.Error(),
					Output:       []string{err.Error()},
				}
				return status, err
			}
			statusMsg := fmt.Sprintf("unit %q is unmasked", unitName)
			status := &resource.Status{
				Status: statusMsg,
				Output: []string{
					statusMsg,
					fmt.Sprintf("symlink for %q was %q. link between %q and %q", unitName, changes[0].Type, changes[0].Filename, changes[0].Destination),
				},
			}
			return status, nil
		}
		// if hasEnaablementInfo is false then the unit will become static
		hasEnaablementInfo, changes, err := dbusConn.EnableUnitFiles([]string{unitName}, t.UnitFileState.IsRuntimeState(), true)
		if err != nil {
			status := &resource.Status{
				WarningLevel: resource.StatusFatal,
				Status:       err.Error(),
				Output:       []string{err.Error()},
			}
			return status, err
		}
		ufs := t.UnitFileState
		if hasEnaablementInfo {
			ufs = "static"
		}
		statusMsg := fmt.Sprintf("unit %q is %s", unitName, ufs)
		status := &resource.Status{
			Status: statusMsg,
			Output: []string{
				statusMsg,
				fmt.Sprintf("symlink for %q was %q. link between %q and %q", unitName, changes[0].Type, changes[0].Filename, changes[0].Destination),
			},
		}
		return status, nil

	/*///////////////////////////////////
	*Disabling
	 */ ///////////////////////////////////
	case systemd.UFSDisabled:
		if state.IsMaskedState() {
			statusMsg := fmt.Sprintf("unit %q is maked", unitName)
			return &resource.Status{
				Status: statusMsg,
				Output: []string{statusMsg},
			}, nil
		}
		if state.Equal(systemd.UFSStatic) {
			changes, err := dbusConn.MaskUnitFiles([]string{unitName}, false, true)
			if err != nil {
				status := &resource.Status{
					WarningLevel: resource.StatusFatal,
					Status:       err.Error(),
					Output:       []string{err.Error()},
				}
				return status, err
			}
			statusMsg := fmt.Sprintf("unit %q is maksed", unitName)
			status := &resource.Status{
				Status: statusMsg,
				Output: []string{
					statusMsg,
					fmt.Sprintf("symlink for %q was %q. link between %q and %q", unitName, changes[0].Type, changes[0].Filename, changes[0].Destination),
				},
			}
			return status, err
		}

		_, err := dbusConn.DisableUnitFiles([]string{unitName}, state.IsRuntimeState())
		if err != nil {
			status := &resource.Status{
				WarningLevel: resource.StatusFatal,
				Status:       err.Error(),
				Output:       []string{err.Error()},
			}
			return status, err
		}
		statusMsg := fmt.Sprintf("unit %q has been disabled", unitName)
		return &resource.Status{
			Status: statusMsg,
			Output: []string{
				statusMsg,
			},
		}, nil
	/*///////////////////////////////////
	*Linking
	 */ ///////////////////////////////////
	case systemd.UFSLinked:
		fallthrough
	case systemd.UFSLinkedRuntime:
		// must change from linked to linked-runtime or vice versa
		if state.IsLinked() {
			_, err := dbusConn.DisableUnitFiles([]string{unitName}, state.IsRuntimeState())
			if err != nil {
				status := &resource.Status{
					WarningLevel: resource.StatusFatal,
					Status:       err.Error(),
					Output:       []string{err.Error()},
				}
				return status, err
			}

			changes, e := dbusConn.LinkUnitFiles([]string{t.Name}, t.UnitFileState.IsRuntimeState(), true)
			if e != nil {
				status := &resource.Status{
					WarningLevel: resource.StatusFatal,
					Status:       e.Error(),
				}
				return status, e
			}
			statusMsg := fmt.Sprintf("unit %q has been linked", unitName)
			diffMsg := fmt.Sprintf("unit %q does not exist", unitName)
			diff := resource.TextDiff{Default: diffMsg, Values: [2]string{diffMsg, fmt.Sprintf("unit %q is linked", unitName)}}
			return &resource.Status{
				Status:       statusMsg,
				WarningLevel: resource.StatusNoChange,
				WillChange:   false,
				Differences:  map[string]resource.Diff{unitName: diff},
				Output: []string{
					statusMsg,
					fmt.Sprintf("file %q has been linkded to %q", changes[0].Filename, changes[0].Destination),
				},
			}, nil
		}

	}

	statusMsg := fmt.Sprintf("unit %q is in unknown state, %s", unitName, state)
	return &resource.Status{
		Status:       statusMsg,
		WarningLevel: resource.StatusFatal,
		Output:       []string{statusMsg},
	}, fmt.Errorf(statusMsg)
}

func (t *Unit) StartOrStop(dbusConn *dbus.Conn) (*resource.Status, error) {
	var err error
	_, unitName := filepath.Split(t.Name)
	// Apply the activeState
	job := make(chan string)
	if t.Active {
		_, err = dbusConn.ReloadOrRestartUnit(unitName, string(t.StartMode), job)
	} else {
		_, err = dbusConn.StopUnit(unitName, string(t.StartMode), job)
	}
	if err != nil {
		status := &resource.Status{
			WarningLevel: resource.StatusFatal,
			Status:       err.Error(),
		}
		return status, err
	}
	<-job
	activeMsg := "inactive"
	if t.Active {
		activeMsg = "active"
	}
	statusMsg := fmt.Sprintf("unit %q is %s", unitName, activeMsg)
	return &resource.Status{
		Status: statusMsg,
		Output: []string{statusMsg},
	}, nil
}

var validUnitFileStates = systemd.UnitFileStates{systemd.UFSEnabled, systemd.UFSEnabledRuntime, systemd.UFSLinked, systemd.UFSLinkedRuntime, systemd.UFSDisabled}

func (t *Unit) Validate() error {
	if t.Name == "" {
		return fmt.Errorf("task requires a %q parameter", "name")
	}
	if !systemd.IsValidUnitFileState(t.UnitFileState) {
		return fmt.Errorf("invalid %q parameter. can be one of [%s]", "state", validUnitFileStates)
	}
	if !systemd.IsValidStartMode(t.StartMode) {
		return fmt.Errorf("task's parameter %q is not one of %s, is %q", "mode", systemd.ValidStartModes, t.StartMode)
	}
	return nil
}
