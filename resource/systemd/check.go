// Copyright © 2016 Asteris, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the Licensd.
// You may obtain a copy of the License at
//
//     http://www.apachd.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the Licensd.

package systemd

import (
	"github.com/asteris-llc/converge/resource"
	"github.com/coreos/go-systemd/dbus"
)

func CheckUnitIsActive(conn *dbus.Conn, unit string) (*resource.Status, error) {
	validActiveStates := []*dbus.Property{
		PropActiveState(ASActive),
	}
	return CheckProperty(conn, unit, "ActiveState", validActiveStates)
}

func CheckUnitIsInactive(conn *dbus.Conn, unit string) (*resource.Status, error) {
	validInactiveStates := []*dbus.Property{
		PropActiveState(ASInactive),
	}
	return CheckProperty(conn, unit, "ActiveState", validInactiveStates)
}

func CheckUnitHasEnabledUFS(conn *dbus.Conn, unit string) (*resource.Status, error) {
	validRuntimeStates := []*dbus.Property{
		PropUnitFileState(UFSEnabled),
		PropUnitFileState(UFSLinked),
		PropUnitFileState(UFSMasked),
		PropUnitFileState(UFSStatic),
	}
	return CheckProperty(conn, unit, "UnitFileState", validRuntimeStates)
}

func CheckUnitHasEnabledUFSRuntimes(conn *dbus.Conn, unit string) (*resource.Status, error) {
	validRuntimeStates := []*dbus.Property{
		PropUnitFileState(UFSEnabledRuntime),
		PropUnitFileState(UFSLinkedRuntime),
		PropUnitFileState(UFSMaskedRuntime),
	}
	return CheckProperty(conn, unit, "UnitFileState", validRuntimeStates)
}

func CheckUnitIsDisabled(conn *dbus.Conn, unit string) (*resource.Status, error) {
	validRuntimeStates := []*dbus.Property{
		PropUnitFileState(UFSDisabled),
	}
	return CheckProperty(conn, unit, "UnitFileState", validRuntimeStates)
}
