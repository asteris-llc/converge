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
	"golang.org/x/net/context"
)

// Module holds stringified values for parameters
type Module struct {
	resource.Status

	// the params configured for the module
	Params map[string]resource.Value `export:"params"`
}

// Check just returns the current value of the moduleeter. It should never have to change.
func (m *Module) Check(context.Context, resource.Renderer) (resource.TaskStatus, error) {
	m.Status = resource.Status{Output: []string{m.String()}}

	return m, nil
}

// Apply doesn't do anything since modules are final values
func (m *Module) Apply(context.Context) (resource.TaskStatus, error) {
	return m, nil
}

// String is the final value of thie Module
func (m *Module) String() string {
	var lines []string
	for key, val := range m.Params {
		lines = append(lines, fmt.Sprintf("%s: %s", key, val))
	}
	return strings.Join(lines, "\n")
}
