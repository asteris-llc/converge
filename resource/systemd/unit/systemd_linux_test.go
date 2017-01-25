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
	"errors"
	"reflect"
	"testing"

	"github.com/coreos/go-systemd/dbus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInterface(t *testing.T) {
	assert.Implements(t, (*SystemdExecutor)(nil), new(LinuxExecutor))
}

// TestListUnits runs a test
func TestListUnits(t *testing.T) {
	t.Parallel()

	t.Run("list-units-returns-error", func(t *testing.T) {
		t.Parallel()
		m := &DbusMock{}
		expected := errors.New("error1")
		m.On("ListUnits").Return([]dbus.UnitStatus{}, expected)
		l := LinuxExecutor{m}
		_, err := l.ListUnits()
		assert.Equal(t, expected, err)
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

		expected := randomUnit()
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

		expected := &Properties{
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
