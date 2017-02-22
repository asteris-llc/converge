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
	"reflect"
	"testing"

	"github.com/pkg/errors"

	"github.com/coreos/go-systemd/dbus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInterface(t *testing.T) {
	//	assert.Implements(t, (*SystemdExecutor)(nil), new(LinuxExecutor))
}

// TestListUnits runs a test
func TestListUnits(t *testing.T) {
	t.Parallel()

	t.Run("when-dbus-errors", func(t *testing.T) {
		t.Run("list-units-returns-error", func(t *testing.T) {
			t.Parallel()
			m := &DbusMock{}
			expected := errors.New("error1")
			m.On("ListUnits").Return([]dbus.UnitStatus{}, expected)
			l := LinuxExecutor{m}
			_, err := l.ListUnits()
			assert.Equal(t, expected, err)
		})

		t.Run("GetUnitProperties-fails", func(t *testing.T) {
			t.Parallel()
			m := &DbusMock{}
			expected := errors.New("error1")
			m.On("ListUnits").Return([]dbus.UnitStatus{dbus.UnitStatus{}}, nil)
			m.On("GetUnitProperties", any).Return(map[string]interface{}{}, expected)
			l := LinuxExecutor{m}
			_, err := l.ListUnits()
			assert.Equal(t, expected, err)
		})
		t.Run("GetUnitTypeProperties-fails", func(t *testing.T) {
			t.Run("when-should-not-have-properties", func(t *testing.T) {
				t.Parallel()
				m := &DbusMock{}
				expected := errors.New("error1")
				m.On("ListUnits").Return([]dbus.UnitStatus{dbus.UnitStatus{}}, nil)
				m.On("GetUnitProperties", any).Return(map[string]interface{}{}, nil)
				m.On("GetUnitTypeProperties", any).Return(map[string]interface{}{}, expected)
				l := LinuxExecutor{m}
				_, err := l.ListUnits()
				assert.NoError(t, err)
			})
			t.Run("when-should-have-type-properties", func(t *testing.T) {
				t.Parallel()
				m := &DbusMock{}
				expected := errors.New("error1")
				m.On("ListUnits").Return([]dbus.UnitStatus{dbus.UnitStatus{Name: "foo.service"}}, nil)
				m.On("GetUnitProperties", any).Return(map[string]interface{}{}, nil)
				m.On("GetUnitTypeProperties", any, any).Return(map[string]interface{}{}, expected)
				l := LinuxExecutor{m}
				_, err := l.ListUnits()
				assert.Equal(t, expected, err)
			})
		})
	})

	t.Run("returns-a-unit-for-each-returned-unit", func(t *testing.T) {
		t.Parallel()
		var units []dbus.UnitStatus
		for i := 0; i < 10; i++ {
			units = append(units, dbus.UnitStatus{})
		}
		m := &DbusMock{}
		m.On("ListUnits").Return(units, nil)
		m.On("GetUnitProperties", any).Return(map[string]interface{}{}, nil)
		m.On("GetUnitTypeProperties", any, any).Return(map[string]interface{}{}, nil)
		l := LinuxExecutor{m}
		actual, err := l.ListUnits()
		assert.NoError(t, err)
		assert.Equal(t, len(units), len(actual))
	})

	t.Run("sets-unit-fields", func(t *testing.T) {
		t.Parallel()

		expected := randomUnitStatus()
		units := []dbus.UnitStatus{expected}
		m := &DbusMock{}
		m.On("ListUnits").Return(units, nil)
		m.On("GetUnitProperties", any).Return(map[string]interface{}{}, nil)
		m.On("GetUnitTypeProperties", any, any).Return(map[string]interface{}{}, nil)
		l := LinuxExecutor{m}
		actualSlice, err := l.ListUnits()
		require.NoError(t, err)
		require.Equal(t, 1, len(actualSlice))
		actual := actualSlice[0]
		assert.Equal(t, expected.Name, actual.Name)
		assert.Equal(t, expected.LoadState, actual.LoadState)
		assert.Equal(t, expected.Description, actual.Description)
		assert.Equal(t, expected.ActiveState, actual.ActiveState)
	})

	t.Run("sets-global-properties", func(t *testing.T) {
		t.Parallel()

		propMap := map[string]interface{}{
			"DefaultDependencies": true,
			"LoadState":           "loaded",
			"Names":               []string{"name1", "name2", "name3"},
			"StartLimitBurst":     uint32(7),
		}

		expected := Properties{
			DefaultDependencies: true,
			LoadState:           "loaded",
			Names:               []string{"name1", "name2", "name3"},
			StartLimitBurst:     uint32(7),
		}

		m := &DbusMock{}
		m.On("ListUnits").Return([]dbus.UnitStatus{dbus.UnitStatus{}}, nil)
		m.On("GetUnitProperties", any).Return(propMap, nil)
		m.On("GetUnitTypeProperties", any, any).Return(map[string]interface{}{}, nil)
		l := LinuxExecutor{m}
		actualSlice, err := l.ListUnits()
		require.NoError(t, err)
		require.Equal(t, 1, len(actualSlice))
		actual := actualSlice[0]
		assert.True(t, reflect.DeepEqual(expected, actual.Properties))
	})

	t.Run("sets-type-properties", func(t *testing.T) {
		t.Parallel()

		t.Run("service", func(t *testing.T) {
			typeName := "service"
			m := &DbusMock{}
			typeProps := map[string]interface{}{}
			m.On("ListUnits").Return([]dbus.UnitStatus{dbus.UnitStatus{Name: "name." + typeName}}, nil)
			m.On("GetUnitProperties", any, any).Return(map[string]interface{}{}, nil)
			m.On("GetUnitTypeProperties", any, any).Return(typeProps, nil)
			l := LinuxExecutor{m}
			actualSlice, err := l.ListUnits()
			require.NoError(t, err)
			require.Equal(t, 1, len(actualSlice))
			actual := actualSlice[0]
			assert.NotNil(t, actual.ServiceProperties)
			assert.Nil(t, actual.SocketProperties)
			assert.Nil(t, actual.DeviceProperties)
			assert.Nil(t, actual.MountProperties)
			assert.Nil(t, actual.AutomountProperties)
			assert.Nil(t, actual.SwapProperties)
			assert.Nil(t, actual.PathProperties)
			assert.Nil(t, actual.TimerProperties)
			assert.Nil(t, actual.SliceProperties)
			assert.Nil(t, actual.ScopeProperties)
		})
		t.Run("socket", func(t *testing.T) {
			typeName := "socket"
			typeProps := map[string]interface{}{}
			m := &DbusMock{}
			m.On("ListUnits").Return([]dbus.UnitStatus{dbus.UnitStatus{Name: ("name." + typeName)}}, nil)
			m.On("GetUnitProperties", any, any).Return(map[string]interface{}{}, nil)
			m.On("GetUnitTypeProperties", any, any).Return(typeProps, nil)
			l := LinuxExecutor{m}
			actualSlice, err := l.ListUnits()
			require.NoError(t, err)
			require.Equal(t, 1, len(actualSlice))
			actual := actualSlice[0]
			assert.Nil(t, actual.ServiceProperties)
			assert.NotNil(t, actual.SocketProperties)
			assert.Nil(t, actual.DeviceProperties)
			assert.Nil(t, actual.MountProperties)
			assert.Nil(t, actual.AutomountProperties)
			assert.Nil(t, actual.SwapProperties)
			assert.Nil(t, actual.PathProperties)
			assert.Nil(t, actual.TimerProperties)
			assert.Nil(t, actual.SliceProperties)
			assert.Nil(t, actual.ScopeProperties)
		})
		t.Run("device", func(t *testing.T) {
			typeName := "device"
			typeProps := map[string]interface{}{}
			m := &DbusMock{}
			m.On("ListUnits").Return([]dbus.UnitStatus{dbus.UnitStatus{Name: ("name." + typeName)}}, nil)
			m.On("GetUnitProperties", any, any).Return(map[string]interface{}{}, nil)
			m.On("GetUnitTypeProperties", any, any).Return(typeProps, nil)
			l := LinuxExecutor{m}
			actualSlice, err := l.ListUnits()
			require.NoError(t, err)
			require.Equal(t, 1, len(actualSlice))
			actual := actualSlice[0]
			assert.Nil(t, actual.ServiceProperties)
			assert.Nil(t, actual.SocketProperties)
			assert.NotNil(t, actual.DeviceProperties)
			assert.Nil(t, actual.MountProperties)
			assert.Nil(t, actual.AutomountProperties)
			assert.Nil(t, actual.SwapProperties)
			assert.Nil(t, actual.PathProperties)
			assert.Nil(t, actual.TimerProperties)
			assert.Nil(t, actual.SliceProperties)
			assert.Nil(t, actual.ScopeProperties)
		})
		t.Run("mount", func(t *testing.T) {
			typeName := "mount"
			typeProps := map[string]interface{}{}
			m := &DbusMock{}
			m.On("ListUnits").Return([]dbus.UnitStatus{dbus.UnitStatus{Name: ("name." + typeName)}}, nil)
			m.On("GetUnitProperties", any, any).Return(map[string]interface{}{}, nil)
			m.On("GetUnitTypeProperties", any, any).Return(typeProps, nil)
			l := LinuxExecutor{m}
			actualSlice, err := l.ListUnits()
			require.NoError(t, err)
			require.Equal(t, 1, len(actualSlice))
			actual := actualSlice[0]
			assert.Nil(t, actual.ServiceProperties)
			assert.Nil(t, actual.SocketProperties)
			assert.Nil(t, actual.DeviceProperties)
			assert.NotNil(t, actual.MountProperties)
			assert.Nil(t, actual.AutomountProperties)
			assert.Nil(t, actual.SwapProperties)
			assert.Nil(t, actual.PathProperties)
			assert.Nil(t, actual.TimerProperties)
			assert.Nil(t, actual.SliceProperties)
			assert.Nil(t, actual.ScopeProperties)
		})
		t.Run("automount", func(t *testing.T) {
			typeName := "automount"
			typeProps := map[string]interface{}{}
			m := &DbusMock{}
			m.On("ListUnits").Return([]dbus.UnitStatus{dbus.UnitStatus{Name: ("name." + typeName)}}, nil)
			m.On("GetUnitProperties", any, any).Return(map[string]interface{}{}, nil)
			m.On("GetUnitTypeProperties", any, any).Return(typeProps, nil)
			l := LinuxExecutor{m}
			actualSlice, err := l.ListUnits()
			require.NoError(t, err)
			require.Equal(t, 1, len(actualSlice))
			actual := actualSlice[0]
			assert.Nil(t, actual.ServiceProperties)
			assert.Nil(t, actual.SocketProperties)
			assert.Nil(t, actual.DeviceProperties)
			assert.Nil(t, actual.MountProperties)
			assert.NotNil(t, actual.AutomountProperties)
			assert.Nil(t, actual.SwapProperties)
			assert.Nil(t, actual.PathProperties)
			assert.Nil(t, actual.TimerProperties)
			assert.Nil(t, actual.SliceProperties)
			assert.Nil(t, actual.ScopeProperties)
		})
		t.Run("swap", func(t *testing.T) {
			typeName := "swap"
			typeProps := map[string]interface{}{}
			m := &DbusMock{}
			m.On("ListUnits").Return([]dbus.UnitStatus{dbus.UnitStatus{Name: ("name." + typeName)}}, nil)
			m.On("GetUnitProperties", any, any).Return(map[string]interface{}{}, nil)
			m.On("GetUnitTypeProperties", any, any).Return(typeProps, nil)
			l := LinuxExecutor{m}
			actualSlice, err := l.ListUnits()
			require.NoError(t, err)
			require.Equal(t, 1, len(actualSlice))
			actual := actualSlice[0]
			assert.Nil(t, actual.ServiceProperties)
			assert.Nil(t, actual.SocketProperties)
			assert.Nil(t, actual.DeviceProperties)
			assert.Nil(t, actual.MountProperties)
			assert.Nil(t, actual.AutomountProperties)
			assert.NotNil(t, actual.SwapProperties)
			assert.Nil(t, actual.PathProperties)
			assert.Nil(t, actual.TimerProperties)
			assert.Nil(t, actual.SliceProperties)
			assert.Nil(t, actual.ScopeProperties)
		})
		t.Run("target", func(t *testing.T) {
			typeName := "target"
			typeProps := map[string]interface{}{}
			m := &DbusMock{}
			m.On("ListUnits").Return([]dbus.UnitStatus{dbus.UnitStatus{Name: ("name." + typeName)}}, nil)
			m.On("GetUnitProperties", any, any).Return(map[string]interface{}{}, nil)
			m.On("GetUnitTypeProperties", any, any).Return(typeProps, nil)
			l := LinuxExecutor{m}
			actualSlice, err := l.ListUnits()
			require.NoError(t, err)
			require.Equal(t, 1, len(actualSlice))
			actual := actualSlice[0]
			assert.Nil(t, actual.ServiceProperties)
			assert.Nil(t, actual.SocketProperties)
			assert.Nil(t, actual.DeviceProperties)
			assert.Nil(t, actual.MountProperties)
			assert.Nil(t, actual.AutomountProperties)
			assert.Nil(t, actual.SwapProperties)
			assert.Nil(t, actual.PathProperties)
			assert.Nil(t, actual.TimerProperties)
			assert.Nil(t, actual.SliceProperties)
			assert.Nil(t, actual.ScopeProperties)
		})
		t.Run("path", func(t *testing.T) {
			typeName := "path"
			typeProps := map[string]interface{}{}
			m := &DbusMock{}
			m.On("ListUnits").Return([]dbus.UnitStatus{dbus.UnitStatus{Name: ("name." + typeName)}}, nil)
			m.On("GetUnitProperties", any, any).Return(map[string]interface{}{}, nil)
			m.On("GetUnitTypeProperties", any, any).Return(typeProps, nil)
			l := LinuxExecutor{m}
			actualSlice, err := l.ListUnits()
			require.NoError(t, err)
			require.Equal(t, 1, len(actualSlice))
			actual := actualSlice[0]
			assert.Nil(t, actual.ServiceProperties)
			assert.Nil(t, actual.SocketProperties)
			assert.Nil(t, actual.DeviceProperties)
			assert.Nil(t, actual.MountProperties)
			assert.Nil(t, actual.AutomountProperties)
			assert.Nil(t, actual.SwapProperties)
			assert.NotNil(t, actual.PathProperties)
			assert.Nil(t, actual.TimerProperties)
			assert.Nil(t, actual.SliceProperties)
			assert.Nil(t, actual.ScopeProperties)
		})
		t.Run("timer", func(t *testing.T) {
			typeName := "timer"
			typeProps := map[string]interface{}{}
			m := &DbusMock{}
			m.On("ListUnits").Return([]dbus.UnitStatus{dbus.UnitStatus{Name: ("name." + typeName)}}, nil)
			m.On("GetUnitProperties", any, any).Return(map[string]interface{}{}, nil)
			m.On("GetUnitTypeProperties", any, any).Return(typeProps, nil)
			l := LinuxExecutor{m}
			actualSlice, err := l.ListUnits()
			require.NoError(t, err)
			require.Equal(t, 1, len(actualSlice))
			actual := actualSlice[0]
			assert.Nil(t, actual.ServiceProperties)
			assert.Nil(t, actual.SocketProperties)
			assert.Nil(t, actual.DeviceProperties)
			assert.Nil(t, actual.MountProperties)
			assert.Nil(t, actual.AutomountProperties)
			assert.Nil(t, actual.SwapProperties)
			assert.Nil(t, actual.PathProperties)
			assert.NotNil(t, actual.TimerProperties)
			assert.Nil(t, actual.SliceProperties)
			assert.Nil(t, actual.ScopeProperties)
		})
		t.Run("snapshot", func(t *testing.T) {
			typeName := "snapshot"
			typeProps := map[string]interface{}{}
			m := &DbusMock{}
			m.On("ListUnits").Return([]dbus.UnitStatus{dbus.UnitStatus{Name: ("name." + typeName)}}, nil)
			m.On("GetUnitProperties", any, any).Return(map[string]interface{}{}, nil)
			m.On("GetUnitTypeProperties", any, any).Return(typeProps, nil)
			l := LinuxExecutor{m}
			actualSlice, err := l.ListUnits()
			require.NoError(t, err)
			require.Equal(t, 1, len(actualSlice))
			actual := actualSlice[0]
			assert.Nil(t, actual.ServiceProperties)
			assert.Nil(t, actual.SocketProperties)
			assert.Nil(t, actual.DeviceProperties)
			assert.Nil(t, actual.MountProperties)
			assert.Nil(t, actual.AutomountProperties)
			assert.Nil(t, actual.SwapProperties)
			assert.Nil(t, actual.PathProperties)
			assert.Nil(t, actual.TimerProperties)
			assert.Nil(t, actual.SliceProperties)
			assert.Nil(t, actual.ScopeProperties)
		})
		t.Run("slice", func(t *testing.T) {
			typeName := "slice"
			typeProps := map[string]interface{}{}
			m := &DbusMock{}
			m.On("ListUnits").Return([]dbus.UnitStatus{dbus.UnitStatus{Name: ("name." + typeName)}}, nil)
			m.On("GetUnitProperties", any, any).Return(map[string]interface{}{}, nil)
			m.On("GetUnitTypeProperties", any, any).Return(typeProps, nil)
			l := LinuxExecutor{m}
			actualSlice, err := l.ListUnits()
			require.NoError(t, err)
			require.Equal(t, 1, len(actualSlice))
			actual := actualSlice[0]
			assert.Nil(t, actual.ServiceProperties)
			assert.Nil(t, actual.SocketProperties)
			assert.Nil(t, actual.DeviceProperties)
			assert.Nil(t, actual.MountProperties)
			assert.Nil(t, actual.AutomountProperties)
			assert.Nil(t, actual.SwapProperties)
			assert.Nil(t, actual.PathProperties)
			assert.Nil(t, actual.TimerProperties)
			assert.NotNil(t, actual.SliceProperties)
			assert.Nil(t, actual.ScopeProperties)
		})
		t.Run("scope", func(t *testing.T) {
			typeName := "scope"
			typeProps := map[string]interface{}{}
			m := &DbusMock{}
			m.On("ListUnits").Return([]dbus.UnitStatus{dbus.UnitStatus{Name: ("name." + typeName)}}, nil)
			m.On("GetUnitProperties", any, any).Return(map[string]interface{}{}, nil)
			m.On("GetUnitTypeProperties", any, any).Return(typeProps, nil)
			l := LinuxExecutor{m}
			actualSlice, err := l.ListUnits()
			require.NoError(t, err)
			require.Equal(t, 1, len(actualSlice))
			actual := actualSlice[0]
			assert.Nil(t, actual.ServiceProperties)
			assert.Nil(t, actual.SocketProperties)
			assert.Nil(t, actual.DeviceProperties)
			assert.Nil(t, actual.MountProperties)
			assert.Nil(t, actual.AutomountProperties)
			assert.Nil(t, actual.SwapProperties)
			assert.Nil(t, actual.PathProperties)
			assert.Nil(t, actual.TimerProperties)
			assert.Nil(t, actual.SliceProperties)
			assert.NotNil(t, actual.ScopeProperties)
		})
	})
}

