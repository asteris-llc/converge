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

	"github.com/asteris-llc/converge/load/registry"
	"github.com/asteris-llc/converge/resource"
)

// Preparer for modules
//
// Module remotely sources other modules and adds them to the tree
type Preparer struct {
	// Params is a map of strings to anything you'd like. It will be passed to
	// the called module as the default values for the `param`s there.
	Params map[string]interface{} `hcl:"params"`
}

// NewPreparer returns a new preparer for modules
func NewPreparer(params map[string]interface{}) *Preparer {
	return &Preparer{Params: params}
}

// Prepare a new task
func (p *Preparer) Prepare(render resource.Renderer) (resource.Task, error) {
	module := &Module{Params: map[string]string{}}

	for key, value := range p.Params {
		switch value.(type) {
		case string:
			rendered, err := render.Render(key, value.(string))
			if err != nil {
				return nil, err
			}

			module.Params[key] = rendered

		default:
			module.Params[key] = fmt.Sprintf("%v", value)
		}
	}

	return module, nil
}

func init() {
	registry.Register("module", (*Preparer)(nil), (*Module)(nil))
}
