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

package group

import (
	"fmt"
	"math"

	"github.com/asteris-llc/converge/load/registry"
	"github.com/asteris-llc/converge/resource"
)

// Preparer for Group
//
// Group renders group data
type Preparer struct {
	// Gid is the group gid.
	GID uint32 `hcl:"gid"`

	// Name is the group name.
	Name string `hcl:"name" required:"true"`

	// State is whether the group should be present.
	State State `hcl:"state" valid_values:"present,absent"`
}

// Prepare a new task
func (p *Preparer) Prepare(render resource.Renderer) (resource.Task, error) {
	if p.GID == math.MaxUint32 {
		// the maximum gid on linux is MaxUint32 - 1
		return nil, fmt.Errorf("group \"gid\" parameter out of range")
	}

	if p.State == "" {
		p.State = StatePresent
	}

	grp := NewGroup(new(System))
	grp.Name = p.Name
	grp.GID = fmt.Sprintf("%v", p.GID)
	grp.State = p.State

	return grp, nil
}

func init() {
	registry.Register("user.group", (*Preparer)(nil), (*Group)(nil))
}