// TestQueryUnit runs a test
func TestQueryUnit(t *testing.T) {
	t.Parallel()
	t.Run("when-verify", func(t *testing.T) {
		t.Parallel()
		t.Run("when-unit-exists", func(t *testing.T) {
			t.Parallel()
			unit := randomUnitStatus()
			m := &DbusMock{}
			m.On("ListUnits").Return([]dbus.UnitStatus{unit}, nil)
			m.On("GetUnitProperties", any, any).Return(map[string]interface{}{}, nil)
			m.On("GetUnitTypeProperties", any, any).Return(map[string]interface{}{}, nil)
			l := LinuxExecutor{m}
			actual, err := l.QueryUnit(unit.Name, true)
			assert.NoError(t, err)
			assertUnitStatusEqUnit(t, unit, actual)
		})
		t.Run("when-unit-not-exists", func(t *testing.T) {
			t.Parallel()
			unit := randomUnitStatus()
			m := &DbusMock{}
			m.On("ListUnits").Return([]dbus.UnitStatus{randomUnitStatus()}, nil)
			m.On("GetUnitProperties", any, any).Return(map[string]interface{}{}, nil)
			m.On("GetUnitTypeProperties", any, any).Return(map[string]interface{}{}, nil)
			l := LinuxExecutor{m}
			_, err := l.QueryUnit(unit.Name, true)
			assert.Error(t, err)
		})
		t.Run("when-list-units-error", func(t *testing.T) {
			t.Parallel()
			err := errors.New("error1")
			m := &DbusMock{}
			m.On("ListUnits").Return([]dbus.UnitStatus{}, err)
			l := LinuxExecutor{m}
			_, actual := l.QueryUnit("name1", true)
			assert.Equal(t, err, errors.Cause(actual))
		})
	})

	t.Run("when-not-verify", func(t *testing.T) {
		t.Run("when-list-units-returns-value", func(t *testing.T) {
			t.Parallel()
			unit := randomUnitStatus()
			m := &DbusMock{}
			m.On("ListUnitsByNames", any).Return([]dbus.UnitStatus{unit}, nil)
			m.On("GetUnitProperties", any, any).Return(map[string]interface{}{}, nil)
			m.On("GetUnitTypeProperties", any, any).Return(map[string]interface{}{}, nil)
			l := LinuxExecutor{m}
			actual, err := l.QueryUnit(unit.Name, false)
			assert.NoError(t, err)
			assertUnitStatusEqUnit(t, unit, actual)
		})
		t.Run("when-list-units-returns-error", func(t *testing.T) {
			t.Parallel()
			expected := errors.New("error1")
			m := &DbusMock{}
			m.On("ListUnitsByNames", any).Return([]dbus.UnitStatus{}, expected)
			l := LinuxExecutor{m}
			_, actual := l.QueryUnit("name1", false)
			assert.Equal(t, expected, errors.Cause(actual))
		})
	})
}

