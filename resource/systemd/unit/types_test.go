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

package unit_test

import (
	"strings"
	"testing"

	"github.com/asteris-llc/converge/resource/systemd/unit"
	"github.com/stretchr/testify/assert"
)

var typeMap = map[unit.UnitType]string{
	unit.UnitTypeService:   "service",
	unit.UnitTypeSocket:    "socket",
	unit.UnitTypeDevice:    "device",
	unit.UnitTypeMount:     "mount",
	unit.UnitTypeAutoMount: "automount",
	unit.UnitTypeSwap:      "swap",
	unit.UnitTypeTarget:    "target",
	unit.UnitTypePath:      "path",
	unit.UnitTypeTimer:     "timer",
	unit.UnitTypeSnapshot:  "snapshot",
	unit.UnitTypeSlice:     "slice",
	unit.UnitTypeScope:     "scope",
	unit.UnitTypeUnknown:   "",
}

// TestUnitTypeZeroValue ensures that the zero value is unit.UnitTypeUnknown
func TestUnitTypeZeroValue(t *testing.T) {
	var u unit.UnitType
	assert.Equal(t, unit.UnitTypeUnknown, u)
}

// Test conversion from a unit file name or service name
func TestUnitTypeFromName(t *testing.T) {
	t.Parallel()
	t.Run("when-name", func(t *testing.T) {
		t.Parallel()
		for k, v := range typeMap {
			assert.Equal(t, k, unit.UnitTypeFromName(v))
		}
	})
	t.Run("when-dot-name", func(t *testing.T) {
		t.Parallel()
		for k, v := range typeMap {
			name := "." + v
			assert.Equal(t, k, unit.UnitTypeFromName(name))
		}
	})
	t.Run("when-path", func(t *testing.T) {
		t.Parallel()
		for k, v := range typeMap {
			name := "/foo/bar/baz/name." + v
			assert.Equal(t, k, unit.UnitTypeFromName(name))
		}
	})
	t.Run("when-dashes", func(t *testing.T) {
		t.Parallel()
		s := "sys-devices-platform-serial8250-tty-ttyS16.device"
		assert.Equal(t, unit.UnitTypeDevice, unit.UnitTypeFromName(s))
	})
	t.Run("when-upper-case", func(t *testing.T) {
		t.Parallel()
		for k, v := range typeMap {
			assert.Equal(t, k, unit.UnitTypeFromName(strings.ToUpper(v)))
		}
	})
	t.Run("when-error", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, unit.UnitTypeUnknown, unit.UnitTypeFromName("unknown-service"))
		assert.Equal(t, unit.UnitTypeUnknown, unit.UnitTypeFromName(""))
		assert.Equal(t, unit.UnitTypeUnknown, unit.UnitTypeFromName("service1"))
		assert.Equal(t, unit.UnitTypeUnknown, unit.UnitTypeFromName("service.service1"))
	})
}

func TestSuffix(t *testing.T) {
	t.Parallel()
	for u, s := range typeMap {
		assert.Equal(t, s, u.Suffix())
	}
}

func TestUnitTypeString(t *testing.T) {
	t.Parallel()
	typeMap := map[unit.UnitType]string{
		unit.UnitTypeService:   "Service",
		unit.UnitTypeSocket:    "Socket",
		unit.UnitTypeDevice:    "Device",
		unit.UnitTypeMount:     "Mount",
		unit.UnitTypeAutoMount: "Automount",
		unit.UnitTypeSwap:      "Swap",
		unit.UnitTypeTarget:    "Target",
		unit.UnitTypePath:      "Path",
		unit.UnitTypeTimer:     "Timer",
		unit.UnitTypeSnapshot:  "Snapshot",
		unit.UnitTypeSlice:     "Slice",
		unit.UnitTypeScope:     "Scope",
		unit.UnitTypeUnknown:   "",
	}

	for k, v := range typeMap {
		assert.Equal(t, v, k.UnitTypeString())
	}
}
