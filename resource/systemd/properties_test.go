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
// limitations under the License.package systemd

package systemd_test

import (
	"testing"

	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/systemd"
	"github.com/coreos/go-systemd/dbus"
	"github.com/stretchr/testify/assert"
)

// TestPropertyDiffInterface ensures that the PropertyDiff implemnts the
// Diff interface
func TestPropertyDiffInterface(t *testing.T) {
	t.Parallel()

	assert.Implements(t, (*resource.Diff)(nil), new(systemd.PropertyDiff))
}

// TestCheckPropertyFunc ensures that CheckPropertyFunc returns a NoChange Status
// when property values are what is expected.
func TestCheckPropertyFunc(t *testing.T) {
	t.Parallel()
	if !HasSystemd() {
		t.Skipf("System does not have systemd. Skipping.")
	}
	conn, err := systemd.GetDbusConnection()
	assert.NoError(t, err)
	defer conn.Return()
	dbusConn := conn.Connection
	units, err := dbusConn.ListUnits()
	assert.NoError(t, err)
	if len(units) == 0 {
		t.Skipf("System does not have any unit files. Skipping.")
	}
	unit := units[0]

	// Ensure CheckProperty can read active states of units.
	status, err := systemd.CheckProperty(
		dbusConn,
		unit.Name,
		"ActiveState",
		[]*dbus.Property{systemd.PropActiveState(systemd.ActiveState(unit.ActiveState))})

	assert.NoError(t, err)
	assert.Equal(t, resource.StatusNoChange, status.StatusCode())
	assert.Nil(t, status.Diffs())

	status, err = systemd.CheckProperty(
		dbusConn,
		unit.Name,
		"LoadState",
		[]*dbus.Property{systemd.PropLoadState(systemd.LoadState(unit.LoadState))})

	assert.NoError(t, err)
	assert.Equal(t, resource.StatusNoChange, status.StatusCode())
	assert.Nil(t, status.Diffs())

}

// HasSystemd checks if dbus is available
func HasSystemd() bool {
	conn, err := systemd.GetDbusConnection()
	if err != nil {
		return false
	}
	defer conn.Return()
	return true
}