// TestStartUnit runs a test
func TestStartUnit(t *testing.T) {
	t.Parallel()
	t.Run("call-start-unit", func(t *testing.T) {
		t.Parallel()
		u := randomUnit(UnitTypeService)
		m := &DbusMock{startResp: "done"}
		m.On("StartUnit", any, any, any).Return(1, nil)
		l := LinuxExecutor{m}
		err := l.StartUnit(u)
		assert.NoError(t, err)
		m.AssertCalled(t, "StartUnit", any, any, any)
	})
	t.Run("test-channel-return-values", func(t *testing.T) {
		t.Parallel()
		t.Run("done", func(t *testing.T) {
			t.Parallel()
			u := randomUnit(UnitTypeService)
			m := &DbusMock{startResp: "done"}
			m.On("StartUnit", any, any, any).Return(1, nil)
			l := LinuxExecutor{m}
			err := l.StartUnit(u)
			assert.NoError(t, err)
		})
		t.Run("canceled", func(t *testing.T) {
			t.Parallel()
			u := randomUnit(UnitTypeService)
			m := &DbusMock{startResp: "canceled"}
			m.On("StartUnit", any, any, any).Return(1, nil)
			l := LinuxExecutor{m}
			err := l.StartUnit(u)
			assert.Equal(t, err, fmt.Errorf("operation was cancelled while starting: %s", u.Name))
		})
		t.Run("timeout", func(t *testing.T) {
			t.Parallel()
			u := randomUnit(UnitTypeService)
			m := &DbusMock{startResp: "timeout"}
			m.On("StartUnit", any, any, any).Return(1, nil)
			l := LinuxExecutor{m}
			err := l.StartUnit(u)
			assert.Equal(t, err, fmt.Errorf("operation timed out while starting: %s", u.Name))
		})
		t.Run("failed", func(t *testing.T) {
			t.Parallel()
			u := randomUnit(UnitTypeService)
			m := &DbusMock{startResp: "failed"}
			m.On("StartUnit", any, any, any).Return(1, nil)
			l := LinuxExecutor{m}
			err := l.StartUnit(u)
			assert.Equal(t, err, fmt.Errorf("operation failed while starting: %s", u.Name))
		})
		t.Run("dependency", func(t *testing.T) {
			t.Parallel()
			u := randomUnit(UnitTypeService)
			m := &DbusMock{startResp: "dependency"}
			m.On("StartUnit", any, any, any).Return(1, nil)
			l := LinuxExecutor{m}
			err := l.StartUnit(u)
			assert.Equal(t, err, fmt.Errorf("operation depends on a failed unit when starting: %s", u.Name))
		})
		t.Run("skipped", func(t *testing.T) {
			t.Parallel()
			u := randomUnit(UnitTypeService)
			m := &DbusMock{startResp: "skipped"}
			m.On("StartUnit", any, any, any).Return(1, nil)
			l := LinuxExecutor{m}
			err := l.StartUnit(u)
			assert.NoError(t, err)
		})
		t.Run("bad-message", func(t *testing.T) {
			t.Parallel()
			u := randomUnit(UnitTypeService)
			m := &DbusMock{startResp: "msg1"}
			m.On("StartUnit", any, any, any).Return(1, nil)
			l := LinuxExecutor{m}
			err := l.StartUnit(u)
			assert.Equal(t, err, fmt.Errorf("unknown systemd status: msg1"))
		})
	})
	t.Run("start-unit-returns-error", func(t *testing.T) {
		t.Parallel()
		expected := errors.New("err1")
		u := randomUnit(UnitTypeService)
		m := &DbusMock{}
		m.On("StartUnit", any, any, any).Return(1, expected)
		l := LinuxExecutor{m}
		actual := l.StartUnit(u)
		assert.Equal(t, expected, actual)
	})
}

