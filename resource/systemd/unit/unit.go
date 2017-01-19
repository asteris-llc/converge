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
	"fmt"

	"github.com/coreos/go-systemd/dbus"
)

type Unit struct {
	Name        string
	Description string
	ActiveState string
	Path        string
	Type        UnitType
}

func (u *Unit) IsServiceUnit() bool {
	return UnitTypeService == UnitTypeFromName(u.Path)
}

func newFromStatus(status *dbus.UnitStatus, opts map[string]interface{}) *Unit {
	var path string
	if fragment, ok := opts["FragmentPath"]; ok {
		path = fragment.(string)
	}
	return &Unit{
		Name:        status.Name,
		Description: status.Description,
		ActiveState: status.ActiveState,
		Path:        path,
	}
}

func PPUnit(u *Unit) string {
	fmtStr := `
Unit
---------------
Name:        %s
Description: %s
ActiveState: %s
Path:        %s
---------------
`
	return fmt.Sprintf(fmtStr, u.Name, u.Description, u.ActiveState, u.Path)
}

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
