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

package param

import (
	"errors"
	"fmt"

	"github.com/asteris-llc/converge/load/registry"
	"github.com/asteris-llc/converge/resource"
)

// Preparer for params
//
// Param controls the flow of values through `module` calls. You can use the
// `{{param "name"}}` template call anywhere you need the value of a param
// inside the current module.
type Preparer struct {
	// Default is an optional field that provides a default value if none is
	// provided to this parameter. If this field is not set, this param will be
	// treated as required.
	Default interface{} `hcl:"default"`
}

// Prepare a new task
func (p *Preparer) Prepare(render resource.Renderer) (resource.Task, error) {
	if val, present := render.Value(); present {
		return &Param{Value: val}, nil
	}

	if p.Default == nil {
		return nil, errors.New("param is required")
	}

	var def interface{}

	switch v := p.Default.(type) {
	case string:
		var err error
		def, err = render.Render("default", v)
		if err != nil {
			return nil, err
		}

	case bool, int, float32, float64:
		def = p.Default

	default:
		return nil, fmt.Errorf("composite values are not allowed in params, but got %T", v)
	}

	return &Param{Value: def}, nil
}

func init() {
	registry.Register("param", (*Preparer)(nil), (*Param)(nil))
}
