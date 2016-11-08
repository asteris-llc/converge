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

package lv

import (
	"github.com/asteris-llc/converge/load/registry"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/lvm/lowlevel"
	"golang.org/x/net/context"
)

// Preparer for LVM LV resource
//
// Logical volume creation
type Preparer struct {
	// Group where volume will be created
	Group string `hcl:"group" required:"true"`

	// Name of volume, which will be created
	Name string `hcl:"name" required:"true"`

	// Size of volume. Can be relative or absolute.
	// Relative size set in forms like `100%FREE`
	// (words after percent sign can be `FREE`, `VG`, `PV`)
	// Absolute size specified with suffix `BbKkMmGgTtPp`, upper case
	// suffix mean S.I. sizes (power of 10), lower case mean powers of 1024.
	// Also special suffixes `Ss`, which mean sectors.
	// Refer to LVM manpages for details.
	Size string `hcl:"size" required:"true"`
}

// Prepare a new task
func (p *Preparer) Prepare(_ context.Context, render resource.Renderer) (resource.Task, error) {
	size, err := lowlevel.ParseSize(p.Size)
	if err != nil {
		return nil, err
	}

	r := NewResourceLV(lowlevel.MakeLvmBackend(), p.Group, p.Name, size)
	return r, nil
}

func init() {
	registry.Register("lvm.logicalvolume", (*Preparer)(nil), (*resourceLV)(nil))
}
