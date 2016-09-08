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

package module

import (
	"fmt"
	"strings"

	"github.com/asteris-llc/converge/resource"
)

// Module holds stringified values for parameters
type Module struct {
	Params map[string]string
}

// Check just returns the current value of the moduleeter. It should never have to change.
func (m *Module) Check(resource.Renderer) (resource.TaskStatus, error) {
	return &resource.Status{Status: m.String(), WillChange: false}, nil
}

// Apply doesn't do anything since modules are final values
func (m *Module) Apply(r resource.Renderer) (resource.TaskStatus, error) {
	return m.Check(r)
}

// String is the final value of thie Module
func (m *Module) String() string {
	var lines []string
	for key, val := range m.Params {
		lines = append(lines, fmt.Sprintf("%s: %s", key, val))
	}
	return strings.Join(lines, "\n")
}