// TestStopUnit runs a test
func TestStopUnit(t *testing.T) {
	t.Parallel()
	t.Run("call-stop-unit", func(t *testing.T) {
		t.Parallel()
		u := randomUnit(UnitTypeService)
		m := &DbusMock{stopResp: "done"}
		m.On("StopUnit", any, any, any).Return(1, nil)
		l := LinuxExecutor{m}
		err := l.StopUnit(u)
		assert.NoError(t, err)
		m.AssertCalled(t, "StopUnit", any, any, any)
	})
	t.Run("test-channel-return-values", func(t *testing.T) {
		t.Parallel()
		t.Run("done", func(t *testing.T) {
			t.Parallel()
			u := randomUnit(UnitTypeService)
			m := &DbusMock{stopResp: "done"}
			m.On("StopUnit", any, any, any).Return(1, nil)
			l := LinuxExecutor{m}
			err := l.StopUnit(u)
			assert.NoError(t, err)
		})
		t.Run("canceled", func(t *testing.T) {
			t.Parallel()
			u := randomUnit(UnitTypeService)
			m := &DbusMock{stopResp: "canceled"}
			m.On("StopUnit", any, any, any).Return(1, nil)
			l := LinuxExecutor{m}
			err := l.StopUnit(u)
			assert.Equal(t, err, fmt.Errorf("operation was cancelled while stopping: %s", u.Name))
		})
		t.Run("timeout", func(t *testing.T) {
			t.Parallel()
			u := randomUnit(UnitTypeService)
			m := &DbusMock{stopResp: "timeout"}
			m.On("StopUnit", any, any, any).Return(1, nil)
			l := LinuxExecutor{m}
			err := l.StopUnit(u)
			assert.Equal(t, err, fmt.Errorf("operation timed out while stopping: %s", u.Name))
		})
		t.Run("failed", func(t *testing.T) {
			t.Parallel()
			u := randomUnit(UnitTypeService)
			m := &DbusMock{stopResp: "failed"}
			m.On("StopUnit", any, any, any).Return(1, nil)
			l := LinuxExecutor{m}
			err := l.StopUnit(u)
			assert.Equal(t, err, fmt.Errorf("operation failed while stopping: %s", u.Name))
		})
		t.Run("dependency", func(t *testing.T) {
			t.Parallel()
			u := randomUnit(UnitTypeService)
			m := &DbusMock{stopResp: "dependency"}
			m.On("StopUnit", any, any, any).Return(1, nil)
			l := LinuxExecutor{m}
			err := l.StopUnit(u)
			assert.Equal(t, err, fmt.Errorf("operation depends on a failed unit when stopping: %s", u.Name))
		})
		t.Run("skipped", func(t *testing.T) {
			t.Parallel()
			u := randomUnit(UnitTypeService)
			m := &DbusMock{stopResp: "skipped"}
			m.On("StopUnit", any, any, any).Return(1, nil)
			l := LinuxExecutor{m}
			err := l.StopUnit(u)
			assert.NoError(t, err)
		})
		t.Run("bad-message", func(t *testing.T) {
			t.Parallel()
			u := randomUnit(UnitTypeService)
			m := &DbusMock{stopResp: "msg1"}
			m.On("StopUnit", any, any, any).Return(1, nil)
			l := LinuxExecutor{m}
			err := l.StopUnit(u)
			assert.Equal(t, err, fmt.Errorf("unknown systemd status: msg1"))
		})
	})
	t.Run("stop-unit-returns-error", func(t *testing.T) {
		t.Parallel()
		expected := errors.New("err1")
		u := randomUnit(UnitTypeService)
		m := &DbusMock{}
		m.On("StopUnit", any, any, any).Return(1, expected)
		l := LinuxExecutor{m}
		actual := l.StopUnit(u)
		assert.Equal(t, expected, actual)
	})
}

