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
	"fmt"
	"reflect"

	log "github.com/Sirupsen/logrus"
	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/parse"
	"github.com/asteris-llc/converge/render/extensions"
	"github.com/asteris-llc/converge/render/preprocessor"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/param"
	"github.com/pkg/errors"
)

// ErrUnresolvable is returned by Render if the template string tries to resolve
// unaccesible node properties.
type ErrUnresolvable struct{}

func (ErrUnresolvable) Error() string { return "node is unresolvable" }

// ErrBadTemplate is returned by Render if the template string causes an error.
// It is likely to be returned in cases where a predicate function is rendered
// and does not result in a boolean value.
type ErrBadTemplate struct{ Err error }

func (e ErrBadTemplate) Error() string {
	return fmt.Sprintf("%s: cannot execute template", e.Err)
}

// Renderer to be passed to preparers, which will render strings
type Renderer struct {
	Graph           func() *graph.Graph
	ID              string
	DotValue        resource.Value
	DotValuePresent bool
	resolverErr     bool
	Language        *extensions.LanguageExtension
}

// GetID returns the ID of this renderer
func (r *Renderer) GetID() string {
	return r.ID
}

// Value of this renderer
func (r *Renderer) Value() (value resource.Value, present bool) {
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
		return "", ErrBadTemplate{Err: err}
	}
	return out.String(), err
}

func getNearestAncestor(g *graph.Graph, id, node string) (string, bool) {
	if graph.IsRoot(node) || node == "" {
		return "", false
	}
	siblingID := graph.SiblingID(id, node)
	val, ok := g.Get(siblingID)
	if !ok {
		return getNearestAncestor(g, graph.ParentID(id), node)
	}
	if elem, ok := val.Value().(*parse.Node); ok {
		if elem.Kind() == "module" {
			return "", false
		}
	}
	return siblingID, true
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

	ancestor, found := getNearestAncestor(r.Graph(), r.ID, "param."+name)
	if !found {
		return "", errors.New("param not found (no such ancestor)")
	}
	ancestorMeta, _ := r.Graph().Get(ancestor)
	task, ok := resource.ResolveTask(ancestorMeta.Value())

	if task == nil || !ok {
		return "", errors.New("param not found")
	}

	if _, ok = task.(*PrepareThunk); ok {
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

	vertexName, terms, found := preprocessor.VertexSplitTraverse(g,
		name,
		r.ID,
		preprocessor.TraverseUntilModule,
		make(map[string]struct{}),
	)

	if !validateLookup(g, r.ID, vertexName) {
		return "", fmt.Errorf("%s cannot resolve inner-branch node at %s", r.ID, vertexName)
	}

	if !found {
		return "", fmt.Errorf("%s does not resolve to a valid node", fqgn)
	}

	meta, ok := g.Get(vertexName)
	if !ok {
		return "", fmt.Errorf("%s is empty", vertexName)
	}

	if _, isThunk := meta.Value().(*PrepareThunk); isThunk {
		log.WithField("proxy-reference",
			vertexName,
		).Warn(fmt.Sprintf("%s: cannot resolve %s in node %s in prepare thunk",
			r.ID,
			vertexName, terms,
		))
		r.resolverErr = true
		return "", ErrUnresolvable{}
	}

	if _, isPreparer := meta.Value().(*resource.Preparer); isPreparer {
		log.WithField("preparer-reference", vertexName).Warn("node is unresolvable")
		log.WithField("proxy-reference", vertexName).Warn(fmt.Sprintf("%s: cannot resolve %s in node %s from preparer", r.ID, vertexName, terms))
		r.resolverErr = true
		return "", ErrUnresolvable{}
	}

	asTasker, ok := meta.Value().(resource.Tasker)
	if !ok {
		log.WithField("get-value", vertexName).Error(fmt.Sprintf("%s: lookup would address unevaluated field %s", r.ID, vertexName))
		return "", errors.New("cannot lookup unevaluated field")
	}

	status := asTasker.GetStatus()

	if status == nil {
		log.WithField("status-reference", vertexName).Warn(r.ID + " no status for node " + vertexName)
		r.resolverErr = true
		return "", ErrUnresolvable{}
	}

	result, ok := status.ExportedFields()[terms]

	if !ok {
		var keys []string
		for key := range status.ExportedFields() {
			keys = append(keys, key)
		}
		innerTask, _ := asTasker.GetTask()
		innerTask, _ = resource.ResolveTask(innerTask)
		log.WithField("current-node", r.ID).Warn(fmt.Sprintf("%s is not one of the exported fields for type %T: %v at %s", terms, innerTask, keys, vertexName))
		return "", ErrUnresolvable{}
	}

	return fmt.Sprintf("%v", result), nil
}

// validateLookup ensures that the lookup is valid and resolvable over cases of
// nesting and conditional evaluation.  It restricts lookups such that a nested
// value may depend on an outer value, but an outer value may not depend on a
// nested value.
func validateLookup(g *graph.Graph, src, dst string) bool {
	if g.AreSiblings(src, dst) {
		return true
	}
	if g.IsNibling(src, dst) {
		return false
	}
	return true
}
