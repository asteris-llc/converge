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

import (
	"bytes"
	"html/template"
)

// NewRenderer creates a new Renderer
func NewRenderer(ctx *Module) (*Renderer, error) {
	renderer := &Renderer{
		ctx: ctx,
	}
	return renderer, nil
}

// Renderer renders template strings in Resources
type Renderer struct {
	ctx *Module
}

// Render the given template using the set context
func (r *Renderer) Render(source string) (string, error) {
	tmpl, err := template.New("").Parse(source)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, r.ctx)
	if err != nil {
		return "", nil
	}

	return buf.String(), nil
}
