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

	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/render"
)

// ErrTreeContainsErrors is a signal value to indicate errors in the graph
var ErrTreeContainsErrors = errors.New("plan has errors, check graph")

// Plan the execution of a Graph of resource.Tasks
func Plan(ctx context.Context, in *graph.Graph) (*graph.Graph, error) {
	return WithNotify(ctx, in, nil)
}

// WithNotify is plan, but with a notification feature
func WithNotify(ctx context.Context, in *graph.Graph, notify *graph.Notifier) (*graph.Graph, error) {
	var hasErrors error

	renderingPlant, err := render.NewFactory(ctx, in)
	if err != nil {
		return nil, err
	}

	out, err := in.Transform(ctx,
		notify.Transform(func(id string, out *graph.Graph) error {
			renderingPlant.Graph = out

			pipeline := Pipeline(out, id, renderingPlant)

			meta, _ := out.Get(id)

			val, pipelineErr := pipeline.Exec(meta.Value())
			if pipelineErr != nil {
				fmt.Printf("pipeline returned Right %v\n", val)
				return pipelineErr
			}

			asResult, ok := val.(*Result)
			if !ok {
				fmt.Printf("expected *Result but got %T\n", val)
				return fmt.Errorf("expected asResult but got %T", val)
			}

			if nil != asResult.Error() {
				hasErrors = ErrTreeContainsErrors
			}

			out.Add(meta.WithValue(asResult))

			return nil
		}),
	)
	if err != nil {
		return out, err
	}
	return out, hasErrors
}
