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

import "github.com/coreos/go-systemd/dbus"
import godbus "github.com/godbus/dbus"

func propDependency(name string, units []string) *dbus.Property {
	return &dbus.Property{
		Name:  name,
		Value: godbus.MakeVariant(units),
	}
}

type LoadState string
type LoadStates []LoadState

const (
	LSLoaded LoadState = "loaded"
	LSError  LoadState = "error"
	LSMasked LoadState = "masked"
)

func PropLoadState(ls LoadState) *dbus.Property {
	return &dbus.Property{
		Name:  "LoadState",
		Value: godbus.MakeVariant(string(ls)),
	}
}

func (states LoadStates) Contains(state LoadState) bool {
	for _, s := range states {
		if s == state {
			return true
		}
	}
	return false
}

type ActiveState string
type ActiveStates []ActiveState

const (
	ASActive       ActiveState = "active"
	ASReloading    ActiveState = "reloading"
	ASInactive     ActiveState = "inactive"
	ASFailed       ActiveState = "failed"
	ASActivating   ActiveState = "activating"
	ASDeactivating ActiveState = "deactivating"
)

func (states ActiveStates) Contains(state ActiveState) bool {
	for _, s := range states {
		if s == state {
			return true
		}
	}
	return false
}

func PropActiveState(as ActiveState) *dbus.Property {
	return &dbus.Property{
		Name:  "ActiveState",
		Value: godbus.MakeVariant(string(as)),
	}
}

type UnitFileState string
type UnitFileStates []UnitFileState

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
)

func PropUnitFileState(ufs UnitFileState) *dbus.Property {
	return &dbus.Property{
		Name:  "UnitFileState",
		Value: godbus.MakeVariant(string(ufs)),
	}
}

// Determines whether unit should be enabled
func (state UnitFileState) IsEnabled() bool {
	return state == UFSLinked || state == UFSLinkedRuntime
}

// Determines whether service should be linked to usual locations
func (state UnitFileState) IsLinked() bool {
	return state == UFSLinked || state == UFSLinkedRuntime
}

// IsRuntimeState returns true if unit should be in the /run folder
func (state UnitFileState) IsRuntimeState() bool {
	return state == UFSEnabledRuntime || state == UFSLinkedRuntime || state == UFSMaskedRuntime
}

// IsMaskedState
func (state UnitFileState) IsMaskedState() bool {
	return state == UFSMasked || state == UFSMaskedRuntime
}

var ValidUnitFileStates = UnitFileStates{UFSEnabled, UFSEnabledRuntime, UFSLinked, UFSLinkedRuntime, UFSMasked, UFSMaskedRuntime, UFSStatic, UFSDisabled, UFSInvalid}
var ValidUnitFileStatesWithoutInvalid = ValidUnitFileStates[:len(ValidUnitFileStates)-1]

func (states UnitFileStates) Contains(s UnitFileState) bool {
	for i, _ := range states {
		if s == states[i] {
			return true
		}
	}
	return false
}

func IsValidUnitFileState(ufs UnitFileState) bool {
	return ValidUnitFileStates.Contains(ufs)
}

type StartMode string
type StartModes []StartMode

const (
	SMReplace            StartMode = "replace"
	SMFail               StartMode = "fail"
	SMIsolate            StartMode = "isolate"
	SMIgnoreDependencies StartMode = "ignore-dependencies"
	SMIgnoreRequirements StartMode = "ignore-requirements"
)

func (states StartModes) Contains(s StartMode) bool {
	for i, _ := range states {
		if s == states[i] {
			return true
		}
	}
	return false
}

var ValidStartModes = StartModes{SMReplace, SMFail, SMIsolate, SMIgnoreDependencies, SMIgnoreRequirements}

func IsValidStartMode(sm StartMode) bool {
	return ValidStartModes.Contains(sm)
}
