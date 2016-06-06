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

import "fmt"

// Value contains the different values for a param
type Value string

func (v Value) String() string {
	return string(v)
}

// Values is a named collection of values
type Values map[string]Value

// Param is essentially the calling arguments of a module
type Param struct {
	ParamName string
	Default   Value  `hcl:"default"`
	Type      string `hcl:"type"`

	parent *Module
}

// ValidationError is the type returned by each resource's Validate method. It
// describes both what went wrong and which stanza caused the problem.
type ValidationError struct {
	Location string
	Err      error
}

func (v ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", v.Location, v.Err)
}

// Name returns the name of this param
func (p *Param) Name() string {
	return p.ParamName
}

// Validate that this value is correct
func (p *Param) Validate() error {
	return nil
}

// Prepare this module for use
func (p *Param) Prepare(parent *Module) error {
	p.parent = parent
	return nil
}

// Value returns either a value set by the parameters or a default.
func (p *Param) Value() Value {
	if val, ok := p.parent.Args[p.ParamName]; ok {
		return val
	}
	return p.Default
}
