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

package apply

import (
	"context"
	"fmt"

	"github.com/asteris-llc/converge/executor"
	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/graph/node"
	"github.com/asteris-llc/converge/plan"
	"github.com/asteris-llc/converge/render"
	"github.com/pkg/errors"
)

// MkPipelineF is a function to generate a pipeline given an id
type MkPipelineF func(*graph.Graph, string) executor.Pipeline

// ErrTreeContainsErrors is a signal value to indicate errors in the graph
var ErrTreeContainsErrors = errors.New("apply had errors, check graph")

// Apply the actions in a Graph of resource.Tasks
func Apply(ctx context.Context, in *graph.Graph) (*graph.Graph, error) {
	renderingPlant, err := render.NewFactory(ctx, in)
	if err != nil {
		return nil, err
	}
	pipeline := func(g *graph.Graph, id string) executor.Pipeline {
		return Pipeline(g, id, renderingPlant)
	}
	return execPipeline(ctx, in, pipeline, renderingPlant, nil)
}

// PlanAndApply plans and applies each node
func PlanAndApply(ctx context.Context, in *graph.Graph) (*graph.Graph, error) {
	return WithNotify(ctx, in, nil)
}

// WithNotify calls PlanAndApply with a notifier
func WithNotify(ctx context.Context, in *graph.Graph, notify *graph.Notifier) (*graph.Graph, error) {
	renderingPlant, err := render.NewFactory(ctx, in)
	if err != nil {
		return nil, err
	}
	pipeline := func(g *graph.Graph, id string) executor.Pipeline {
		return plan.Pipeline(g, id, renderingPlant).Connect(Pipeline(g, id, renderingPlant))
	}
	return execPipeline(ctx, in, pipeline, renderingPlant, notify)
}

// Apply the actions in a Graph of resource.Tasks
func execPipeline(ctx context.Context, in *graph.Graph, pipelineF MkPipelineF, renderingPlant *render.Factory, notify *graph.Notifier) (*graph.Graph, error) {
	var hasErrors error

	out, err := in.Transform(ctx,
		notify.Transform(func(id string, out *graph.Graph) error {
			renderingPlant.Graph = out
			pipeline := pipelineF(out, id)
			val, pipelineError := pipeline.Exec(out.Get(id))
			if pipelineError != nil {
				hasErrors = ErrTreeContainsErrors
				return pipelineError
			}
			asResult, ok := val.(*Result)
			if !ok {
				return fmt.Errorf("expected asResult but got %T", val)
			}

			if nil != asResult.Error() {
				hasErrors = ErrTreeContainsErrors
			}

			out.Add(node.New(id, asResult))
			return nil
		}),
	)

	if err != nil {
		return out, err
	}

	return out, hasErrors
}
