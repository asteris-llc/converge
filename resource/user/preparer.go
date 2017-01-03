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
	"time"

	"github.com/asteris-llc/converge/load/registry"
	"github.com/asteris-llc/converge/resource"
	"golang.org/x/net/context"
)

// Preparer for User
//
// User renders user data
type Preparer struct {
	// Username is the user login name.
	Username string `hcl:"username" required:"true" nonempty:"true"`

	// NewUsername is used when modifying a user.
	// Username will be changed to NewUsername. No changes to the home directory
	// name or location of the contents will be made. This can be done using
	// HomeDir and MoveDir options.
	NewUsername string `hcl:"new_username" nonempty:"true"`

	// UID is the user ID.
	UID *uint32 `hcl:"uid"`

	// GroupName is the primary group for user and must already exist.
	// Only one of GID or Groupname may be indicated.
	GroupName string `hcl:"groupname" mutually_exclusive:"gid,groupname" nonempty:"true"`

	// Gid is the primary group ID for user and must refer to an existing group.
	// Only one of GID or Groupname may be indicated.
	GID *uint32 `hcl:"gid" mutually_exclusive:"gid,groupname"`

	// Name is the user description.
	// This field can be indicated when adding or modifying a user.
	Name string `hcl:"name" nonempty:"true"`

	// CreateHome when set to true will create the home directory for the user.
	// The files and directories contained in the skeleton directory (which can be
	// defined with the SkelDir option) will be copied to the home directory.
	CreateHome bool `hcl:"create_home"`

	// SkelDir contains files and directories to be copied in the user's home
	// directory when adding a user. If not set, the skeleton directory is defined
	// by the SKEL variable in /etc/default/useradd or, by default, /etc/skel.
	// SkelDir is only valid is CreatHome is specified.
	SkelDir string `hcl:"skel_dir" nonempty:"true"`

	// HomeDir is the name of the user's login directory. If not set, the home
	// directory is defined by appending the value of Username to the HOME
	// variable in /etc/default/useradd, resulting in /HOME/Username.
	// This field can be indicated when adding or modifying a user.
	HomeDir string `hcl:"home_dir" nonempty:"true"`

	// MoveDir is used to move the contents of HomeDir when modifying a user.
	// HomeDir must also be indicated if MoveDir is set to true.
	MoveDir bool `hcl:"move_dir"`

	// Expiry is the date on which the user account will be disabled. The date is
	// specified in the format YYYY-MM-DD. If not specified, the default expiry
	// date specified by the EXPIRE variable in /ect/default/useradd, or an empty
	// string (no expiry) will be used by default.
	Expiry time.Time `hcl:"expiry"`

	// State is whether the user should be present.
	// The default value is present.
	State State `hcl:"state" valid_values:"present,absent"`
}

// Prepare a new task
func (p *Preparer) Prepare(ctx context.Context, render resource.Renderer) (resource.Task, error) {
	if p.UID != nil && *p.UID == math.MaxUint32 {
		// the maximum uid on linux is MaxUint32 - 1
		return nil, fmt.Errorf("user \"uid\" parameter out of range")
	}

	if p.GID != nil && *p.GID == math.MaxUint32 {
		// the maximum gid on linux is MaxUint32 - 1
		return nil, fmt.Errorf("user \"gid\" parameter out of range")
	}

	if p.SkelDir != "" && !p.CreateHome {
		return nil, fmt.Errorf("user \"create_home\" parameter required with \"skel_dir\" parameter")
	}

	if p.MoveDir && p.HomeDir == "" {
		return nil, fmt.Errorf("user \"home_dir\" parameter required with \"move_dir\" parameter")
	}

	if p.State == "" {
		p.State = StatePresent
	}

	usr := NewUser(new(System))
	usr.Username = p.Username
	usr.NewUsername = p.NewUsername
	usr.GroupName = p.GroupName
	usr.Name = p.Name
	usr.CreateHome = p.CreateHome
	usr.SkelDir = p.SkelDir
	usr.HomeDir = p.HomeDir
	usr.MoveDir = p.MoveDir
	usr.State = p.State
	usr.Expiry = p.Expiry

	if p.UID != nil {
		usr.UID = fmt.Sprintf("%v", *p.UID)
	}

	if p.GID != nil {
		usr.GID = fmt.Sprintf("%v", *p.GID)
	}

	return usr, nil
}

func init() {
	registry.Register("user.user", (*Preparer)(nil), (*User)(nil))
}
