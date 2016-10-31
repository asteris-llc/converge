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
	"path/filepath"

	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/lvm/lowlevel"
)

// Preparer for LVM's Volume Group
type Preparer struct {
	Name    string   `hcl:"name",required:"true"`
	Devices []string `hcl:"devices"`
}

// Prepare a new task
func (p *Preparer) Prepare(render resource.Renderer) (resource.Task, error) {
	// Device paths need to be real devices, not symlinks
	// (otherwise it breaks on GCE)
	devices := make([]string, len(p.Devices))
	for i, dev := range p.Devices {
		var err error
		devices[i], err = filepath.EvalSymlinks(dev)
		if err != nil {
			return nil, err
		}
	}

	rvg := NewResourceVG(lowlevel.MakeLvmBackend(), p.Name, devices)
	return rvg, nil
}
