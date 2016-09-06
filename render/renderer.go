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
	Language        *extensions.LanguageExtension
}

// Value of this renderer
func (r *Renderer) Value() (value string, present bool) {
	return r.DotValue, r.DotValuePresent
}

// Render a string with text/template
func (r *Renderer) Render(name, src string) (string, error) {
	r.Language = r.Language.On("param", r.param)
	r.Language = r.Language.On(extensions.RefFuncName, r.lookup)
	out, err := r.Language.Render(r.DotValue, name, src)
	if err != nil {
		return "", err
	}
	return out.String(), err
}

func (r *Renderer) param(name string) (string, error) {
	val, ok := resource.ResolveTask(r.Graph().Get(graph.SiblingID(r.ID, "param."+name)))

	if val == nil || !ok {
		return "", errors.New("param not found")
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
	val, ok := resource.ResolveTask(g.Get(vertexName))
	if !ok {
		p := g.Get(vertexName)
		return "", fmt.Errorf("%s is not a valid task node (type: %T)", vertexName, p)
	}
	result, err := preprocessor.EvalTerms(val, preprocessor.SplitTerms(terms)...)
	if err != nil {
		return "", ErrUnresolvable
	}
	return fmt.Sprintf("%v", result), nil
}

// RenderLater returns at render thunk
func (r *Renderer) RenderLater(name, src string) *Thunk {
	return &Thunk{RenderCtx: r, Src: src, Name: name}
}

// Thunk represents a rendered thunk that can be stored and evaluated at a
// future time
type Thunk struct {
	RenderCtx *Renderer
	Src       string
	Name      string
	value     interface{}
}

// Value gets the value from a thunk
func (t *Thunk) Value() (interface{}, error) {
	if t.value != nil {
		return t.value, nil
	}
	t, err := t.eval()
	return t.value, err
}

func (t *Thunk) eval() (*Thunk, error) {
	result, err := t.RenderCtx.Render(t.Name, t.Src)
	if err == nil {
		t.value = result
	}
	return t, err
}

// Available returns true if a value is available, false if ErrUnresolvable, and
// an error on some other error
func (t *Thunk) Available() (bool, error) {
	if t.value != nil {
		return true, nil
	}
	val, err := t.Value()
	if err == nil {
		t.value = val
		return true, nil
	}
	if err == ErrUnresolvable {
		return false, nil
	}
	return false, err
}
