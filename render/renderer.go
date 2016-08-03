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

package render

import (
	"bytes"
	"errors"
	"fmt"
	"text/template"

	"github.com/asteris-llc/converge/extensions"
	"github.com/asteris-llc/converge/graph"
)

// Renderer to be passed to preparers, which will render strings
type Renderer struct {
	Graph           *graph.Graph
	ID              string
	DotValue        string
	DotValuePresent bool
}

// Value of this renderer
func (r *Renderer) Value() (value string, present bool) {
	return r.DotValue, r.DotValuePresent
}

// Render a string with text/template
func (r *Renderer) Render(name, src string) (string, error) {
	tmpl, err := template.New(name).Funcs(r.funcs()).Parse(src)
	if err != nil {
		return "", err
	}

	var dest bytes.Buffer
	err = tmpl.Execute(&dest, r.DotValue)

	return dest.String(), err
}

func (r *Renderer) funcs() template.FuncMap {
	language := extensions.MakeLanguage()
	language.On("split", extensions.DefaultSplit)
	language.On("param", r.param)
	return language.Funcs
}

func (r *Renderer) param(name string) (string, error) {
	val := r.Graph.GetSibling(r.ID, "param."+name)
	if val == nil {
		return "", errors.New("param not found")
	}

	return fmt.Sprintf("%+v", val), nil
}
