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
	// vaild states are [  "present", "absent" ]
	State string `hcl:"state"`

	// Type sets the type of the resource
	// valid types are [ "file", "directory", "hardlink", "symlink"]
	Type string `hcl:"type"`

	// Target is the target file for a symbolic or hard link
	// destination -> target
	Target string `hcl:"target"`

	// Force Change the resource. For example, if the target is a file and
	// state is set to directory, the file will be removed
	// Force on a symlink will remove the previous symlink
	Force string `hcl:"force" doc_type:"bool"`

	// Mode is the mode of the file, specified in octal.
	Mode string `hcl:"mode" doc_type:"octal string"`

	// User is the user name of the file owner
	User string `hcl:"user"`

	// Group is the groupname
	Group string `hcl:"group"`

	// Content of the file
	Content string `hcl:"content"`
}

// Prepare a file resource
func (p *Preparer) Prepare(render resource.Renderer) (resource.Task, error) {

	// render Destination
	Destination, err := render.Render("destination", p.Destination)
	if err != nil {
		return nil, err
	}

	State, err := render.Render("state", p.State)
	if err != nil {
		return nil, err
	}

	Type, err := render.Render("type", p.Type)
	if err != nil {
		return nil, err
	}

	Target, err := render.Render("target", p.Target)
	if err != nil {
		return nil, err
	}

	Force, err := render.RenderBool("target", p.Force)
	if err != nil {
		return nil, err
	}

	// render Mode
	Mode, err := render.Render("mode", p.Mode)
	if err != nil {
		return nil, err
	}

	FileMode, err := UnixMode(Mode)
	if err != nil {
		return nil, err
	}

	User, err := render.Render("user", p.User)
	if err != nil {
		return nil, err
	}

	Group, err := render.Render("group", p.Group)
	if err != nil {
		return nil, err
	}

	Content, err := render.Render("content", p.Content)
	if err != nil {
		return nil, err
	}

	fileTask := &File{
		Destination: Destination,
		State:       State,
		Type:        Type,
		Target:      Target,
		Force:       Force,
		FileMode:    FileMode,
		User:        User,
		Group:       Group,
		Content:     Content,
	}

	return fileTask, fileTask.Validate()
}

func setDefault(s string, def string) string {
	if s != "" {
		return s
	}
	return def
}

func init() {
	registry.Register("file", (*Preparer)(nil), (*File)(nil))
}
