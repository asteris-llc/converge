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

// LinuxExecutor provides a command executor for interacting with systemd on Linux
type LinuxExecutor struct{}

// ListUnits will use dbus to get a list of all units
func (l LinuxExecutor) ListUnits() ([]*Unit, error) {
	var units []*Unit
	conn, err := dbus.New()
	if err != nil {
		return units, err
	}
	defer conn.Close()
	unitStatuses, err := conn.ListUnits()
	if err != nil {
		return units, err
	}
	for _, status := range unitStatuses {
		properties, err := conn.GetUnitProperties(status.Name)
		if err != nil {
			return units, err
		}
		typeProperties, err := conn.GetUnitTypeProperties(u.Name, u.Type.UnitTypeString())
		if err != nil {
			return units, err
		}
		u := newFromStatus(&status, properties, typeProperties)
		units = append(units, u)
	}
	return units, nil
}

// QueryUnit will use dbus to get the unit status by name
func (l LinuxExecutor) QueryUnit(string) (Unit, error) {
	return Unit{}, nil
}

// StartUnit will use dbus to start a unit
func (l LinuxExecutor) StartUnit(Unit) error {
	return nil
}

// StopUnit will use dbus to stop a unit
func (l LinuxExecutor) StopUnit(Unit) error {
	return nil
}

// RestartUnit will use dbus to restart a unit
func (l LinuxExecutor) RestartUnit(Unit) error {
	return nil
}

// ReloadUnit will use dbus to reload a unit
func (l LinuxExecutor) ReloadUnit(Unit) error {
	return nil
}

// UnitStatus will use dbus to get the unit status
func (l LinuxExecutor) UnitStatus(Unit) (Unit, error) {
	return Unit{}, nil
}

func realExecutor() (SystemdExecutor, error) {
	return LinuxExecutor{}, nil
}
