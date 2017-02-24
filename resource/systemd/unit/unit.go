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

import "fmt"

// Unit represents a systemd unit
type Unit struct {
	// the global embedded properties of the unit
	Properties

	// the name of the unit
	Name string

	// the description of the unit
	Description string

	// the units current active state
	ActiveState string

	// the units current load state
	LoadState string

	// the path to the unit file, if any
	Path string

	// the type of unit file
	Type UnitType

	// ServiceProperties contain properties specific to Service unit types
	ServiceProperties *ServiceTypeProperties `export:"service_properties"`

	// SocketProperties contain properties specific to Socket unit types
	SocketProperties *SocketTypeProperties `export:"SocketProperties"`

	// DeviceProperties contain properties specific to Device unit types
	DeviceProperties *DeviceTypeProperties `export:"DeviceProperties"`

	// MountProperties contain properties specific to Mount unit types
	MountProperties *MountTypeProperties `export:"MountProperties"`

	// AutomountProperties contain properties specific for Autoumount unit types
	AutomountProperties *AutomountTypeProperties `export:"AutomountProperties"`

	// SwapProperties contain properties specific to Swap unit types
	SwapProperties *SwapTypeProperties `export:"SwapProperties"`

	// PathProperties contain properties specific to Path unit types
	PathProperties *PathTypeProperties `export:"PathProperties"`

	// TimerProperties contain properties specific to Timer unit types
	TimerProperties *TimerTypeProperties `export:"TimerProperties"`

	// SliceProperties contain properties specific to Slice unit types
	SliceProperties *SliceTypeProperties `export:"SliceProperties"`

	// ScopeProperties contain properties specific to Scope unit types
	ScopeProperties *ScopeTypeProperties `export:"ScopeProperties"`
}

// IsServiceUnit returns true if the unit is a service
func (u *Unit) IsServiceUnit() bool {
	return UnitTypeService == UnitTypeFromName(u.Path)
}

// PPUnit pretty-prints a unit
func PPUnit(u *Unit) string {
	fmtStr := `
Unit
=================
Name:        %s
Type:        %s
Description: %s
ActiveState: %s
Path:        %s
=================
`
	return fmt.Sprintf(fmtStr, u.Name, u.Type, u.Description, u.ActiveState, u.Path)
}
