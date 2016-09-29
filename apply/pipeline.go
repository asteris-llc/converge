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
	"errors"
	"fmt"

	"github.com/asteris-llc/converge/executor"
	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/plan"
	"github.com/asteris-llc/converge/render"
	"github.com/asteris-llc/converge/resource"
)

type pipelineGen struct {
	Graph          *graph.Graph
	ID             string
	RenderingPlant *render.Factory
}

type resultWrapper struct {
	Plan *plan.Result
}

// Pipeline generates a pipeline to evaluate a single graph node
func Pipeline(g *graph.Graph, id string, factory *render.Factory) executor.Pipeline {
	gen := &pipelineGen{Graph: g, RenderingPlant: factory, ID: id}
	return executor.NewPipeline().
		AndThen(gen.GetTask).
		AndThen(gen.DependencyCheck).
		AndThen(gen.maybeSkipApplication).
		AndThen(gen.applyNode).
		AndThen(gen.maybeRunFinalCheck)
}

// GetResult returns Right resultWrapper if the value is a *plan.Result, or Left
// Error if not
func (g *pipelineGen) GetTask(idi interface{}) (interface{}, error) {
	if plan, ok := idi.(*plan.Result); ok {
		return resultWrapper{Plan: plan}, nil
	}
	return nil, fmt.Errorf("expected plan.Result but got %T", idi)
}

// DependencyCheck looks for failing dependency nodes.  If an error is
// encountered it returns `Left error`, if failing dependencies are encountered
// it returns `Right (Left apply.Result)` and otherwise returns `Right (Right
// plan.Result)`. The return values are structured to short-circuit `PlanNode`
// if we have failures.
func (g *pipelineGen) DependencyCheck(taskI interface{}) (interface{}, error) {
	result, ok := taskI.(resultWrapper)
	if !ok {
		return nil, errors.New("input node is not a task wrapper")
	}
	for _, depID := range graph.Targets(g.Graph.DownEdges(g.ID)) {
		elem := g.Graph.Get(depID)
		dep, ok := elem.(executor.Status)
		if !ok {
			return nil, fmt.Errorf("apply.DependencyCheck: expected %s to have type executor.Status but got type %T", depID, elem)
		}
		if err := dep.Error(); err != nil {
			errResult := &Result{
				Ran:    false,
				Status: &resource.Status{Level: resource.StatusWillChange},
				Err:    fmt.Errorf("error in dependency %q", depID),
			}
			return errResult, nil
		}
	}
	return result, nil
}

// maybeSkipAppliation will return a result if it's given a *Result, if it's
// given a taskWrapper it will return a result if there are no changes,
// otherwise it returns the taskWrapper
func (g *pipelineGen) maybeSkipApplication(resultI interface{}) (interface{}, error) {
	if asResult, ok := resultI.(*Result); ok {
		return asResult, nil
	}
	asPlan, ok := resultI.(resultWrapper)
	if !ok {
		return nil, fmt.Errorf("expected *Result or *resultWrapper but got type %T", resultI)
	}
	if !asPlan.Plan.Status.HasChanges() {
		return &Result{
			Ran:  false,
			Task: asPlan.Plan.Task,
			Plan: asPlan.Plan,
			Err:  asPlan.Plan.Err,
		}, nil
	}
	return asPlan, nil
}

// applyNode runs apply on the node, it takes an Either *apply.Result
// *plan.Result and, if the input value is Left, returns it as a Right value,
// otherwise it attempts to run apply on the *plan.Result.Task and returns an
// appropriate Left or Right value.
func (g *pipelineGen) applyNode(val interface{}) (interface{}, error) {
	if asResult, ok := val.(*Result); ok {
		return asResult, nil
	}
	twrapper, ok := val.(resultWrapper)
	if !ok {
		return nil, fmt.Errorf("apply expected a resultWrappert but got %T", val)
	}
	applyStatus, err := twrapper.Plan.Task.Apply()
	if err != nil {
		err = fmt.Errorf("error applying %s: %s", g.ID, err)
	}
	return &Result{
		Ran:    true,
		Status: applyStatus,
		Task:   twrapper.Plan.Task,
		Plan:   twrapper.Plan,
		Err:    err,
	}, nil
}

// maybeRunFinalCheck :: *Result -> Either error *Result; looks to see if the
// current result ran, and if so it re-runs plan and sets PostCheck to the
// resulting status.
func (g *pipelineGen) maybeRunFinalCheck(resultI interface{}) (interface{}, error) {
	result, ok := resultI.(*Result)
	if !ok {
		return nil, fmt.Errorf("expected *Result but got %T", resultI)
	}
	if !result.Ran {
		return result, nil
	}
	task := result.Plan.Task
	val, pipelineError := plan.Pipeline(g.Graph, g.ID, g.RenderingPlant).Exec(task)
	if pipelineError != nil {
		return nil, pipelineError
	}
	planned, ok := val.(*plan.Result)
	if !ok {
		return nil, fmt.Errorf("expected *plan.Result but got %T", val)
	}
	result.PostCheck = planned.Status
	if planned.HasChanges() {
		result.Err = fmt.Errorf("%s still has changes after apply", g.ID)
	}
	return result, nil
}

func (g *pipelineGen) Renderer(id string) (*render.Renderer, error) {
	return g.RenderingPlant.GetRenderer(id)
}
