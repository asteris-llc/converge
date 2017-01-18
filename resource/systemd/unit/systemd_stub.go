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

// +build !linux

package unit

import "errors"

// StubExecutor provides a struct for stub functions that return
// ErrUnsupportedOS
type StubExecutor struct{}

// ErrUnsupportedOS is returned by every function attached to StubExecutor
var ErrUnsupportedOS = errors.New("Error: Unsupported OS. Systemd is only supported on Linux systems.")

// ListUnits is a stub
func (s StubExecutor) ListUnits() (units []string, err error) {
	err = ErrUnsupportedOS
	return
}

// QueryUnit is a stub
func (s StubExecutor) QueryUnit(string) (u Unit, err error) {
	err = ErrUnsupportedOS
	return
}

// StartUnit is a stub
func (s StubExecutor) StartUnit(Unit) error {
	return ErrUnsupportedOS
}

// StopUnit is a stub
func (s StubExecutor) StopUnit(Unit) error {
	return ErrUnsupportedOS
}

// RestartUnit is a stub
func (s StubExecutor) RestartUnit(Unit) error {
	return ErrUnsupportedOS
}

// ReloadUnit is a stub
func (s StubExecutor) ReloadUnit(Unit) error {
	return ErrUnsupportedOS
}

// UnitStatus is a stub
func (s StubExecutor) UnitStatus(Unit) (u Unit, err error) {
	err = ErrUnsupportedOS
	return
}

func realExecutor() (SystemdExecutor, error) {
	return StubExecutor{}, ErrUnsupportedOS
}
