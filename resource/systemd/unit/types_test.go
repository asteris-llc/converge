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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var typeMap = map[UnitType]string{
	UnitTypeService:   "service",
	UnitTypeSocket:    "socket",
	UnitTypeDevice:    "device",
	UnitTypeMount:     "mount",
	UnitTypeAutoMount: "automount",
	UnitTypeSwap:      "swap",
	UnitTypeTarget:    "target",
	UnitTypePath:      "path",
	UnitTypeTimer:     "timer",
	UnitTypeSnapshot:  "snapshot",
	UnitTypeSlice:     "slice",
	UnitTypeScope:     "scope",
	UnitTypeUnknown:   "",
}

// TestUnitTypeZeroValue ensures that the zero value is UnitTypeUnknown
func TestUnitTypeZeroValue(t *testing.T) {
	var u UnitType
	assert.Equal(t, UnitTypeUnknown, u)
}

// Test conversion from a unit file name or service name
func TestUnitTypeFromName(t *testing.T) {
	t.Parallel()
	t.Run("when-name", func(t *testing.T) {
		t.Parallel()
		for k, v := range typeMap {
			assert.Equal(t, k, UnitTypeFromName(v))
		}
	})
	t.Run("when-dot-name", func(t *testing.T) {
		t.Parallel()
		for k, v := range typeMap {
			name := "." + v
			assert.Equal(t, k, UnitTypeFromName(name))
		}
	})
	t.Run("when-path", func(t *testing.T) {
		t.Parallel()
		for k, v := range typeMap {
			name := "/foo/bar/baz/name." + v
			assert.Equal(t, k, UnitTypeFromName(name))
		}
	})
	t.Run("when-dashes", func(t *testing.T) {
		t.Parallel()
		s := "sys-devices-platform-serial8250-tty-ttyS16.device"
		assert.Equal(t, UnitTypeDevice, UnitTypeFromName(s))
	})
	t.Run("when-upper-case", func(t *testing.T) {
		t.Parallel()
		for k, v := range typeMap {
			assert.Equal(t, k, UnitTypeFromName(strings.ToUpper(v)))
		}
	})
	t.Run("when-error", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, UnitTypeUnknown, UnitTypeFromName("unknown-service"))
		assert.Equal(t, UnitTypeUnknown, UnitTypeFromName(""))
		assert.Equal(t, UnitTypeUnknown, UnitTypeFromName("service1"))
		assert.Equal(t, UnitTypeUnknown, UnitTypeFromName("service.service1"))
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
	typeMap := map[UnitType]string{
		UnitTypeService:   "Service",
		UnitTypeSocket:    "Socket",
		UnitTypeDevice:    "Device",
		UnitTypeMount:     "Mount",
		UnitTypeAutoMount: "Automount",
		UnitTypeSwap:      "Swap",
		UnitTypeTarget:    "Target",
		UnitTypePath:      "Path",
		UnitTypeTimer:     "Timer",
		UnitTypeSnapshot:  "Snapshot",
		UnitTypeSlice:     "Slice",
		UnitTypeScope:     "Scope",
		UnitTypeUnknown:   "",
	}

	for k, v := range typeMap {
		assert.Equal(t, v, k.UnitTypeString())
	}
}

func TestUnitHasProperties(t *testing.T) {
	typeMap := map[UnitType]bool{
		UnitTypeTarget:    false,
		UnitTypeSnapshot:  false,
		UnitTypeUnknown:   false,
		UnitTypeService:   true,
		UnitTypeSocket:    true,
		UnitTypeDevice:    true,
		UnitTypeMount:     true,
		UnitTypeAutoMount: true,
		UnitTypeSwap:      true,
		UnitTypePath:      true,
		UnitTypeTimer:     true,
		UnitTypeSlice:     true,
		UnitTypeScope:     true,
	}

	for k, v := range typeMap {
		assert.Equal(t, v, k.HasProperties())
	}
}
