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

import (
	"fmt"

	"github.com/coreos/go-systemd/dbus"
	"github.com/pkg/errors"
)

// LinuxExecutor provides a command executor for interacting with systemd on Linux
type LinuxExecutor struct {
	dbusConn SystemdConnection
}

// ListUnits will use dbus to get a list of all units
func (l LinuxExecutor) ListUnits() ([]*Unit, error) {
	var units []*Unit
	unitStatuses, err := l.dbusConn.ListUnits()

	if err != nil {
		return units, err
	}

	for _, status := range unitStatuses {
		unit, err := unitFromStatus(l.dbusConn, &status)
		if err != nil {
			return units, err
		}
		units = append(units, unit)
	}
	return units, nil
}

// QueryUnit will use dbus to get the unit status by name
func (l LinuxExecutor) QueryUnit(unitName string, verify bool) (*Unit, error) {
	units, err := l.ListUnits()
	var toReturn *Unit

	if err != nil {
		return nil, errors.Wrap(err, "Cannot query unit by name")
	}

	for _, u := range units {
		if u.Name == unitName {
			toReturn = u
		}
	}

	if toReturn == nil {
		if verify {
			return nil, fmt.Errorf("%s: no such unit known", unitName)
		}
		toReturn = &Unit{ActiveState: "unknown"}
	}
	return toReturn, nil
}

// StartUnit will use dbus to start a unit
func (l LinuxExecutor) StartUnit(u *Unit) error {
	return runDbusCommand(l.dbusConn.StartUnit, u.Name, "replace", "starting")
}

// StopUnit will use dbus to stop a unit
func (l LinuxExecutor) StopUnit(u *Unit) error {
	return runDbusCommand(l.dbusConn.StopUnit, u.Name, "replace", "stopping")
}

// RestartUnit will restart a unit
func (l LinuxExecutor) RestartUnit(u *Unit) error {
	return runDbusCommand(l.dbusConn.RestartUnit, u.Name, "replace", "restarting")
}

// ReloadUnit will use dbus to reload a unit
func (l LinuxExecutor) ReloadUnit(u *Unit) error {
	return runDbusCommand(l.dbusConn.ReloadUnit, u.Name, "replace", "reloading")
}

// SendSignal will send a signal
func (l LinuxExecutor) SendSignal(u *Unit, signal Signal) {
	l.dbusConn.KillUnit(u.Name, int32(signal))
}

func runDbusCommand(f func(string, string, chan<- string) (int, error), name, mode, operation string) error {
	ch := make(chan string)
	defer close(ch)
	_, err := f(name, mode, ch)
	if err != nil {
		return err
	}
	msg := <-ch
	switch msg {
	case "done":
		return nil
	case "canceled":
		return fmt.Errorf("operation was cancelled while %s: %s", operation, name)
	case "timeout":
		return fmt.Errorf("operation timed out while %s: %s", operation, name)
	case "failed":
		return fmt.Errorf("operation failed while %s: %s", operation, name)
	case "dependency":
		return fmt.Errorf("operation depends on a failed unit when %s: %s", operation, name)
	case "skipped":
		return nil
	}
	return fmt.Errorf("unknown systemd status: %s", msg)
}

func realExecutor() (SystemdExecutor, error) {
	conn, err := dbus.New()
	if err != nil {
		return nil, err
	}
	return LinuxExecutor{conn}, nil
}

// Close will close a connection
func (l LinuxExecutor) Close() {
	l.dbusConn.Close()
}

// NewSystemExecutor will generate a new real executor
func NewSystemExecutor() SystemdExecutor {
	executor, err := realExecutor()
	if err != nil {
		panic(err)
	}
	return executor
}

func unitFromStatus(conn SystemdConnection, status *dbus.UnitStatus) (*Unit, error) {
	u := newFromStatus(status)

	properties, err := conn.GetUnitProperties(status.Name)
	if err != nil {
		return nil, err
	}
	u.SetProperties(properties)

	if u.Type.HasProperties() {
		typeProperties, err := conn.GetUnitTypeProperties(status.Name, u.Type.UnitTypeString())
		if err != nil {
			return nil, err
		}
		u.SetTypedProperties(typeProperties)
	}

	return u, nil
}

func (l LinuxExecutor) unitExists(unitName string) (bool, error) {
	units, err := l.ListUnits()
	if err != nil {
		return false, err
	}
	for _, u := range units {
		if u.Name == unitName {
			return true, nil
		}
	}
	return false, nil
}
