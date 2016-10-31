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
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/lvm/lowlevel"

	"strings"
)

// Preparer for LVM FS Task
type Preparer struct {
	Device     string   `hcl:"device",required:"true"`
	Mount      string   `hcl:"mount",required:"true"`
	Fstype     string   `hcl:"fstype",reqired:"true"`
	RequiredBy []string `hcl:"requiredBy"`
	WantedBy   []string `hcl:"requiredBy"`
	Before     []string `hcl:"requiredBy"`
}

// Prepare a new task
func (p *Preparer) Prepare(render resource.Renderer) (resource.Task, error) {

	m := &Mount{
		What:       p.Device,
		Where:      p.Mount,
		Type:       p.Fstype,
		RequiredBy: strings.Join(p.RequiredBy, " "),
		WantedBy:   strings.Join(p.WantedBy, " "),
		Before:     strings.Join(p.Before, " "),
	}

	return NewResourceFS(lowlevel.MakeLvmBackend(), m)
}
