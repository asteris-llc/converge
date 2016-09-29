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
// limitations under the License.package systemd

package systemd

import (
	"strings"

	"github.com/coreos/go-systemd/dbus"
	godbus "github.com/godbus/dbus"
)

func propDependency(name string, units []string) *dbus.Property {
	return &dbus.Property{
		Name:  name,
		Value: godbus.MakeVariant(units),
	}
}

// LoadState reflects dbus values for a unit's "LoadState" property
type LoadState string

// LoadStates defines a slice of LoadState
type LoadStates []LoadState

const (
	// LSLoaded says that the unit is loaded
	LSLoaded LoadState = "loaded"
	// LSError says that an error occured when loading
	LSError LoadState = "error"
	// LSMasked says that the unit is masked
	LSMasked LoadState = "masked"
	// LSLoaded says that the unit is no longer found
	// LSError will happen when the unit can't be loaded because the unit file
	// is not found. This happens when the unit is already loaded, but the
	// unit file is no longer found
	LSNotFound LoadState = "not-found"
)

// PropActiveState creates a valid `*dbus.Property` with the given LoadState
func PropLoadState(ls LoadState) *dbus.Property {
	return &dbus.Property{
		Name:  "LoadState",
		Value: godbus.MakeVariant(string(ls)),
	}
}

// Contains checks if a slice of LoadState has a specific LoadState
func (states LoadStates) Contains(state LoadState) bool {
	for _, s := range states {
		if s == state {
			return true
		}
	}
	return false
}

// ActiveState reflects dbus values for a unit's "ActiveState" property
type ActiveState string

// ActiveState defines a slice of ActiveState
type ActiveStates []ActiveState

// This block defines string values for properties given by dbus
const (
	ASActive       ActiveState = "active"
	ASReloading    ActiveState = "reloading"
	ASInactive     ActiveState = "inactive"
	ASFailed       ActiveState = "failed"
	ASActivating   ActiveState = "activating"
	ASDeactivating ActiveState = "deactivating"
)

// Equal checks if two ActiveStates are equal. It automatically strips quotes.
func (state ActiveState) Equal(other ActiveState) bool {
	if state == other {
		return true
	}
	stateStr := strings.Trim(string(state), "\"")
	otherStr := strings.Trim(string(other), "\"")
	return stateStr == otherStr
}

// Contains checks if in a list of ActiveState one state is the given
func (states ActiveStates) Contains(state ActiveState) bool {
	for _, s := range states {
		if s == state {
			return true
		}
	}
	return false
}

// PropActiveState creates a valid `*dbus.Property` with the given ActiveState
func PropActiveState(as ActiveState) *dbus.Property {
	return &dbus.Property{
		Name:  "ActiveState",
		Value: godbus.MakeVariant(string(as)),
	}
}

// UnitFileState reflects the dbus property "UnitFileState" for a unit
type UnitFileState string

// UnitFileStates defines a slice of UnitFileState
type UnitFileStates []UnitFileState

// This block defines the possible values of the property "UnitFileState"
const (
	UFSEnabled        UnitFileState = "enabled"
	UFSEnabledRuntime UnitFileState = "enabled-runtime"
	UFSLinked         UnitFileState = "linked"
	UFSLinkedRuntime  UnitFileState = "linked-runtime"
	UFSMasked         UnitFileState = "masked"
	UFSMaskedRuntime  UnitFileState = "masked-runtime"
	UFSStatic         UnitFileState = "static"
	UFSDisabled       UnitFileState = "disabled"
	UFSInvalid        UnitFileState = "invalid"
	UFSBad            UnitFileState = "bad"
)

// PropUnitFileState creatues a valid `*dbus.Property` with the given UnitFileState
func PropUnitFileState(ufs UnitFileState) *dbus.Property {
	return &dbus.Property{
		Name:  "UnitFileState",
		Value: godbus.MakeVariant(string(ufs)),
	}
}

// Equal checks if two `UnitFileState`s are equal. Trims quotes.
func (state UnitFileState) Equal(other UnitFileState) bool {
	if state == other {
		return true
	}
	stateStr := strings.Trim(string(state), "\"")
	otherStr := strings.Trim(string(other), "\"")
	return stateStr == otherStr
}

// IsEnabled Determines whether unit should be enabled
func (state UnitFileState) IsEnabled() bool {
	return state.Equal(UFSEnabled) || state.Equal(UFSEnabledRuntime)
}

// IsLinked Determines whether service should be linked to usual locations
func (state UnitFileState) IsLinked() bool {
	return state.Equal(UFSLinked) || state.Equal(UFSLinkedRuntime)
}

// IsRuntimeState returns true if unit should be in the /run folder
func (state UnitFileState) IsRuntimeState() bool {
	return state.Equal(UFSEnabledRuntime) || state.Equal(UFSLinkedRuntime) || state.Equal(UFSMaskedRuntime)
}

// IsMaskedState Determines if a static unit is masked
func (state UnitFileState) IsMaskedState() bool {
	return state.Equal(UFSMasked) || state.Equal(UFSMaskedRuntime)
}

// ValidUnitFileStates are states that a user can make a unit take
var ValidUnitFileStates = UnitFileStates{UFSEnabled, UFSEnabledRuntime, UFSLinked, UFSLinkedRuntime, UFSMasked, UFSMaskedRuntime, UFSStatic, UFSDisabled, UFSInvalid}

// Contains checks if in a list of `UnitFileState`s one state is the given
func (states UnitFileStates) Contains(s UnitFileState) bool {
	for i := range states {
		if s == states[i] {
			return true
		}
	}
	return false
}

// Checks if the `UnitFileState` is one the user can make a unit take
func IsValidUnitFileState(ufs UnitFileState) bool {
	return ValidUnitFileStates.Contains(ufs)
}

// StartMode reflects valid parameters to the `StartUnit` function
type StartMode string

// StartModes defines a slice of `StartMode`
type StartModes []StartMode

// This block defines valid paramaters to the `StartUnit` function
const (
	SMReplace            StartMode = "replace"
	SMFail               StartMode = "fail"
	SMIsolate            StartMode = "isolate"
	SMIgnoreDependencies StartMode = "ignore-dependencies"
	SMIgnoreRequirements StartMode = "ignore-requirements"
)

// Contains checks if the given `StartMode` is in a list of `StartModes`
func (states StartModes) Contains(s StartMode) bool {
	for i := range states {
		if s == states[i] {
			return true
		}
	}
	return false
}

// ValidStartModes is a list of valid paramaters to the `StartUnit` function
var ValidStartModes = StartModes{SMReplace, SMFail, SMIsolate, SMIgnoreDependencies, SMIgnoreRequirements}

// IsValidStartMode checks if a given StartMode is valid
func IsValidStartMode(sm StartMode) bool {
	return ValidStartModes.Contains(sm)
}
