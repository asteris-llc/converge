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
	"context"
	"fmt"

	"github.com/asteris-llc/converge/executor"
	"github.com/asteris-llc/converge/executor/either"
	"github.com/asteris-llc/converge/executor/monad"
	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/module"
)

// Values for rendering
type Values map[string]interface{}

// Render a graph with the provided values
func Render(ctx context.Context, g *graph.Graph, top Values) (*graph.Graph, error) {
	renderingPlant, err := NewFactory(ctx, g)
	if err != nil {
		return nil, err
	}
	return g.RootFirstTransform(ctx, func(id string, out *graph.Graph) error {
		pipeline := Pipeline(out, id, renderingPlant, top)
		result := pipeline.Exec(either.ReturnM(out.Get(id)))
		value, isRight := result.FromEither()
		if !isRight {
			return fmt.Errorf("%v", value)
		}
		out.Add(id, value)
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
func (p pipelineGen) maybeTransformRoot(idi interface{}) monad.Monad {
	if p.ID == "root" {
		return either.RightM(module.NewPreparer(p.Top))
	}
	if res, ok := idi.(resource.Resource); ok {
		return either.RightM(res)
	}
	return either.LeftM(typeError("resource.Renderer", idi))
}

// Run prepare on the node and return the resource.Resource to be wrapped
func (p pipelineGen) prepareNode(idi interface{}) monad.Monad {
	res, ok := idi.(resource.Resource)
	if !ok {
		return either.LeftM(typeError("resource.Resource", idi))
	}
	renderer, err := p.RenderingPlant.GetRenderer(p.ID)
	if err != nil {
		return either.LeftM(err)
	}
	prepared, err := res.Prepare(renderer)
	if err != nil {
		return either.LeftM(err)
	}
	return either.RightM(prepared)
}

// Takes a resource.Task and wraps it in resource.TaskWrapper
func (p pipelineGen) wrapTask(taski interface{}) monad.Monad {
	if task, ok := taski.(resource.Task); ok {
		return either.RightM(resource.WrapTask(task))
	}
	return either.LeftM(typeError("resource.Task", taski))
}

func typeError(expected string, actual interface{}) error {
	return fmt.Errorf("type error: expected type %s but received type %T", expected, actual)
}
