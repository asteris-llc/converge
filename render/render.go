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
	"crypto/rand"
	"fmt"

	"github.com/asteris-llc/converge/executor"
	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/graph/node"
	"github.com/asteris-llc/converge/graph/node/conditional"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/module"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

// Values for rendering
type Values map[string]resource.Value

// Render a graph with the provided values
func Render(ctx context.Context, g *graph.Graph, top Values) (*graph.Graph, error) {
	fmt.Println("beginning to render graph...")
	renderingPlant, err := NewFactory(ctx, g)
	if err != nil {
		return nil, errors.Wrap(err, "render.Render is unable to render graph")
	}
	return g.RootFirstTransform(ctx, func(meta *node.Node, out *graph.Graph) error {
		fmt.Println("generating pipeline")
		pipeline := Pipeline(out, meta.ID, renderingPlant, top)
		value, err := pipeline.Exec(ctx, meta.Value())
		if err != nil {
			fmt.Println("render pipeline failed at ", meta.ID, "with ", err)
			return err
		}
		out.Add(meta.WithValue(value))
		renderingPlant.Graph = out
		return nil
	})
}

type pipelineGen struct {
	Graph          *graph.Graph
	RenderingPlant *Factory
	ID             string
	Top            Values
}

// Pipeline generates a pipelined form of rendering
func Pipeline(g *graph.Graph, id string, factory *Factory, top Values) executor.Pipeline {
	p := pipelineGen{Graph: g, RenderingPlant: factory, Top: top, ID: id}
	return executor.NewPipeline().
		AndThen(p.maybeTransformRoot).
		AndThen(p.prepareNode).
		AndThen(p.wrapTask)
}

// Check to see if the current id is "root", if so generate a new module
// preparer for it and add in all of the command-line parameters; otherwise if
// the node is a valide resource.Resource return it.  If it's not root and not a
// resource.Resource return an error.
func (p pipelineGen) maybeTransformRoot(ctx context.Context, idi interface{}) (interface{}, error) {
	if graph.IsRoot(p.ID) {
		return module.NewPreparer(p.Top), nil
	}
	if res, ok := idi.(resource.Resource); ok {
		return res, nil
	}
	return nil, typeError("render.maybeTransformRoot", p.ID, "resource.Resource", idi)
}

// Run prepare on the node and return the resource.Resource to be wrapped
func (p pipelineGen) prepareNode(ctx context.Context, idi interface{}) (interface{}, error) {
	var metadataErr error

	res, ok := idi.(resource.Resource)
	if !ok {
		return nil, typeError("render.prepareNode", p.ID, "resource.Resource", idi)
	}

	renderer, renderErr := p.RenderingPlant.GetRenderer(p.ID)

	if p.shouldRenderMetadata() {
		_, metadataErr = p.renderMetadata(renderer)
	}

	merged := mergeMaybeUnresolvables(metadataErr, renderErr)

	var prepared interface{}
	var err error
	if renderer != nil {
		prepared, err = res.Prepare(ctx, renderer)
		merged = mergeMaybeUnresolvables(merged, err)
	}

	if merged != nil {
		if errIsUnresolvable(merged) {
			return createThunk(func(factory *Factory) (resource.Task, error) {
				dynamicRenderer, rendErr := factory.GetRenderer(p.ID)
				if rendErr != nil {
					return nil, rendErr
				}
				if p.shouldRenderMetadata() {
					_, rendErr := p.renderMetadata(dynamicRenderer)
					if rendErr != nil {
						return nil, rendErr
					}
				}
				return res.Prepare(ctx, dynamicRenderer)
			}), nil
		}
		return nil, merged
	}
	return prepared, nil
}

func mergeMaybeUnresolvables(err1, err2 error) error {
	if err1 == nil {
		return err2
	}
	if err2 == nil {
		return err1
	}
	if errIsUnresolvable(err1) && errIsUnresolvable(err2) {
		return err1
	}
	if !errIsUnresolvable(err1) {
		return err1
	}
	if !errIsUnresolvable(err2) {
		return err2
	}
	return multierror.Append(err1, err2)
}

func errIsUnresolvable(err error) bool {
	_, ok := errors.Cause(err).(ErrUnresolvable)
	return ok
}

func errIsBadTemplate(err error) bool {
	_, ok := errors.Cause(err).(ErrBadTemplate)
	return ok
}

func (p pipelineGen) shouldRenderMetadata() bool {
	meta, ok := p.Graph.Get(p.ID)
	if !ok {
		return false
	}
	return conditional.IsConditional(meta)
}

func (p pipelineGen) renderMetadata(r *Renderer) (string, error) {
	meta, ok := p.Graph.Get(p.ID)
	if !ok {
		return "", errors.New(p.ID + " does not exist in graph, cannot render metadata")
	}
	if !conditional.IsConditional(meta) {
		return "", nil
	}
	return conditional.RenderPredicate(meta, r.Render)
}

// Takes a resource.Task and wraps it in resource.TaskWrapper
func (p pipelineGen) wrapTask(ctx context.Context, taski interface{}) (interface{}, error) {
	if task, ok := taski.(*PrepareThunk); ok {
		return task, nil
	}
	if task, ok := taski.(resource.Task); ok {
		return resource.WrapTask(task), nil
	}
	return nil, typeError("render.wrapTask", p.ID, "resource.Task", taski)
}

func typeError(where, what, expected string, actual interface{}) error {
	return fmt.Errorf("type error in %s: expected %s to be type %s but received type %T", where, what, expected, actual)
}

// PrepareThunk returns a possibly lazily evaluated preparer
type PrepareThunk struct {
	// prevent hashing thunks into a single value
	Data  []byte
	Thunk func(*Factory) (resource.Task, error) `hash:"ignore"`
}

// Check allows thunk to implement resource.Task
func (p *PrepareThunk) Check(context.Context, resource.Renderer) (resource.TaskStatus, error) {
	return nil, errors.New("Unresolved thunk: cannot be evaluated")
}

// Apply allows thunk to implement resource.Task
func (p *PrepareThunk) Apply(context.Context) (resource.TaskStatus, error) {
	return nil, errors.New("Unresolved thunk: cannot be evaluated")
}

func createThunk(f func(*Factory) (resource.Task, error)) *PrepareThunk {
	junk := make([]byte, 32)
	rand.Read(junk)
	return &PrepareThunk{
		Thunk: f,
		Data:  junk,
	}
}
