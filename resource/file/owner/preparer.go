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
	"strconv"

	"github.com/asteris-llc/converge/load/registry"
	"github.com/asteris-llc/converge/resource"
	"golang.org/x/net/context"
)

// Preparer for Owner
//
// Owner sets the file and group ownership of a file or directory.  If
// `recursive` is set to true and `destination` is a directory, then it will
// also recursively change ownership of all files and subdirectories.  Symlinks
// are ignored.  If the file or directory does not exist during the plan phase
// of application the differences will be calculated during application.
// Otherwise changes will be limited to the files identified during the plan
// phase of application.
type Preparer struct {
	// Destination is the location on disk where the content will be rendered.
	Destination string `hcl:"destination" required:"true" nonempty:"true"`

	// Recursive indicates whether ownership changes should be applied
	// recursively.  Symlinks are not followed.
	Recursive bool `hcl:"recursive"`

	// Username specifies user-owernship by user name
	Username string `hcl:"user" mutally_exclusive:"user,uid"`

	// UID specifies user-ownership by UID
	UID *int `hcl:"uid" mutually_exclusive:"user,uid"`

	// Groupname specifies group-ownership by groupname
	Groupname string `hcl:"group" mutually_exclusive:"group,gid"`

	// GID specifies group ownership by gid
	GID *int `hcl:"gid" mutually_exclusive:"group,gid"`

	// Verbose specifies that when performing recursive updates, a diff should be
	// shown for each file to be changed
	Verbose bool `hcl:"verbose"`

	osProxy OSProxy
}

// Prepare a new task
func (p *Preparer) Prepare(ctx context.Context, render resource.Renderer) (resource.Task, error) {
	var uidStr string
	var gidStr string

	if p.osProxy == nil {
		p.osProxy = &OSExecutor{}
	}

	if p.UID != nil {
		uidStr = strconv.Itoa(*p.UID)
	}

	if p.GID != nil {
		gidStr = strconv.Itoa(*p.GID)
	}

	user, uid, err := normalizeUser(p.osProxy, p.Username, uidStr)
	if err != nil {
		return nil, err
	}

	group, gid, err := normalizeGroup(p.osProxy, p.Groupname, gidStr)
	if err != nil {
		return nil, err
	}

	return (&Owner{
		Destination: p.Destination,
		Recursive:   p.Recursive,
		Username:    user,
		UID:         uid,
		Group:       group,
		GID:         gid,
		HideDetails: !p.Verbose,
	}).SetOSProxy(p.osProxy), nil
}

// SetOSProxy sets the private os proxy for mocking in tests
func (p *Preparer) SetOSProxy(o OSProxy) *Preparer {
	p.osProxy = o
	return p
}

func init() {
	registry.Register("file.owner", (*Preparer)(nil), (*Owner)(nil))
}
