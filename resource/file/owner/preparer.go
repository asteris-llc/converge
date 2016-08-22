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

package owner

import (
	"fmt"
	"os/user"

	"github.com/asteris-llc/converge/resource"
)

// Preparer for file mode
type Preparer struct {
	Destination string `hcl:"Destination"`
	User        string `hcl:"user"`
}

// Prepare this resource for use
func (p *Preparer) Prepare(render resource.Renderer) (resource.Task, error) {
	// render Destination
	destination, err := render.Render("Destination", p.Destination)
	if err != nil {
		return nil, err
	}
	username, err := render.Render("user", p.User)
	if err != nil {
		return nil, err
	}
	if username == "" {
		return nil, fmt.Errorf("task requires a %q parameter", "user")

	}
	actualUser, err := user.Lookup(username)
	if err != nil {
		return nil, err
	}

	ownerTask := &Owner{User: actualUser, Destination: destination}
	return ownerTask, ownerTask.Validate()
}
