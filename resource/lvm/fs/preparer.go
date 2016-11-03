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

package fs

import (
	"github.com/asteris-llc/converge/load/registry"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/lvm/lowlevel"

	"strings"
)

// Preparer for LVM FS Task
//
// Filesystem do formatting and mounting for LVM volumes
// (also capable to format usual block devices as well)
type Preparer struct {
	// Device path to be mount
	// Examples: `/dev/sda1`, `/dev/mapper/vg0-data`
	Device string `hcl:"device" required:"true"`

	// Mountpoint where device will be mounted
	// (should be an existing directory)
	// Example: /mnt/data
	Mountpoint string `hcl:"mount" required:"true"`

	// Fstype is filesystem type
	// (actually any linux filesystem, except `ZFS`)
	// Example:  `ext4`, `xfs`
	Fstype string `hcl:"fstype" required:"true"`

	// RequiredBy is a list of dependencies, to pass to systemd .mount unit
	RequiredBy []string `hcl:"requiredBy"`

	// WantedBy is a list of dependencies, to pass to systemd .mount unit
	WantedBy []string `hcl:"wantedBy"`

	// Before is a list of dependencies, to pass to systemd .mount unit
	Before []string `hcl:"before"`
}

// Prepare a new task
func (p *Preparer) Prepare(render resource.Renderer) (resource.Task, error) {

	m := &Mount{
		What:       p.Device,
		Where:      p.Mountpoint,
		Type:       p.Fstype,
		RequiredBy: strings.Join(p.RequiredBy, " "),
		WantedBy:   strings.Join(p.WantedBy, " "),
		Before:     strings.Join(p.Before, " "),
	}

	return NewResourceFS(lowlevel.MakeLvmBackend(), m)
}

func init() {
	registry.Register("filesystem", (*Preparer)(nil), (*resourceFS)(nil))
}