// Test RestartUnit runs a test
func TestRestartUnit(t *testing.T) {
	t.Parallel()
	t.Run("call-restart-unit", func(t *testing.T) {
		t.Parallel()
		u := randomUnit(UnitTypeService)
		m := &DbusMock{restartResp: "done"}
		m.On("RestartUnit", any, any, any).Return(1, nil)
		l := LinuxExecutor{m}
		err := l.RestartUnit(u)
		assert.NoError(t, err)
		m.AssertCalled(t, "RestartUnit", any, any, any)
	})
	t.Run("test-channel-return-values", func(t *testing.T) {
		t.Parallel()
		t.Run("done", func(t *testing.T) {
			t.Parallel()
			u := randomUnit(UnitTypeService)
			m := &DbusMock{restartResp: "done"}
			m.On("RestartUnit", any, any, any).Return(1, nil)
			l := LinuxExecutor{m}
			err := l.RestartUnit(u)
			assert.NoError(t, err)
		})
		t.Run("canceled", func(t *testing.T) {
			t.Parallel()
			u := randomUnit(UnitTypeService)
			m := &DbusMock{restartResp: "canceled"}
			m.On("RestartUnit", any, any, any).Return(1, nil)
			l := LinuxExecutor{m}
			err := l.RestartUnit(u)
			assert.Equal(t, err, fmt.Errorf("operation was cancelled while restarting: %s", u.Name))
		})
		t.Run("timeout", func(t *testing.T) {
			t.Parallel()
			u := randomUnit(UnitTypeService)
			m := &DbusMock{restartResp: "timeout"}
			m.On("RestartUnit", any, any, any).Return(1, nil)
			l := LinuxExecutor{m}
			err := l.RestartUnit(u)
			assert.Equal(t, err, fmt.Errorf("operation timed out while restarting: %s", u.Name))
		})
		t.Run("failed", func(t *testing.T) {
			t.Parallel()
			u := randomUnit(UnitTypeService)
			m := &DbusMock{restartResp: "failed"}
			m.On("RestartUnit", any, any, any).Return(1, nil)
			l := LinuxExecutor{m}
			err := l.RestartUnit(u)
			assert.Equal(t, err, fmt.Errorf("operation failed while restarting: %s", u.Name))
		})
		t.Run("dependency", func(t *testing.T) {
			t.Parallel()
			u := randomUnit(UnitTypeService)
			m := &DbusMock{restartResp: "dependency"}
			m.On("RestartUnit", any, any, any).Return(1, nil)
			l := LinuxExecutor{m}
			err := l.RestartUnit(u)
			assert.Equal(t, err, fmt.Errorf("operation depends on a failed unit when restarting: %s", u.Name))
		})
		t.Run("skipped", func(t *testing.T) {
			t.Parallel()
			u := randomUnit(UnitTypeService)
			m := &DbusMock{restartResp: "skipped"}
			m.On("RestartUnit", any, any, any).Return(1, nil)
			l := LinuxExecutor{m}
			err := l.RestartUnit(u)
			assert.NoError(t, err)
		})
		t.Run("bad-message", func(t *testing.T) {
			t.Parallel()
			u := randomUnit(UnitTypeService)
			m := &DbusMock{restartResp: "msg1"}
			m.On("RestartUnit", any, any, any).Return(1, nil)
			l := LinuxExecutor{m}
			err := l.RestartUnit(u)
			assert.Equal(t, err, fmt.Errorf("unknown systemd status: msg1"))
		})
	})
	t.Run("restart-unit-returns-error", func(t *testing.T) {
		t.Parallel()
		expected := errors.New("err1")
		u := randomUnit(UnitTypeService)
		m := &DbusMock{}
		m.On("RestartUnit", any, any, any).Return(1, expected)
		l := LinuxExecutor{m}
		actual := l.RestartUnit(u)
		assert.Equal(t, expected, actual)
	})
}

