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
	"runtime"
	"sync"
)

// CheckDaemonReload checks whether the systemd daemon needs to be reloaded.
// it does not check if a unit is failed and needs to be reset though.
func CheckDaemonReload(unit string) (bool, error) {
	conn, err := GetDbusConnection()
	if err != nil {
		return false, err
	}
	prop, err := conn.Connection.GetUnitProperty(unit, "NeedDaemonReload")
	if err != nil {
		return false, err
	}
	shouldReload, ok := prop.Value.Value().(bool)
	return shouldReload && ok, nil
}

var reloadToken = make(chan struct{}, 1)
var reloadOnce sync.Once

func applyDaemonReload() error {
	conn, err := GetDbusConnection()
	if err != nil {
		return err
	}
	defer conn.Return()
	err = conn.Connection.Reload()
	return err
}

// ApplyDaemonReload reloads the daemon
func ApplyDaemonReload() (err error) {
	reloadOnce.Do(func() {
		reloadToken <- struct{}{}
	})
	for true {
		select {
		case canReload := <-reloadToken:
			err = applyDaemonReload()
			reloadToken <- canReload
			return err
		default:
			runtime.Gosched()
		}
	}
	return err
}

// CheckResetFailed checks whether the fail state of the unit should be
// reset
func CheckResetFailed(unit string) (bool, error) {
	conn, err := GetDbusConnection()
	if err != nil {
		return false, err
	}
	defer conn.Return()
	prop, err := conn.Connection.GetUnitProperty(unit, "ActiveState")
	if err != nil {
		return false, err
	}
	shouldReset, ok := prop.Value.Value().(string)
	return ok && ASFailed.Equal(ActiveState(shouldReset)), nil
}

var resetToken = make(chan struct{}, 1)
var resetOnce sync.Once

func applyResetFailed(unit string) error {
	conn, err := GetDbusConnection()
	if err != nil {
		return err
	}
	defer conn.Return()
	err = conn.Connection.ResetFailedUnit(unit)
	return err
}

// ApplyResetFailed resets the failed state of a unit.
func ApplyResetFailed(unit string) (err error) {
	resetOnce.Do(func() {
		resetToken <- struct{}{}
	})
	for true {
		select {
		case canReset := <-resetToken:
			err = applyResetFailed(unit)
			resetToken <- canReset
			return err
		default:
			runtime.Gosched()
		}
	}
	return err

}
