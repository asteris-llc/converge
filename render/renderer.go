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
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/render/extensions"
	"github.com/asteris-llc/converge/render/preprocessor"
	"github.com/asteris-llc/converge/resource"
)

// ErrUnresolvable is returned by Render if the template string tries to resolve
// unaccesible node properties.
var ErrUnresolvable = errors.New("node is unresolvable")

// Renderer to be passed to preparers, which will render strings
type Renderer struct {
	Graph           func() *graph.Graph
	ID              string
	DotValue        string
	DotValuePresent bool
	resolverErr     bool
	Language        *extensions.LanguageExtension
}

// Value of this renderer
func (r *Renderer) Value() (value string, present bool) {
	return r.DotValue, r.DotValuePresent
}

// Render a string with text/template
func (r *Renderer) Render(name, src string) (string, error) {
	r.resolverErr = false
	r.Language = r.Language.On("param", r.param)
	r.Language = r.Language.On(extensions.RefFuncName, r.lookup)
	out, err := r.Language.Render(r.DotValue, name, src)
	if err != nil {
		if r.resolverErr {
			return "", ErrUnresolvable
		}
		return "", err
	}
	return out.String(), err
}

func (r *Renderer) param(name string) (string, error) {
	val, ok := resource.ResolveTask(r.Graph().Get(graph.SiblingID(r.ID, "param."+name)))

	if val == nil || !ok {
		return "", errors.New("param not found")
	}

	if _, ok := val.(*PrepareThunk); ok {
		r.resolverErr = true
		return "", ErrUnresolvable
	}

	return fmt.Sprintf("%+v", val), nil
}

func (r *Renderer) lookup(name string) (string, error) {
	g := r.Graph()
	// fully-qualified graph name

	fqgn := graph.SiblingID(r.ID, name)
	vertexName, terms, found := preprocessor.VertexSplit(g, fqgn)

	if !found {
		return "", fmt.Errorf("%s does not resolve to a valid node", fqgn)
	}

	if _, isThunk := g.Get(vertexName).(*PrepareThunk); isThunk {
		log.Println("[INFO] node is unresolvable by proxy-reference to ", vertexName)
		r.resolverErr = true
		return "", ErrUnresolvable
	}

	val, ok := resource.ResolveTask(g.Get(vertexName))

	if !ok {
		p := g.Get(vertexName)
		return "", fmt.Errorf("%s is not a valid task node (type: %T)", vertexName, p)
	}

	result, err := preprocessor.EvalTerms(val, preprocessor.SplitTerms(terms)...)

	if err != nil {
		if err == preprocessor.ErrUnresolvable {
			r.resolverErr = true
			return "", ErrUnresolvable
		}
		return "", err
	}

	return fmt.Sprintf("%v", result), nil
}

// RequiredRender will return an error if rendered value is an empty string
func (r *Renderer) RequiredRender(name, src string) (string, error) {
	rendered, err := r.Render(name, src)
	if err != nil {
		return "", err
	}

	if rendered == "" {
		return "", fmt.Errorf("%s is required", name)
	}

	return rendered, nil
}

// RenderBool renders a boolean value
func (r *Renderer) RenderBool(name, src string) (bool, error) {
	var b bool
	rendered, err := r.Render(name, src)
	if err != nil {
		return b, err
	}

	if rendered == "" {
		return b, nil
	}

	return strconv.ParseBool(rendered)
}

// RenderStringSlice renders a slice of strings
func (r *Renderer) RenderStringSlice(name string, src []string) ([]string, error) {
	renderedSlice := make([]string, len(src))
	for i, val := range src {
		rendered, err := r.Render(fmt.Sprintf("%s[%d]", name, i), val)
		if err != nil {
			return nil, err
		}
		renderedSlice[i] = rendered
	}
	return renderedSlice, nil
}

// RenderStringMapToStringSlice renders a map of strings to strings as a string
// slice
func (r *Renderer) RenderStringMapToStringSlice(name string, src map[string]string, toString func(string, string) string) ([]string, error) {
	if toString == nil {
		toString = func(k, v string) string { return k + " " + v }
	}

	renderedSlice := make([]string, len(src))
	idx := 0
	for key, val := range src {
		pair := toString(key, val)
		rendered, err := r.Render(fmt.Sprintf("%s[%s]", name, val), pair)
		if err != nil {
			return nil, err
		}
		renderedSlice[idx] = rendered
		idx++
	}

	return renderedSlice, nil
}
