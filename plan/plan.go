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

package plan

import (
	"context"
	"errors"
	"fmt"

	"github.com/asteris-llc/converge/executor/either"
	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/render"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/module"
)

// ErrTreeContainsErrors is a signal value to indicate errors in the graph
var ErrTreeContainsErrors = errors.New("plan has errors, check graph")

// Plan the execution of a Graph of resource.Tasks
func Plan(ctx context.Context, in *graph.Graph, params render.Values) (*graph.Graph, error) {
	var hasErrors error

	if err := ensureRootModule(ctx, in, params); err != nil {
		return nil, err
	}

	renderingPlant, err := render.NewFactory(ctx, in)
	if err != nil {
		return nil, err
	}

	out, err := in.Transform(ctx, func(id string, out *graph.Graph) error {
		renderingPlant.Graph = out

		pipeline := render.Pipeline(ctx, out, id, params).Connect(Pipeline(out, id, renderingPlant))
		result := pipeline.Exec(either.ReturnM(out.Get(id)))
		val, isRight := result.FromEither()
		if !isRight {
			fmt.Printf("pipeline returned Right %v\n", val)
			return fmt.Errorf("%v", val)
		}

		asResult, ok := val.(*Result)
		if !ok {
			fmt.Printf("expected *Result but got %T\n", val)
			return fmt.Errorf("expected asResult but got %T", val)
		}

		if nil != asResult.Error() {
			hasErrors = ErrTreeContainsErrors
		}

		out.Add(id, asResult)
		return nil
	})

	if err != nil {
		return out, err
	}

	return out, hasErrors
}

func ensureRootModule(ctx context.Context, g *graph.Graph, params render.Values) error {
	grRoot, err := g.Root()
	if err != nil {
		return err
	}
	if grRoot != "root" {
		fmt.Printf("[INFO] root node is '%s' not 'root', skipping\n", grRoot)
		return nil
	}
	rootPreparer := module.NewPreparer(params)
	renderingPlant, err := render.NewFactory(ctx, g)
	if err != nil {
		return err
	}
	renderer, err := renderingPlant.GetRenderer(grRoot)
	if err != nil {
		return err
	}
	rootTask, err := rootPreparer.Prepare(renderer)
	if err != nil {
		return err
	}
	wrapped := resource.WrapTask(rootTask)
	g.Add(grRoot, wrapped)
	return nil
}
