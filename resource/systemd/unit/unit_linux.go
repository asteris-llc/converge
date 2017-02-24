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

	"github.com/coreos/go-systemd/dbus"
)

// PPtUnitStatus pretty-prints a UnitStatus
func PPtUnitStatus(u *dbus.UnitStatus) string {
	fmtStr := `
UnitStatus
---------------
Name:        %s
Description: %s
LoadState:   %s
ActiveState: %s
SubState:    %s
Followed:    %s
Path:        %v
JobID:       %d
JobType:     %s
JobPath:     %v
---------------
`
	return fmt.Sprintf(fmtStr,
		u.Name,
		u.Description,
		u.LoadState,
		u.ActiveState,
		u.SubState,
		u.Followed,
		u.Path,
		u.JobId,
		u.JobType,
		u.JobPath,
	)
}

func newFromStatus(status *dbus.UnitStatus) *Unit {
	return &Unit{
		Name:        status.Name,
		Description: status.Description,
		ActiveState: status.ActiveState,
		LoadState:   status.LoadState,
		Type:        UnitTypeFromName(status.Name),
	}
}

// SetProperties sets the global properties of a unit from a properties map
func (u *Unit) SetProperties(m map[string]interface{}) {
	u.Properties = *newPropertiesFromMap(m)
	u.Path = u.FragmentPath
}

// SetTypedProperties sets type specific properties of a unit from a map
func (u *Unit) SetTypedProperties(m map[string]interface{}) {
	switch u.Type {
	case UnitTypeService:
		u.ServiceProperties = newServiceTypePropertiesFromMap(m)
	case UnitTypeSocket:
		u.SocketProperties = newSocketTypePropertiesFromMap(m)
	case UnitTypeDevice:
		u.DeviceProperties = newDeviceTypePropertiesFromMap(m)
	case UnitTypeMount:
		u.MountProperties = newMountTypePropertiesFromMap(m)
	case UnitTypeAutoMount:
		u.AutomountProperties = newAutomountTypePropertiesFromMap(m)
	case UnitTypeSwap:
		u.SwapProperties = newSwapTypePropertiesFromMap(m)
	case UnitTypePath:
		u.PathProperties = newPathTypePropertiesFromMap(m)
	case UnitTypeTimer:
		u.TimerProperties = newTimerTypePropertiesFromMap(m)
	case UnitTypeSlice:
		u.SliceProperties = newSliceTypePropertiesFromMap(m)
	case UnitTypeScope:
		u.ScopeProperties = newScopeTypePropertiesFromMap(m)
	case UnitTypeTarget, UnitTypeSnapshot, UnitTypeUnknown:
		/* No type-specific properties for these unit types */
	}
}
