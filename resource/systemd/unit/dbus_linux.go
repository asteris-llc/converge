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

type SystemdConnection interface {
	Close()
	ListUnits() ([]dbus.UnitStatus, error)
	ListUnitsByNames(units []string) ([]dbus.UnitStatus, error)
	GetUnitProperties(unit string) (map[string]interface{}, error)
	GetUnitTypeProperties(unit, unitType string) (map[string]interface{}, error)
	StartUnit(name string, mode string, ch chan<- string) (int, error)
}
