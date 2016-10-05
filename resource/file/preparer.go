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

package file

import (
	"os/user"

	"github.com/asteris-llc/converge/load/registry"
	"github.com/asteris-llc/converge/resource"
)

// Preparer for file
//
// File manages permissions, content and state for files
//
type Preparer struct {
	// Destination specifies which file will be modified by this resource. The
	// file must exist on the system (for example, having been created with
	// `file.content`.)
	Destination string `hcl:"destination"`

	// State sets whether the resource exists
	State string `hcl:"state" valid_values:"present,absent"`

	// Type sets the type of the resource
	// valid types are [ "file", "directory", "hardlink", "symlink"]
	Type string `hcl:"type" valid_values:"file,directory,hardlink,symlink"`

	// Target is the target file for a symbolic or hard link
	// destination -> target
	Target string `hcl:"target"`

	// Force Change the resource. For example, if the target is a file and
	// state is set to directory, the file will be removed
	// Force on a symlink will remove the previous symlink
	Force bool `hcl:"force" doc_type:"bool"`

	// Mode is the mode of the file, specified in octal (like 0755).
	Mode *uint32 `hcl:"mode" base:"8" doc_type:"uint32"`

	// User is the user name of the file owner
	User string `hcl:"user"`

	// Group is the groupname
	Group string `hcl:"group"`

	// Content of the file
	Content string `hcl:"content"`
}

// Prepare a file resource
func (p *Preparer) Prepare(render resource.Renderer) (resource.Task, error) {

	fileTask := &File{
		Destination: p.Destination,
		Mode:        p.Mode,
		State:       State(p.State),
		Type:        Type(p.Type),
		Target:      p.Target,
		Force:       p.Force,
		UserInfo:    &user.User{Username: p.User},
		GroupInfo:   &user.Group{Name: p.Group},
		Content:     []byte(p.Content),
	}

	return fileTask, fileTask.Validate()
}

func init() {
	registry.Register("file", (*Preparer)(nil), (*File)(nil))
}
