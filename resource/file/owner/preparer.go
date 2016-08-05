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
	"strconv"

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
	if destination == "" {
		return nil, fmt.Errorf("file.owner requires a destination parameter.\ns", PrintExample())
	}
	username, err := render.Render("user", p.User)
	if err != nil {
		return nil, err
	}
	if username == "" {
		return nil, fmt.Errorf("file.owner requires a user parameter.\n%s", PrintExample())
	}
	actualUser, err := user.Lookup(username)
	if err != nil {
		return nil, err
	}
	uid, _ := strconv.Atoi(actualUser.Uid)
	gid, _ := strconv.Atoi(actualUser.Gid)

	return &Owner{Username: username, UID: uid, GID: gid, Destination: destination}, nil
}

func PrintExample() string {
	return fmt.Sprintln(
		`	Example
		--------------------
		file.owner "makenobody's" {
		    destination = "/path/to/file.txt"
		    owner = nobody
		}
		`)
}
