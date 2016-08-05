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

package directory

import (
	"errors"
	"fmt"

	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/file/mode"
	"github.com/asteris-llc/converge/resource/file/owner"
)

// Preparer for Content
type Preparer struct {
	Destination string `hcl:"destination"`
	Recurse     bool   `hcl:"recurse"`

	Mode string `hcl:"mode"`
	User string `hcl:"user"`
}

// Prepare a new task
// Prepare this resource for use
func (p *Preparer) Prepare(render resource.Renderer) (resource.Task, error) {
	// render Destination
	destination, err := render.Render("destination", p.Destination)
	if err != nil {
		return nil, err
	}
	if destination == "" {
		return nil, fmt.Errorf("resource requires a destination parameter\n%s")
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

	directoryModule := &Directory{
		Destination: destination,
		Recurse:     p.Recurse,
		Mode:        fileMode,
		Owner:       fileOwner,
	}
	return directoryModule, ValidateTask(directoryModule)

}

func ValidateTask(d *Directory) error {
	if d.Destination == "" {
		return errors.New("resource requires a `destiination` parameter")
	}
	return nil
}
