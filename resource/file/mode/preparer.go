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

	"github.com/asteris-llc/converge/resource"
	"github.com/pkg/errors"
)

// Preparer for file Mode
type Preparer struct {
	Destination string `hcl:"destination"`
	Mode        string `hcl:"mode"`
}

// Prepare this resource for use
func (p *Preparer) Prepare(render resource.Renderer) (resource.Task, error) {
	// render Destination
	Destination, err := render.Render("destination", p.Destination)
	if err != nil {
		return nil, err
	}
	if Destination == "" {
		return nil, fmt.Errorf("file.mode requires a destination parameter\n%s", PrintExample())
	}
	// render Mode
	sMode, err := render.Render("mode", p.Mode)
	if err != nil {
		return nil, err
	}
	if sMode == "" {
		return nil, fmt.Errorf("file.mode requires a mode parameter\n%s", PrintExample())
	}
	iMode, err := strconv.ParseUint(sMode, 8, 32)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("%q is not a valid file mode", sMode))
	}
	mode := os.FileMode(iMode)

	return &Mode{Destination: Destination, Mode: mode}, nil
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
