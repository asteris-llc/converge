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

	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/file/directory"
	"github.com/asteris-llc/converge/resource/file/touch"
)

// Preparer for file Mode
type Preparer struct {
	Directory string `hcl:"directory"`
	File      string `hcl:"file"`
	Recurse   bool   `hcl:"recurse"`

	Mode string `hcl:"mode"`
	User string `hcl:"user"`
}

// Prepare this resource for use
func (p *Preparer) Prepare(render resource.Renderer) (resource.Task, error) {
	// render Destination
	directoryPath, err := render.Render("directory", p.Directory)
	if err != nil {
		return nil, err
	}
	filePath, err := render.Render("file", p.File)
	if err != nil {
		return nil, err
	}
	var fileDirectory *directory.Directory
	var fileTouch *touch.Touch
	if directoryPath != "" {
		prep := directory.Preparer{Destination: directoryPath, User: p.User, Mode: p.Mode, Recurse: p.Recurse}
		task, err := prep.Prepare(render)
		if err != nil {
			return nil, err
		}
		if task != nil {
			fileDirectory = task.(*directory.Directory)
		}
	}
	if filePath != "" {
		prep := touch.Preparer{Destination: filePath, User: p.User, Mode: p.Mode}
		task, err := prep.Prepare(render)
		if err != nil {
			return nil, err
		}
		if task != nil {
			fileTouch = task.(*touch.Touch)
		}
	}

	fileModule := &File{
		Directory: fileDirectory,
		Touch:     fileTouch,
	}
	return fileModule, ValidateTask(fileModule)

}

func ValidateTask(fileModule *File) error {
	if fileModule.Directory == nil && fileModule.Touch == nil {
		return errors.New("resource requires either a `directory` or a `file` parameter")
	}
	return nil
}
