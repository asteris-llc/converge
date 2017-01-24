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

package unit_test

import (
	"github.com/asteris-llc/converge/resource/systemd/unit"
	"github.com/coreos/go-systemd/dbus"
	"github.com/stretchr/testify/mock"
)

// ConnectorMock mocks DbusConnector
type ConnectorMock struct {
	mock.Mock
}

// New generates a new mock
func (m *ConnectorMock) New() (unit.SystemdConnection, error) {
	args := m.Called()
	return args.Get(0).(unit.SystemdConnection), args.Error(1)
}

// DbusMock mocks the actual dbus connection
type DbusMock struct {
	mock.Mock
}

// ListUnits mocks ListUnits
func (m DbusMock) ListUnits() ([]dbus.UnitStatus, error) {
	args := m.Called()
	return args.Get(0).([]dbus.UnitStatus), args.Error(1)
}

// GetUnitProperties mocks GetUnitProperties
func (m DbusMock) GetUnitProperties(unit string) (map[string]interface{}, error) {
	args := m.Called(unit)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

// GetUnitTypeProperties mocks GetUnitTypeProperties
func (m DbusMock) GetUnitTypeProperties(unit, unitType string) (map[string]interface{}, error) {
	args := m.Called(unit, unitType)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

// Close Closes
func (m DbusMock) Close() {
	m.Called()
	return
}

func basicMockConnector() *ConnectorMock {
	m := &ConnectorMock{}
	m.On("New").Return(&DbusMock{}, nil)
	return m
}

type rets struct {
	Val interface{}
	Err error
}

type unitInfo struct {
	Unit      dbus.UnitStatus
	Props     map[string]interface{}
	TypeProps map[string]interface{}
}

// func dbusMock(returns map[string]rets) *DbusMock {

//	u := []dbus.UnitState{defaultUnit}
//	m := &DbusMock{}
//	if ret, ok := returns["ListUnits"]; ok {
//		m.On("ListUnits").Return(ret.Val, ret.Err)
//	} else {
//		m.On("ListUnits").Return(u, nil)
//	}
//	if ret, ok := returns["GetUnitProperties"]; ok {
//		m.On("GetUnitProperties", mock.Anything).Return(ret.Val, ret.Err)
//	} else {
//		m.On("GetUnitProperties", mock.Anything).Return(map[string]interface{}{}, nil)
//	}
// }

// var defaultUnit = makeUnitStatus("unit1", "description1", "loaded", "active", "/org.freedesktop.system1/")

// func makeUnitStatus(name, description, loadstate, activestate, path string) dbus.UnitStatus {
//	return &dbus.UnitStatus{
//		Name:        name,
//		Description: description,
//		LoadState:   loadstate,
//		ActiveState: activestate,
//		Path:        path,
//	}
// }
