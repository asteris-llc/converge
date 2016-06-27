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
	"os"
	"strconv"

	"github.com/asteris-llc/converge/resource"
)

// Mode controls file mode of underlying resources
type Mode struct {
	resource.DependencyTracker `hcl:",squash"`

	Name           string
	RawDestination string   `hcl:"destination"`
	RawMode        string   `hcl:"mode"`
	Dependencies   []string `hcl:"depends"`

	destination string
	mode        os.FileMode
	renderer    *resource.Renderer
}

// Prepare this resource for use
func (m *Mode) Prepare(parent *resource.Module) (err error) {
	m.renderer, err = resource.NewRenderer(parent)

	// render destination
	m.destination, err = m.renderer.Render(m.String()+".destination", m.RawDestination)
	if err != nil {
		return err
	}

	// render mode
	smode, err := m.renderer.Render(m.String()+".mode", m.RawMode)
	if err != nil {
		return err
	}
	imode, err := strconv.ParseUint(smode, 8, 32)
	if err != nil {
		return &resource.ValidationError{Location: m.String() + ".mode", Err: err}
	}
	m.mode = os.FileMode(imode)

	// render dependencies
	m.Dependencies, err = m.renderer.Dependencies(
		m.String()+".dependencies",
		m.Dependencies,
		m.RawDestination, m.RawMode,
	)
	if err != nil {
		return err
	}

	return nil
}

func (m *Mode) String() string {
	return "file.mode." + m.Name
}

// SetName modifies the name of this mode
func (m *Mode) SetName(name string) {
	m.Name = name
}

// Check whether the destination has the right mode
func (m *Mode) Check() (status string, willChange bool, err error) {
	stat, err := os.Stat(m.destination)
	if err != nil {
		return
	}

	mode := stat.Mode().Perm()

	return strconv.FormatUint(uint64(mode), 8), m.mode != mode, nil
}

// Apply the changes in mode
func (m *Mode) Apply() error {
	return os.Chmod(m.destination, m.mode)
}
