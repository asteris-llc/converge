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

package user

import (
	"fmt"
	"math"
	"strconv"

	"github.com/asteris-llc/converge/load/registry"
	"github.com/asteris-llc/converge/resource"
)

// Preparer for User
//
// User renders user data
type Preparer struct {
	// Username is the user login name.
	Username string `hcl:"username"`

	// UID is the user ID.
	UID string `hcl:"uid"`

	// Groupname is the primary group for user and must already exist.
	// Only one of GID or Groupname may be indicated.
	Groupname string `hcl:"groupname"`

	// Gid is the primary group ID for user and must refer to an existing group.
	// Only one of GID or Groupname may be indicated.
	GID string `hcl:"gid"`

	// Name is the user description.
	Name string `hcl:"name"`

	// HomeDir is the user's login directory. By default,  the login
	// name is appended to the home directory.
	HomeDir string `hcl:"home_dir"`

	// State is whether the user should be present.
	// Options are present and absent; default is present.
	State string `hcl:"state"`
}

// Prepare a new task
func (p *Preparer) Prepare(render resource.Renderer) (resource.Task, error) {
	username, err := render.Render("username", p.Username)
	if err != nil {
		return nil, err
	}
	if username == "" {
		return nil, fmt.Errorf("user requires a \"username\" parameter")
	}
	usr := NewUser(new(System))
	usr.Username = username

	uid, err := render.Render("uid", p.UID)
	if err != nil {
		return nil, err
	}
	if uid != "" {
		uidVal, err := strconv.ParseUint(uid, 10, 32)
		if err != nil {
			return nil, err
		}
		if uidVal == math.MaxUint32 {
			// the maximum uid on linux is MaxUint32 - 1
			return nil, fmt.Errorf("user \"uid\" parameter out of range")
		}
		usr.UID = uid
	}

	groupName, err := render.Render("groupname", p.Groupname)
	if err != nil {
		return nil, err
	}
	gid, err := render.Render("gid", p.GID)
	if err != nil {
		return nil, err
	}
	if groupName != "" && gid != "" {
		return nil, fmt.Errorf("user \"groupname\" and \"gid\" both indicated, choose one")
	}
	if groupName != "" {
		usr.GroupName = groupName
	} else if gid != "" {
		gidVal, err := strconv.ParseUint(gid, 10, 32)
		if err != nil {
			return nil, err
		}
		if gidVal == math.MaxUint32 {
			// the maximum gid on linux is MaxUint32 - 1
			return nil, fmt.Errorf("user \"gid\" parameter out of range")
		}
		usr.GID = gid
	}

	name, err := render.Render("name", p.Name)
	if err != nil {
		return nil, err
	}
	if name != "" {
		usr.Name = name
	}

	homeDir, err := render.Render("home_dir", p.HomeDir)
	if err != nil {
		return nil, err
	}
	if homeDir != "" {
		usr.HomeDir = homeDir
	}

	sstate, err := render.Render("name", p.State)
	state := State(sstate)
	if err != nil {
		return nil, err
	}
	if state == "" {
		state = StatePresent
	} else if state != StatePresent && state != StateAbsent {
		return nil, fmt.Errorf("user \"state\" parameter invalid, use present or absent")
	}
	usr.State = state

	return usr, nil
}

func init() {
	registry.Register("user.user", (*Preparer)(nil), (*User)(nil))
}
