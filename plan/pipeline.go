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
	"errors"
	"fmt"
	"strings"

	"github.com/asteris-llc/converge/executor"
	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/parse/preprocessor/switch"
	"github.com/asteris-llc/converge/render"
	"github.com/asteris-llc/converge/resource"
	"golang.org/x/net/context"
)

type pipelineGen struct {
	Graph          *graph.Graph
	RenderingPlant *render.Factory
	ID             string
}

type taskWrapper struct {
	Task resource.Task
}

// Pipeline generates a pipeline to evaluate a single graph node
func Pipeline(ctx context.Context, g *graph.Graph, id string, factory *render.Factory) executor.Pipeline {
	gen := &pipelineGen{Graph: g, RenderingPlant: factory, ID: id}
	return executor.NewPipeline().
		AndThen(gen.GetTask).
		AndThen(gen.DependencyCheck).
		AndThen(gen.PlanNode)
}

// GetTask returns Right Task if the value is a task, or Left Error if not
func (g *pipelineGen) GetTask(ctx context.Context, idi interface{}) (interface{}, error) {
	if thunk, ok := idi.(*render.PrepareThunk); ok {
		thunked, err := thunk.Thunk(g.RenderingPlant)
		if err != nil {
			return nil, err
		}
		return g.GetTask(ctx, thunked)
	}

	if task, ok := idi.(resource.Task); ok {
		return taskWrapper{Task: task}, nil
	}

	return nil, fmt.Errorf("expected resource.Task but got %T", idi)
}

// MaybeShortcircuitConditional evaluates the conditional predicate in the node,
// if one exists, and then replaces the node with a Nop if it should not be
// evaluated.
func (g *pipelineGen) MaybeShortcircuitConditional(idi interface{}) (interface{}, error) {
	meta, ok := g.Graph.Get(g.ID)
	if !ok {
		return idi, nil
	}
	predicateUnmarshalled, ok := meta.LookupMetadata("conditional-predicate-rendered")
	if !ok {
		return idi, nil
	}

	predicate, ok := predicateUnmarshalled.(string)

	if !ok {
		return nil, fmt.Errorf("%s: predicate should have string type but was %T", g.ID, predicateUnmarshalled)
	}

	if !parseTruth(predicate) {
		return taskWrapper{Task: &control.NopTask{}}, nil
	}

	return idi, nil
}

func parseTruth(predicate string) bool {
	switch strings.ToLower(predicate) {
	case "t", "true":
		return true
	}
	return false
}

// ShouldEvaluate returns true if the node is unconditional, or if it is
func (g *pipelineGen) ShouldEvaluate() bool {
	return true
}

// DependencyCheck looks for failing dependency nodes.  If an error is
// encountered it returns `Left error`, if failing dependencies are encountered
// it returns `Right (Left Status)` and otherwise returns `Right (Right
// Task)`. The return values are structured to short-circuit `PlanNode` if we
// have failures.
func (g *pipelineGen) DependencyCheck(ctx context.Context, taskI interface{}) (interface{}, error) {
	task, ok := taskI.(taskWrapper)
	if !ok {
		return nil, errors.New("input node is not a task wrapper")
	}
	for _, depID := range graph.Targets(g.Graph.DownEdges(g.ID)) {
		meta, ok := g.Graph.Get(depID)
		if !ok {
			return nil, nil
		}

		dep, ok := meta.Value().(executor.Status)
		if !ok {
			return nil, fmt.Errorf("expected executor.Status but got %T", meta.Value())
		}
		if err := dep.Error(); err != nil {
			errResult := &Result{
				Status: &resource.Status{Level: resource.StatusWillChange},
				Task:   task.Task,
				Err:    fmt.Errorf("error in dependency %q", depID),
			}
			return errResult, nil
		}
	}
	return task, nil
}

// PlanNode runs plan on the node, it takes an Either *Result TaskWrapper and,
// if the input value is Left, returns it as a Right value, otherwise it
// attempts to run plan on the TaskWrapper and returns an appropriate Left or
// Right value.
func (g *pipelineGen) PlanNode(ctx context.Context, taski interface{}) (interface{}, error) {
	twrapper, ok := taski.(taskWrapper)
	if !ok {
		asResult, ok := taski.(*Result)
		if ok {
			return asResult, nil
		}
		return nil, fmt.Errorf("expected type *Result or taskWrapper but got %T", taski)
	}

	renderer, err := g.Renderer(g.ID)
	if err != nil {
		return nil, fmt.Errorf("unable to get renderer for %s", g.ID)
	}
	status, err := twrapper.Task.Check(ctx, renderer)

	// create empty Status structure, if it not created in .Check()
	if status == nil {
		status = &resource.Status{}
	}

	type settable interface {
		SetError(error)
	}
	if inner, ok := status.(settable); ok && err != nil {
		inner.SetError(err)
	}

	return &Result{
		Status: status,
		Task:   twrapper.Task,
		Err:    status.Error(),
	}, nil
}

func (g *pipelineGen) Renderer(id string) (*render.Renderer, error) {
	return g.RenderingPlant.GetRenderer(id)
}
