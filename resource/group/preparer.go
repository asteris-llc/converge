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
	"strconv"

	"github.com/asteris-llc/converge/load/registry"
	"github.com/asteris-llc/converge/resource"
)

// Preparer for Group
//
// Group renders group data
type Preparer struct {
	// Gid is the group gid.
	GID string `hcl:"gid"`

	// Name is the group name.
	Name string `hcl:"name"`

	// State is whether the group should be present.
	// Options are present and absent; default is present.
	State string `hcl:"state"`
}

// Prepare a new task
func (p *Preparer) Prepare(render resource.Renderer) (resource.Task, error) {
	name, err := render.Render("name", p.Name)
	if err != nil {
		return nil, err
	}
	if name == "" {
		return nil, fmt.Errorf("group requires a \"name\" parameter")
	}
	grp := NewGroup(new(System))
	grp.Name = name

	gid, err := render.Render("gid", p.GID)
	if err != nil {
		return nil, err
	}
	if gid != "" {
		gidVal, err := strconv.ParseUint(gid, 10, 32)
		if err != nil {
			return nil, err
		}
		if gidVal == math.MaxUint32 {
			// the maximum gid on linux is MaxUint32 - 1
			return nil, fmt.Errorf("group \"gid\" parameter out of range")
		}
		grp.GID = gid
	}

	sstate, err := render.Render("state", p.State)
	state := State(sstate)
	if err != nil {
		return nil, err
	}
	if state == "" {
		state = StatePresent
	} else if state != StatePresent && state != StateAbsent {
		return nil, fmt.Errorf("group \"state\" parameter invalid, use present or absent")
	}
	grp.State = state

	return grp, nil
}

func init() {
	registry.Register("user.group", (*Preparer)(nil), (*Group)(nil))
}
