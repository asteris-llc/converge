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

package systemd

import (
	"fmt"

	"github.com/asteris-llc/converge/resource"
	"github.com/coreos/go-systemd/dbus"
)

// CheckProperty checks if the value of a unit matches one of the provided properties
func CheckProperty(conn *dbus.Conn, unit string, propertyName string, wants []*dbus.Property) (*resource.Status, error) {
	prop, err := conn.GetUnitProperty(unit, propertyName)
	if err != nil {
		return nil, err
	}

	if len(wants) == 0 {
		errMsg := fmt.Errorf("property %q of unit %q has no expected states", propertyName, unit)
		return &resource.Status{
			Level: resource.StatusFatal,
			Output: []string{
				errMsg.Error(),
			},
		}, errMsg
	}
	possibilities, rest := wants[0].Value.Value(), wants[1:]
	found := prop.Value.Value() == possibilities
	for i := range rest {
		if prop.Value.Value() == rest[i].Value.Value() {
			found = true
		}
		possibilities = fmt.Sprintf("%s, %s", possibilities, rest[i].Value.Value())
	}
	// Create property diffs
	propDiff := PropertyDiff{
		Actual:   prop,
		Expected: wants,
	}
	var diffs map[string]resource.Diff
	warningLevel := resource.StatusNoChange
	if !found {
		warningLevel = resource.StatusWillChange
		diffs = map[string]resource.Diff{
			//unit:propertyname:shouldbe="1,2,3"
			fmt.Sprintf("%s:%s:shouldbe=%q", unit, propertyName, possibilities): &propDiff,
		}
	}
	statusMsg := fmt.Sprintf("property %q of unit %q is %q, expected one of [%q]", propertyName, unit, prop.Value.Value(), possibilities)
	return &resource.Status{
		Level:       warningLevel,
		Differences: diffs,
		Output:      []string{statusMsg},
	}, nil
}

// PropertyDiff shows the difference between a given property and the expected
type PropertyDiff struct {
	Actual   *dbus.Property
	Expected []*dbus.Property
}

// Original shows the origial property
func (diff *PropertyDiff) Original() string {
	return fmt.Sprintf("property %q is %q", diff.Actual.Name, diff.Actual.Value.Value())
}

// Current shows what the property should be
func (diff *PropertyDiff) Current() string {
	possibilities, rest := diff.Expected[0].Value.Value(), diff.Expected[1:]
	for i := range rest {
		possibilities = fmt.Sprintf("%s, %s", possibilities, rest[i].Value.Value())
	}
	return fmt.Sprintf("property %q should be one of [%q]", diff.Expected[0].Name, possibilities)
}

// Changes returns true if the expected file mode differs from the current mode
func (diff *PropertyDiff) Changes() bool {
	for i := range diff.Expected {
		if diff.Actual.Value.Value() == diff.Expected[i].Value.Value() {
			return false
		}
	}
	return true
}
