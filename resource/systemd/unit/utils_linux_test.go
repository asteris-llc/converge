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

package unit

import (
	"testing"

	"github.com/coreos/go-systemd/dbus"
	"github.com/stretchr/testify/assert"
)

func assertUnitStatusEqUnit(t *testing.T, status dbus.UnitStatus, unit *Unit) {
	assert.Equal(t, status.Name, unit.Name)
	assert.Equal(t, status.Description, unit.Description)
	assert.Equal(t, status.ActiveState, unit.ActiveState)
	assert.Equal(t, status.LoadState, unit.LoadState)
	assert.Equal(t, UnitTypeFromName(status.Name), unit.Type)
	assertCorrectTypeProperties(t, unit)
}

func assertCorrectTypeProperties(t *testing.T, u *Unit) {
	checks := map[string]func() bool{
		"service":   func() bool { return u.ServiceProperties == nil },
		"socket":    func() bool { return u.SocketProperties == nil },
		"device":    func() bool { return u.DeviceProperties == nil },
		"mount":     func() bool { return u.MountProperties == nil },
		"automount": func() bool { return u.AutomountProperties == nil },
		"swap":      func() bool { return u.SwapProperties == nil },
		"path":      func() bool { return u.PathProperties == nil },
		"timer":     func() bool { return u.TimerProperties == nil },
		"slice":     func() bool { return u.SliceProperties == nil },
		"scope":     func() bool { return u.ScopeProperties == nil },
	}
	switch u.Type {
	case UnitTypeService:
		checks["service"] = func() bool { return u.ServiceProperties != nil }
	case UnitTypeSocket:
		checks["socket"] = func() bool { return u.SocketProperties != nil }
	case UnitTypeDevice:
		checks["device"] = func() bool { return u.DeviceProperties != nil }
	case UnitTypeMount:
		checks["mount"] = func() bool { return u.MountProperties != nil }
	case UnitTypeAutoMount:
		checks["automount"] = func() bool { return u.AutomountProperties != nil }
	case UnitTypeSwap:
		checks["swap"] = func() bool { return u.SwapProperties != nil }
	case UnitTypePath:
		checks["path"] = func() bool { return u.PathProperties != nil }
	case UnitTypeTimer:
		checks["timer"] = func() bool { return u.TimerProperties != nil }
	case UnitTypeSlice:
		checks["slice"] = func() bool { return u.SliceProperties != nil }
	case UnitTypeScope:
		checks["scope"] = func() bool { return u.ScopeProperties != nil }
	}

	for k, v := range checks {
		t.Run("ensure type properties for "+k+" type", func(t *testing.T) {
			assert.True(t, v())
		})
	}
}
