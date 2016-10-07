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
	"reflect"

	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/render/extensions"
	"github.com/asteris-llc/converge/render/preprocessor"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/param"
)

// ErrUnresolvable is returned by Render if the template string tries to resolve
// unaccesible node properties.
type ErrUnresolvable struct{}

func (ErrUnresolvable) Error() string { return "node is unresolvable" }

// Renderer to be passed to preparers, which will render strings
type Renderer struct {
	Graph           func() *graph.Graph
	ID              string
	DotValue        string
	DotValuePresent bool
	resolverErr     bool
	Language        *extensions.LanguageExtension
}

// GetID returns the ID of this renderer
func (r *Renderer) GetID() string {
	return r.ID
}

// Value of this renderer
func (r *Renderer) Value() (value string, present bool) {
	return r.DotValue, r.DotValuePresent
}

// Render a string with text/template
func (r *Renderer) Render(name, src string) (string, error) {
	r.resolverErr = false

	r.Language = r.Language.On("param", r.param)
	r.Language = r.Language.On("paramList", r.paramList)
	r.Language = r.Language.On("paramMap", r.paramMap)

	r.Language = r.Language.On(extensions.RefFuncName, r.lookup)
	out, err := r.Language.Render(r.DotValue, name, src)
	if err != nil {
		if r.resolverErr {
			return "", ErrUnresolvable{}
		}
		return "", err
	}
	return out.String(), err
}

func (r *Renderer) param(name string) (string, error) {
	raw, err := r.paramRawValue(name)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%v", raw), nil
}

func (r *Renderer) paramList(name string) ([]string, error) {
	raw, err := r.paramRawValue(name)
	if err != nil {
		return nil, err
	}

	vals := reflect.ValueOf(raw)
	if vals.Kind() != reflect.Slice {
		return nil, fmt.Errorf("param is not a list, it is a %s (%s)", vals.Kind(), vals)
	}

	var out []string
	for i := 0; i < vals.Len(); i++ {
		val := vals.Index(i)
		out = append(out, fmt.Sprintf("%v", val.Interface()))
	}

	return out, nil
}

func (r *Renderer) paramMap(name string) (map[string]interface{}, error) {
	raw, err := r.paramRawValue(name)
	if err != nil {
		return nil, err
	}

	vals := reflect.ValueOf(raw)
	if vals.Kind() != reflect.Map {
		return nil, fmt.Errorf("param is not a map, it is a %s (%s)", vals.Kind(), vals)
	}

	out := map[string]interface{}{}
	for _, key := range vals.MapKeys() {
		k := fmt.Sprintf("%v", key)

		out[k] = vals.MapIndex(key).Interface()
	}

	return out, nil
}

func (r *Renderer) paramRawValue(name string) (interface{}, error) {
	task, ok := resource.ResolveTask(r.Graph().Get(graph.SiblingID(r.ID, "param."+name)))

	if task == nil || !ok {
		return "", errors.New("param not found")
	}

	if _, ok := task.(*PrepareThunk); ok {
		r.resolverErr = true
		return "", ErrUnresolvable{}
	}

	// grab the value
	param, ok := task.(*param.Param)
	if !ok {
		return nil, fmt.Errorf("task it not a param, but a %T", task)
	}

	return param.Val, nil
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
		return "", ErrUnresolvable{}
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
			return "", ErrUnresolvable{}
		}
		return "", err
	}

	return fmt.Sprintf("%v", result), nil
}
