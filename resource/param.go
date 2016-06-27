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

package resource

// Param is essentially the calling arguments of a module
type Param struct {
	DependencyTracker `hcl:",squash"`

	Name    string
	Default Value  `hcl:"default"`
	Type    string `hcl:"type"`

	parent *Module
}

// String returns the name of this param. It satisfies the fmt.Stringer
// interface.
func (p *Param) String() string {
	return "param." + p.Name
}

// Prepare this module for use
func (p *Param) Prepare(parent *Module) error {
	p.parent = parent
	return nil
}

// Value returns either a value set by the parameters or a default.
func (p *Param) Value() Value {
	if val, ok := p.parent.RenderedArgs[p.Name]; ok {
		return val
	}
	return p.Default
}

// SetName modifies the name of this param
func (p *Param) SetName(name string) {
	p.Name = name
}
