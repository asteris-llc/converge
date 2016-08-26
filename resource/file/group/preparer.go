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
	"os/user"

	"github.com/asteris-llc/converge/resource"
)

// Preparer for file mode
type Preparer struct {
	Destination string `hcl:"destination"`
	Group       string `hcl:"group"`
	Gid         string `hcl:"group"`
}

// Prepare this resource for use
func (p *Preparer) Prepare(render resource.Renderer) (resource.Task, error) {
	// render Destination
	destination, err := render.Render("destination", p.Destination)
	if err != nil {
		return nil, err
	}
	group, err := render.Render("group", p.Group)
	if err != nil {
		return nil, err
	}
	gid, err := render.Render("gid", p.Gid)
	if err != nil {
		return nil, err
	}
	if group == "" && gid == "" {
		return nil, fmt.Errorf("task requires a %q or %q parameter", "group", "gid")
	}
	var g *user.Group
	if group == "" {
		g, err = user.LookupGroupId(gid)
	} else {
		g, err = user.LookupGroup(group)
	}
	if err != nil {
		return nil, err
	}

	groupTask := &Group{Group: g, Destination: destination}
	return groupTask, groupTask.Validate()
}
