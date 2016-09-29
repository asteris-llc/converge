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

	resource.TaskStatus
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
	dir, unitName := filepath.Split(t.Name)
	shouldBeLinked := false
	if dir != "" && t.UnitFileState.IsLinked() {
		// The fullpath to a unit was given. Assume that this isn't in the
		// normal search paths
		shouldBeLinked = true
	}

	// Get the connection from the pool
	conn, err := systemd.GetDbusConnection()
	if err != nil {
		t.TaskStatus = &resource.Status{
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

	if err != nil {
		// Unit doesn't exist should be linked
		if err.Error() == loadError {
			if shouldBeLinked {
				// Check if file to be linked exist on disk
				_, statErr := os.Stat(t.Name)
				if statErr != nil {
					t.TaskStatus = &resource.Status{
						Level:  resource.StatusFatal,
						Output: []string{statErr.Error()},
					}
					return t, statErr
				}
				statusMsg := fmt.Sprintf("unit %q does not exist, will be linked", unitName)
				diffMsg := fmt.Sprintf("unit %q does not exist", unitName)
				diff := resource.TextDiff{Default: diffMsg, Values: [2]string{diffMsg, fmt.Sprintf("unit %q is linked", unitName)}}
				t.TaskStatus = &resource.Status{
					Level:       resource.StatusWillChange,
					Differences: map[string]resource.Diff{unitName: diff},
					Output:      []string{statusMsg},
				}
				return t, nil
			} else if t.UnitFileState.Equal(systemd.UFSDisabled) {
				// Consider it the same as it being disabled.
				statusMsg := fmt.Sprintf("unit %q does not exist, considered disabled", unitName)
				t.TaskStatus = &resource.Status{
					Level:  resource.StatusWontChange,
					Output: []string{statusMsg},
				}
				return t, nil
			}
		}
		t.TaskStatus = &resource.Status{
			Level:  resource.StatusFatal,
			Output: []string{err.Error()},
		}
		return t, err

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

	if ufsErr != nil {
		if asErr != nil {
			err = multierror.Append(asErr, asErr)
		}
		err = ufsErr
	}

	t.TaskStatus = ufsStatus
	return t, err
}

// Apply sets the properties
// Apply active state
// Apply UFS
func (t *Unit) Apply() (resource.TaskStatus, error) {
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
		t.TaskStatus = &resource.Status{
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
	if err != nil {
		// Unit doesn't exist should be linked
		if err.Error() == loadError {
			if shouldBeLinked {
				// Check if file to be linked exist on disk
				_, statErr := os.Stat(t.Name)
				if statErr != nil {
					t.TaskStatus = &resource.Status{
						Level:  resource.StatusFatal,
						Output: []string{err.Error()},
					}
					return t, err
				}
				// Link the Unit
				changes, e := dbusConn.LinkUnitFiles([]string{t.Name}, t.UnitFileState.IsRuntimeState(), true)
				if e != nil {
					t.TaskStatus = &resource.Status{
						Level:  resource.StatusFatal,
						Output: []string{err.Error()},
					}
					return t, err
				}
				startStatus, e := t.startOrStop(dbusConn)
				if e != nil {
					return startStatus, e
				}
				statusMsg := fmt.Sprintf("unit %q has been linked", unitName)
				diffMsg := fmt.Sprintf("unit %q does not exist", unitName)
				diff := resource.TextDiff{Default: diffMsg, Values: [2]string{diffMsg, fmt.Sprintf("unit %q is linked", unitName)}}
				t.TaskStatus = &resource.Status{
					Level:       resource.StatusNoChange,
					Differences: map[string]resource.Diff{unitName: diff},
					Output: []string{
						statusMsg,
						fmt.Sprintf("file %q has been linkded to %q", changes[0].Filename, changes[0].Destination),
					},
				}
				return t, nil
			}
		}
		t.TaskStatus = &resource.Status{
			Level:  resource.StatusFatal,
			Output: []string{err.Error()},
		}
		return t, err
	}

	// Start the unit
	startStatus, err := t.startOrStop(dbusConn)
	systemd.ApplyDaemonReload()
	if err != nil {
		return startStatus, err
	}
	// Get the current UnitFileState
	prop, err := dbusConn.GetUnitProperty(unitName, "UnitFileState")
	if err != nil {
		t.TaskStatus = &resource.Status{
			Level:  resource.StatusFatal,
			Output: []string{err.Error()},
		}
		return t, err
	}
	state := systemd.UnitFileState(prop.Value.String())

	//////////////////////////
	// Apply the UnitFileState
	//////////////////////////

	// First thing first, if the state you want is the state you have, quit.
	if t.UnitFileState.Equal(state) {
		statusMsg := fmt.Sprintf("unit %q is %s", unitName, state)
		t.TaskStatus = &resource.Status{
			Level:  resource.StatusNoChange,
			Output: []string{statusMsg},
		}
		return t, nil
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
			t.TaskStatus = &resource.Status{
				Level:  resource.StatusWontChange,
				Output: []string{statusMsg},
			}
			return t, nil
		}
		// check if unit is in a state that cannot be enabled
		if state.IsLinked() || state.Equal(systemd.UFSInvalid) || state.Equal(systemd.UFSBad) {
			statusMsg := fmt.Errorf("unit %q cannot be enabled. state is %s", unitName, state)
			t.TaskStatus = &resource.Status{
				Level: resource.StatusFatal,
				Output: []string{
					statusMsg.Error(),
				},
			}
			return t, statusMsg
		}
		// At this point state must be disabled, static, or masked
		if state.Equal(systemd.UFSStatic) { // If static do nothing
			statusMsg := fmt.Sprintf("uint %q does not need to be enabled. state is static", unitName)
			t.TaskStatus = &resource.Status{
				Level: resource.StatusNoChange,
				Output: []string{
					statusMsg,
				},
			}
			return t, err
		}
		if state.IsMaskedState() { // If masked unmask
			changes, err := dbusConn.UnmaskUnitFiles([]string{unitName}, state.IsRuntimeState())
			if err != nil {
				t.TaskStatus = &resource.Status{
					Level:  resource.StatusFatal,
					Output: []string{err.Error()},
				}
				return t, err
			}
			statusMsg := fmt.Sprintf("unit %q is unmasked", unitName)
			t.TaskStatus = &resource.Status{
				Level: resource.StatusNoChange,
				Output: []string{
					statusMsg,
					fmt.Sprintf("symlink for %q was %q. link between %q and %q", unitName, changes[0].Type, changes[0].Filename, changes[0].Destination),
				},
			}
			return t, nil
		}
		// if hasEnaablementInfo is false then the unit will become static
		hasEnaablementInfo, changes, err := dbusConn.EnableUnitFiles([]string{unitName}, t.UnitFileState.IsRuntimeState(), true)
		if err != nil {
			t.TaskStatus = &resource.Status{
				Level:  resource.StatusFatal,
				Output: []string{err.Error()},
			}
			return t, err
		}
		ufs := t.UnitFileState
		if hasEnaablementInfo {
			ufs = "static"
		}
		statusMsg := fmt.Sprintf("unit %q is %s", unitName, ufs)
		t.TaskStatus = &resource.Status{
			Level: resource.StatusNoChange,
			Output: []string{
				statusMsg,
				fmt.Sprintf("symlink for %q was %q. link between %q and %q", unitName, changes[0].Type, changes[0].Filename, changes[0].Destination),
			},
		}
		return t, nil

	/*///////////////////////////////////
	*Disabling
	 */ ///////////////////////////////////
	case systemd.UFSDisabled:
		if state.IsMaskedState() {
			statusMsg := fmt.Sprintf("unit %q is maked", unitName)
			t.TaskStatus = &resource.Status{
				Level:  resource.StatusNoChange,
				Output: []string{statusMsg},
			}
			return t, nil
		}
		if state.Equal(systemd.UFSStatic) {
			changes, err := dbusConn.MaskUnitFiles([]string{unitName}, false, true)
			if err != nil {
				t.TaskStatus = &resource.Status{
					Level:  resource.StatusFatal,
					Output: []string{err.Error()},
				}
				return t, err
			}
			statusMsg := fmt.Sprintf("unit %q is maksed", unitName)
			t.TaskStatus = &resource.Status{
				Level: resource.StatusNoChange,
				Output: []string{
					statusMsg,
					fmt.Sprintf("symlink for %q was %q. link between %q and %q", unitName, changes[0].Type, changes[0].Filename, changes[0].Destination),
				},
			}
			return t, err
		}

		_, err := dbusConn.DisableUnitFiles([]string{unitName}, state.IsRuntimeState())
		if err != nil {
			t.TaskStatus = &resource.Status{
				Level:  resource.StatusFatal,
				Output: []string{err.Error()},
			}
			return t, err
		}
		statusMsg := fmt.Sprintf("unit %q has been disabled", unitName)
		t.TaskStatus = &resource.Status{
			Level: resource.StatusNoChange,
			Output: []string{
				statusMsg,
			},
		}
		return t, nil
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
				t.TaskStatus = &resource.Status{
					Level:  resource.StatusFatal,
					Output: []string{err.Error()},
				}
				return t, err
			}

			changes, e := dbusConn.LinkUnitFiles([]string{t.Name}, t.UnitFileState.IsRuntimeState(), true)
			if e != nil {
				t.TaskStatus = &resource.Status{
					Level:  resource.StatusFatal,
					Output: []string{e.Error()},
				}
				return t, e
			}
			statusMsg := fmt.Sprintf("unit %q has been linked", unitName)
			diffMsg := fmt.Sprintf("unit %q does not exist", unitName)
			diff := resource.TextDiff{Default: diffMsg, Values: [2]string{diffMsg, fmt.Sprintf("unit %q is linked", unitName)}}
			t.TaskStatus = &resource.Status{
				Level:       resource.StatusNoChange,
				Differences: map[string]resource.Diff{unitName: diff},
				Output: []string{
					statusMsg,
					fmt.Sprintf("file %q has been linkded to %q", changes[0].Filename, changes[0].Destination),
				},
			}
			return t, nil
		}

	}

	statusMsg := fmt.Errorf("unit %q is in unknown state, %s", unitName, state)
	t.TaskStatus = &resource.Status{
		Level:  resource.StatusFatal,
		Output: []string{statusMsg.Error()},
	}
	return t, statusMsg
}

func (t *Unit) startOrStop(dbusConn *dbus.Conn) (resource.TaskStatus, error) {
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
		t.TaskStatus = &resource.Status{
			Level:  resource.StatusFatal,
			Output: []string{err.Error()},
		}
		return t, err
	}
	<-job
	activeMsg := "inactive"
	if t.Active {
		activeMsg = "active"
	}
	statusMsg := fmt.Sprintf("unit %q is %s", unitName, activeMsg)
	t.TaskStatus = &resource.Status{
		Level:  resource.StatusNoChange,
		Output: []string{statusMsg},
	}
	return t, nil
}
