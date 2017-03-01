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

// +build linux

package unit

import "github.com/coreos/go-systemd/dbus"

// SystemdConnection is a lightweight mock over *dbus.Connection
type SystemdConnection interface {
	// Close a systemd connection
	Close()

	// ListUnits that are currently loaded
	ListUnits() ([]dbus.UnitStatus, error)

	// ListUnitsByNames that they are known by
	ListUnitsByNames(units []string) ([]dbus.UnitStatus, error)

	// GetUnitProperties that are global to all unit types
	GetUnitProperties(unit string) (map[string]interface{}, error)

	// GetUnitTypeProperties that are specific to the specified unit type
	GetUnitTypeProperties(unit, unitType string) (map[string]interface{}, error)

	// StartUnit starts a unit
	StartUnit(name string, mode string, ch chan<- string) (int, error)

	// StopUnit stops a unit
	StopUnit(name string, mode string, ch chan<- string) (int, error)

	// RestartUnit restarts a unit
	RestartUnit(name string, mode string, ch chan<- string) (int, error)

	// ReloadUnit instructs a unit to reload it's configuration file
	ReloadUnit(name string, mode string, ch chan<- string) (int, error)

	// KillUnit sends a unix signal to the process
	KillUnit(name string, signal int32)

	EnableUnitFiles(files []string, runtime bool, force bool) (bool, []dbus.EnableUnitFileChange, error)

	DisableUnitFiles(files []string, runtime bool) ([]dbus.DisableUnitFileChange, error)
}
