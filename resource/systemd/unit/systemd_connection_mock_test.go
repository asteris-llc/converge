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
	"github.com/coreos/go-systemd/dbus"
	"github.com/stretchr/testify/mock"
)

type SystemdMock struct {
	mock.Mock
}

func (m *SystemdMock) Close() {
	m.Called()
	return
}

func (m *SystemdMock) ListUnits() ([]dbus.UnitStatus, error) {
	args := m.Called()
	return args.Get(0).([]dbus.UnitStatus), args.Error(1)
}

func (m *SystemdMock) ListUnitsByNames(units []string) ([]dbus.UnitStatus, error) {
	args := m.Called(units)
	return args.Get(0).([]dbus.UnitStatus), args.Error(1)
}

func (m *SystemdMock) GetUnitProperties(unit string) (map[string]interface{}, error) {
	args := m.Called(unit)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *SystemdMock) GetUnitTypeProperties(unit, unitType string) (map[string]interface{}, error) {
	args := m.Called(unit)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func makeUnit(name, activeState string, unitType UnitType) dbus.UnitStatus {
	if UnitTypeFromName(name) != unitType {
		name = name + "." + unitType.Suffix()
	}
	return dbus.UnitStatus{
		Name:        name,
		Description: name,
		ActiveState: activeState,
	}
}

var samplePropertiesMap = map[string]interface{}{
	"Id": "id1",
}