// Test ReloadUnit runs a test
func TestReloadUnit(t *testing.T) {
	t.Parallel()
	t.Run("call-reload-unit", func(t *testing.T) {
		t.Parallel()
		u := randomUnit(UnitTypeService)
		m := &DbusMock{reloadResp: "done"}
		m.On("ReloadUnit", any, any, any).Return(1, nil)
		l := LinuxExecutor{m}
		err := l.ReloadUnit(u)
		assert.NoError(t, err)
		m.AssertCalled(t, "ReloadUnit", any, any, any)
	})
	t.Run("test-channel-return-values", func(t *testing.T) {
		t.Parallel()
		t.Run("done", func(t *testing.T) {
			t.Parallel()
			u := randomUnit(UnitTypeService)
			m := &DbusMock{reloadResp: "done"}
			m.On("ReloadUnit", any, any, any).Return(1, nil)
			l := LinuxExecutor{m}
			err := l.ReloadUnit(u)
			assert.NoError(t, err)
		})
		t.Run("canceled", func(t *testing.T) {
			t.Parallel()
			u := randomUnit(UnitTypeService)
			m := &DbusMock{reloadResp: "canceled"}
			m.On("ReloadUnit", any, any, any).Return(1, nil)
			l := LinuxExecutor{m}
			err := l.ReloadUnit(u)
			assert.Equal(t, err, fmt.Errorf("operation was cancelled while reloading: %s", u.Name))
		})
		t.Run("timeout", func(t *testing.T) {
			t.Parallel()
			u := randomUnit(UnitTypeService)
			m := &DbusMock{reloadResp: "timeout"}
			m.On("ReloadUnit", any, any, any).Return(1, nil)
			l := LinuxExecutor{m}
			err := l.ReloadUnit(u)
			assert.Equal(t, err, fmt.Errorf("operation timed out while reloading: %s", u.Name))
		})
		t.Run("failed", func(t *testing.T) {
			t.Parallel()
			u := randomUnit(UnitTypeService)
			m := &DbusMock{reloadResp: "failed"}
			m.On("ReloadUnit", any, any, any).Return(1, nil)
			l := LinuxExecutor{m}
			err := l.ReloadUnit(u)
			assert.Equal(t, err, fmt.Errorf("operation failed while reloading: %s", u.Name))
		})
		t.Run("dependency", func(t *testing.T) {
			t.Parallel()
			u := randomUnit(UnitTypeService)
			m := &DbusMock{reloadResp: "dependency"}
			m.On("ReloadUnit", any, any, any).Return(1, nil)
			l := LinuxExecutor{m}
			err := l.ReloadUnit(u)
			assert.Equal(t, err, fmt.Errorf("operation depends on a failed unit when reloading: %s", u.Name))
		})
		t.Run("skipped", func(t *testing.T) {
			t.Parallel()
			u := randomUnit(UnitTypeService)
			m := &DbusMock{reloadResp: "skipped"}
			m.On("ReloadUnit", any, any, any).Return(1, nil)
			l := LinuxExecutor{m}
			err := l.ReloadUnit(u)
			assert.NoError(t, err)
		})
		t.Run("bad-message", func(t *testing.T) {
			t.Parallel()
			u := randomUnit(UnitTypeService)
			m := &DbusMock{reloadResp: "msg1"}
			m.On("ReloadUnit", any, any, any).Return(1, nil)
			l := LinuxExecutor{m}
			err := l.ReloadUnit(u)
			assert.Equal(t, err, fmt.Errorf("unknown systemd status: msg1"))
		})
	})
	t.Run("reload-unit-returns-error", func(t *testing.T) {
		t.Parallel()
		expected := errors.New("err1")
		u := randomUnit(UnitTypeService)
		m := &DbusMock{}
		m.On("ReloadUnit", any, any, any).Return(1, expected)
		l := LinuxExecutor{m}
		actual := l.ReloadUnit(u)
		assert.Equal(t, expected, actual)
	})
}

func TestSendSignal(t *testing.T) {
	t.Parallel()
	u := randomUnit(UnitTypeService)
	m := &DbusMock{}
	m.On("KillUnit", any, any).Return()
	l := LinuxExecutor{m}
	signals := []Signal{SIGHUP, SIGINT, SIGQUIT, SIGILL, SIGTRAP, SIGABRT,
		SIGEMT, SIGFPE, SIGKILL, SIGBUS, SIGSEGV, SIGSYS, SIGPIPE, SIGALRM, SIGTERM,
		SIGURG, SIGSTOP, SIGTSTP, SIGCONT, SIGCHLD, SIGTTIN, SIGTTOU, SIGIO,
		SIGXCPU, SIGXFSZ, SIGVTALRM, SIGPROF, SIGWINCH, SIGINFO, SIGUSR1, SIGUSR2}
	for _, signal := range signals {
		t.Run(signal.String(), func(t *testing.T) {
			t.Parallel()
			l.SendSignal(u, signal)
			m.AssertCalled(t, "KillUnit", u.Name, int32(signal))
		})
	}
}
