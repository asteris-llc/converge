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

package mode

import (
	"fmt"
	"os"
	"strconv"

	"github.com/asteris-llc/converge/load/registry"
	"github.com/asteris-llc/converge/resource"
	"github.com/pkg/errors"
)

// Preparer for file Mode
//
// Mode monitors the mode of a file
type Preparer struct {
	// Destination specifies which file will be modified by this resource. The
	// file must exist on the system (for example, having been created with
	// `file.content`.)
	Destination string `hcl:"destination"`

	// Mode is the mode of the file, specified in octal.
	Mode string `hcl:"mode" doc_type:"octal string"`
}

// Prepare this resource for use
func (p *Preparer) Prepare(render resource.Renderer) (resource.Task, error) {
	// render Destination
	Destination, err := render.Render("destination", p.Destination)
	if err != nil {
		return nil, err
	}
	// render Mode
	sMode, err := render.Render("mode", p.Mode)
	if err != nil {
		return nil, err
	}
	if sMode == "" {
		return nil, fmt.Errorf("task requires a \"mode\" parameter")
	}
	iMode, err := strconv.ParseUint(sMode, 8, 32)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("%q is not a valid file mode", sMode))
	}
	mode := os.FileMode(iMode)

	modeTask := &Mode{Destination: Destination, Mode: mode}
	return modeTask, modeTask.Validate()
}

func init() {
	registry.Register("file.mode", (*Preparer)(nil), (*Mode)(nil))
}
