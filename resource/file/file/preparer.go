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
	"errors"
	"fmt"

	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/file/mode"
	"github.com/asteris-llc/converge/resource/file/owner"
)

// Preparer for file Mode
type Preparer struct {
	Destination string `hcl:"destination"`
	Source      string `hcl:"source"`
	State       string `hcl:"state"`
	Force       string `hcl:"force"`
	Recurse     string `hcl:"recurse"`

	Mode string `hcl:"mode"`
	User string `hcl:"user"`
}

// Prepare this resource for use
func (p *Preparer) Prepare(render resource.Renderer) (resource.Task, error) {
	// render Destination
	destination, err := render.Render("destination", p.Destination)
	if err != nil {
		return nil, err
	}
	if destination == "" {
		return nil, fmt.Errorf("file.mode requires a destination parameter\n%s", PrintExample())
	}
	source, err := render.Render("source", p.Source)
	if err != nil {
		return nil, err
	}
	state, err := render.Render("state", p.State)
	if err != nil {
		return nil, err
	}
	recurse, err := render.Render("recurse", p.Recurse)
	if err != nil {
		return nil, err
	}
	// render Mode
	sMode, err := render.Render("mode", p.Mode)
	if err != nil {
		return nil, err
	}
	// render user
	username, err := render.Render("user", p.User)
	if err != nil {
		return nil, err
	}
	var fileMode *mode.Mode
	var fileOwner *owner.Owner
	if sMode != "" {
		prep := &mode.Preparer{Destination: destination, Mode: sMode}
		task, err := prep.Prepare(render)
		if err != nil {
			return nil, err
		}
		fileMode = task.(*mode.Mode)
	}
	if username != "" {
		prep := &owner.Preparer{Destination: destination, User: username}
		task, err := prep.Prepare(render)
		if err != nil {
			return nil, err
		}
		fileOwner = task.(*owner.Owner)
	}

	fileModule := &File{
		State:       FileState(state),
		Source:      source,
		Destination: destination,
		Recurse:     recurse == "true",
		Mode:        fileMode,
		Owner:       fileOwner,
	}
	return fileModule, ValidateFileModule(fileModule)

}

func ValidateFileModule(fileModule *File) error {
	if fileModule.Destination == "" {
		return errors.New("module 'file' requires a `destiination` parameter")
	}
	if fileModule.Recurse && fileModule.State != FSDirectory {
		return fmt.Errorf("cannot use the `recurse` parameter when `state` is not set to %q", FSDirectory)
	}
	switch fileModule.State {
	case FSAbsent:
		if fileModule.Mode != nil || fileModule.Owner != nil {
			return fmt.Errorf("cannot use `mode` or `owner` parameters when `state` is set to %q", FSAbsent)
		}
		return nil
	case FSTouch:
		return nil
	case FSLink:
		if fileModule.Source == "" {
			return fmt.Errorf("module 'file' requires a `source` parameter when `state`=%q", FSLink)
		}
		return nil
	case FSHard:
		if fileModule.Source == "" {
			return fmt.Errorf("module 'file' requires a `source` parameter when `state`=%q", FSHard)
		}
		return nil
	case FSDirectory:
		return nil
	case FSFile:
		fallthrough
	case "":
		if fileModule.Mode == nil && fileModule.Mode == nil {
			return errors.New("useless file module")
		}
		return nil
	default:
		return fmt.Errorf("invalid value for 'state' parameter. found %q", fileModule.State)
	}
}

func PrintExample() string {
	return fmt.Sprintln(
		`	Example
		--------------------
		file.mode "makepublic" {
		    destination = "/path/to/file.txt"
		    mode = 777
		}
		`)
}
