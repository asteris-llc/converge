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
	"math/rand"

	"github.com/coreos/go-systemd/dbus"
	"github.com/stretchr/testify/mock"
)

// DbusMock mocks the actual dbus connection
type DbusMock struct {
	mock.Mock
	startResp string
}

// ListUnits mocks ListUnits
func (m *DbusMock) ListUnits() ([]dbus.UnitStatus, error) {
	args := m.Called()
	return args.Get(0).([]dbus.UnitStatus), args.Error(1)
}

// ListUnits mocks ListUnitsByNames
func (m *DbusMock) ListUnitsByNames(names []string) ([]dbus.UnitStatus, error) {
	args := m.Called(names)
	return args.Get(0).([]dbus.UnitStatus), args.Error(1)
}

// GetUnitProperties mocks GetUnitProperties
func (m *DbusMock) GetUnitProperties(unit string) (map[string]interface{}, error) {
	args := m.Called(unit)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

// GetUnitTypeProperties mocks GetUnitTypeProperties
func (m *DbusMock) GetUnitTypeProperties(unit, unitType string) (map[string]interface{}, error) {
	args := m.Called(unit, unitType)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

// Close Closes
func (m *DbusMock) Close() {
	m.Called()
	return
}

func (m *DbusMock) StartUnit(name string, mode string, ch chan<- string) (int, error) {
	fmt.Println("StartUnit was called...")
	args := m.Called(name, mode, ch)
	if m.startResp != "" && ch != nil {
		go func() {
			ch <- m.startResp
		}()
	}
	return args.Int(0), args.Error(1)
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

func randomUnitStatus() dbus.UnitStatus {
	loadState := loadStates[rand.Intn(len(loadStates))]
	activeState := activeStates[rand.Intn(len(activeStates))]
	suffix := validTypes[rand.Intn(len(validTypes))]
	var nameBytes []byte
	for i := 0; i < 64; i++ {
		nameBytes = append(nameBytes, alphabet[rand.Intn(len(alphabet))])
	}
	name := fmt.Sprintf("%s.%s", string(nameBytes), suffix)
	return dbus.UnitStatus{
		Name:        name,
		Description: name,
		LoadState:   loadState,
		ActiveState: activeState,
	}
}

func randomUnit(typeVal UnitType) *Unit {
	st := randomUnitStatus()

	name := fmt.Sprintf("%s.%s", st.Name, typeVal.Suffix())

	u := &Unit{
		Properties:  &Properties{},
		Name:        name,
		Description: name,
		ActiveState: st.ActiveState,
		LoadState:   st.LoadState,
		Type:        typeVal,
	}

	switch typeVal {
	case UnitTypeService:
		u.ServiceProperties = &ServiceTypeProperties{}
	case UnitTypeSocket:
		u.SocketProperties = &SocketTypeProperties{}
	case UnitTypeDevice:
		u.DeviceProperties = &DeviceTypeProperties{}
	case UnitTypeMount:
		u.MountProperties = &MountTypeProperties{}
	case UnitTypeAutoMount:
		u.AutomountProperties = &AutomountTypeProperties{}
	case UnitTypeSwap:
		u.SwapProperties = &SwapTypeProperties{}
	case UnitTypePath:
		u.PathProperties = &PathTypeProperties{}
	case UnitTypeTimer:
		u.TimerProperties = &TimerTypeProperties{}
	case UnitTypeSlice:
		u.SliceProperties = &SliceTypeProperties{}
	case UnitTypeScope:
		u.ScopeProperties = &ScopeTypeProperties{}
	}
	return u
}
