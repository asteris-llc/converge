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
	"github.com/asteris-llc/converge/load/registry"
	"github.com/asteris-llc/converge/resource"
	"golang.org/x/net/context"
)

// Preparer for Owern
//
// Owner sets the file and group ownership of a file or directory.  If
// `recursive` is set to true and `destination` is a directory, then it will
// also recursively change ownership of all files and subdirectories.  Symlinks
// are ignored.
type Preparer struct {
	// Destination is the location on disk where the content will be rendered.
	Destination string `hcl:"destination" required:"true" nonempty:"true"`

	// Recursive indicates whether ownership changes should be applied
	// recursively.  Symlinks are not followed.
	Recursive bool `hcl:"recursive" default:"false"`

	// Username specifies user-owernship by user name
	Username string `hcl:"username" mutally_exclusive:"uid"`

	// UID specifies user-ownership by UID
	UID string `hcl:"uid" mutually_exclusive:"username"`

	// Groupname specifies group-ownership by groupname
	Groupname string `hcl:"groupname" mutually_exclusive:"gid"`

	// GID specifies group ownership by gid
	GID string `hcl:"gid" mutually_exclusive:"groupname"`

	osProxy OSProxy
}

// Prepare a new task
func (p *Preparer) Prepare(ctx context.Context, render resource.Renderer) (resource.Task, error) {

	if p.osProxy == nil {
		p.osProxy = &OSExecutor{}
	}

	user, uid, err := normalizeUser(p.osProxy, p.Username, p.UID)
	if err != nil {
		return nil, err
	}

	group, gid, err := normalizeGroup(p.osProxy, p.Groupname, p.GID)
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
