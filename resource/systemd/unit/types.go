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
	"path"
	"strings"
)

// UnitType provides an enumeration over the different types of systemd unit
// file types
type UnitType uint

const (
	// UnitTypeService represents a systemd.service(5)
	UnitTypeService UnitType = iota

	// UnitTypeSocket represents a systemd.socket(5)
	UnitTypeSocket UnitType = iota

	// UnitTypeDevice represents a systemd.device(5)
	UnitTypeDevice UnitType = iota

	// UnitTypeMount represents a systemd.mount(5)
	UnitTypeMount UnitType = iota

	// UnitTypeAutoMount represents a systemd.automount(5)
	UnitTypeAutoMount UnitType = iota

	// UnitTypeSwap represents a systemd.swap(5)
	UnitTypeSwap UnitType = iota

	// UnitTypeTarget represents a systemd.target(5)
	UnitTypeTarget UnitType = iota

	// UnitTypePath represents a systemd.path(5)
	UnitTypePath UnitType = iota

	// UnitTypeTimer represents a systemd.timer(5)
	UnitTypeTimer UnitType = iota

	// UnitTypeSnapshot represents a systemd.snapshot(5)
	UnitTypeSnapshot UnitType = iota

	// UnitTypeSlice represents a systemd.slice(5)
	UnitTypeSlice UnitType = iota

	// UnitTypeScope represents a systemd.scope(5)
	UnitTypeScope UnitType = iota

	// UnitTypeUnknown represents a generic or unknown unit type
	UnitTypeUnknown UnitType = iota
)

// UnitTypeFromName takes a service name in the form of "foo.service" and
// returns an appropriate UnitType representing that service type. If the
// service type isn't recognized from the suffix then UnitTypeUnknown is
// returned.
func UnitTypeFromName(s string) UnitType {
	basename := path.Base(s)
	components := strings.Split(basename, ".")
	if len(components) == 0 {
		return UnitTypeUnknown
	}
	switch strings.ToLower(components[len(components)-1]) {
	case "service":
		return UnitTypeService
	case "socket":
		return UnitTypeSocket
	case "device":
		return UnitTypeDevice
	case "mount":
		return UnitTypeMount
	case "automount":
		return UnitTypeAutoMount
	case "swap":
		return UnitTypeSwap
	case "target":
		return UnitTypeTarget
	case "path":
		return UnitTypePath
	case "timer":
		return UnitTypeTimer
	case "snapshot":
		return UnitTypeSnapshot
	case "slice":
		return UnitTypeSlice
	case "scope":
		return UnitTypeScope
	}
	return UnitTypeUnknown
}

// Suffix is the dual of UnitTypeFromName and generates the correct unit file
// suffix based on the type
func (u UnitType) Suffix() string {
	switch u {
	case UnitTypeService:
		return "service"
	case UnitTypeSocket:
		return "socket"
	case UnitTypeDevice:
		return "device"
	case UnitTypeMount:
		return "mount"
	case UnitTypeAutoMount:
		return "automount"
	case UnitTypeSwap:
		return "swap"
	case UnitTypeTarget:
		return "target"
	case UnitTypePath:
		return "path"
	case UnitTypeTimer:
		return "timer"
	case UnitTypeSnapshot:
		return "snapshot"
	case UnitTypeSlice:
		return "slice"
	case UnitTypeScope:
		return "scope"
	}
	return ""
}

func (u UnitType) UnitTypeString() string {
	return strings.Title(u.Suffix())
}
