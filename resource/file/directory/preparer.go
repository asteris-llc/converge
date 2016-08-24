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

import "github.com/asteris-llc/converge/resource"

// Preparer for file Mode
type Preparer struct {
	Destination string `hcl:"destination"`
	Force       bool   `hcl:"force"`
}

// Prepare this resource for use
func (p *Preparer) Prepare(render resource.Renderer) (resource.Task, error) {
	// render Destination
	Destination, err := render.Render("destination", p.Destination)
	if err != nil {
		return nil, err
	}
	// Force the creation of directory if file is already there

	directoryTask := &Directory{Destination: Destination, Force: p.Force}
	return directoryTask, directoryTask.Validate()
}
