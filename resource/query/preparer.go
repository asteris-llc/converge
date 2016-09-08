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

package query

import (
	"fmt"

	"github.com/asteris-llc/converge/load/registry"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/shell"
)

// Preparer handles querying
type Preparer struct {
	Interpreter string            `hcl:"interpreter"`
	Query       string            `hcl:"query"`
	Flags       []string          `hcl:"flags"`
	Dir         string            `hcl:"dir"`
	Env         map[string]string `hcl:"env"`
}

// Prepare creates a new query type
func (p *Preparer) Prepare(render resource.Renderer) (resource.Task, error) {
	interpreter, err := render.Render("interpreter", p.Interpreter)
	if err != nil {
		return nil, err
	}

	query, err := render.Render("query", p.Query)
	if err != nil {
		return nil, err
	}

	dir, err := render.Render("dir", p.Dir)
	if err != nil {
		return nil, err
	}

	env, err := render.RenderStringMapToStringSlice("env", p.Env, func(k, v string) string {
		return fmt.Sprintf("%s=%s", k, v)
	})

	if err != nil {
		return nil, err
	}

	generator := &shell.CommandGenerator{
		Interpreter: interpreter,
		Dir:         dir,
		Env:         env,
		Flags:       p.Flags,
	}

	return &Query{
		CmdGenerator: generator,
		Query:        query,
	}, nil
}

func init() {
	registry.Register("query", (*Preparer)(nil), (*Query)(nil))
}
