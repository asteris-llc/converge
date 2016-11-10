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

package vg

import (
	"golang.org/x/net/context"

	"github.com/asteris-llc/converge/load/registry"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/lvm/lowlevel"
)

// Preparer for LVM's Volume Group
//
// Volume group is responsible for creation LVM Volume Groups
// from given block devices.
type Preparer struct {
	// Name of created volume group
	Name string `hcl:"name" required:"true"`

	// Devices is list of entities to include into volume group
	Devices []string `hcl:"devices"`

	// Remove is enable removal devices omitted from `Devices` list from
	// from volume group
	Remove bool `hcl:"remove"`

	// ForceRemove control destruction of volumes after removing
	// from volume group
	ForceRemove bool `hcl:"forceRemove"`
}

// Prepare a new task
func (p *Preparer) Prepare(_ context.Context, render resource.Renderer) (resource.Task, error) {
	rvg := NewResourceVG(lowlevel.MakeLvmBackend(), p.Name, p.Devices, p.Remove, p.ForceRemove)
	return rvg, nil
}

func init() {
	registry.Register("lvm.volumegroup", (*Preparer)(nil), (*resourceVG)(nil))
}
