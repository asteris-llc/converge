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

package transform

import (
	"github.com/asteris-llc/converge/graph"
	"golang.org/x/net/context"
)

func ResolveConditionals(ctx context.Context, g *graph.Graph) (*graph.Graph, error) {
	return g, nil
}

/*
// ResolveConditionals will walk the graph and wrap tasks whose parent is a case
// in a conditional resource.
func ResolveConditionals(ctx context.Context, g *graph.Graph) (*graph.Graph, error) {
	logger := logging.GetLogger(ctx).WithField("function", "ResolveConditionals")
	logger.Debug("resolving conditional macros")
	return g.Transform(ctx, func(meta *node.Node, out *graph.Graph) error {
		id := meta.ID
		switchNode, ok := getSwitchNode(id, g)
		if !ok {
			return nil
		}
		for _, caseID := range g.Children(id) {
			caseNode, ok := getCaseNode(caseID, g)
			if caseNode == nil {
				return errors.New("got a nil caseNode for " + id)
			}
			if !ok {
				continue
			}
			switchNode.AppendCase(caseNode)
			for _, targetID := range g.Children(caseID) {
				targetPreparerMeta, ok := g.Get(targetID)
				if !ok {
					continue
				}

				if childThunk, ok := targetPreparerMeta.Value().(*render.PrepareThunk); ok {

					// We're adding a random 32-bit byte slice here ensures that we don't
					// get hash collisions in the PrepareThunk during duplicate merging.
					// This is necessary because we can't hash on the Thunk function, and
					// the Task field is typically a stub value that would overlap for
					// non-mergable nodes.
					junk := make([]byte, 32)
					rand.Read(junk)
					thunkPreparer := &render.PrepareThunk{
						Task: childThunk.Task,
						Data: junk,
						Thunk: func(r *render.Factory) (resource.Task, error) {
							if !caseNode.ShouldEvaluate() {
								return faketask.NoOp(), nil
							}
							innerThunk, err := childThunk.Thunk(r)
							if err != nil {
								return nil, err
							}
							cTask := &control.ConditionalTask{
								Name: targetID,
								Task: innerThunk,
							}
							cTask.SetExecutionController(caseNode)
							return cTask, nil
						},
					}
					out.Add(targetPreparerMeta.WithValue(thunkPreparer))
					continue
				}

				targetPreparer, ok := targetPreparerMeta.Value().(resource.Task)
				if !ok {
					continue
				}
				conditionalTarget := targetPreparer
				conditional := &control.ConditionalTask{
					Name: targetID,
					Task: conditionalTarget,
				}
				conditional.SetExecutionController(caseNode)
				out.Add(targetPreparerMeta.WithValue(conditional))
			}
		}
		switchNode.SortCases()
		return nil
	})
}

func getSwitchNode(id string, g *graph.Graph) (*control.SwitchTask, bool) {
	elemMeta, ok := g.Get(id)
	if !ok {
		return nil, false
	}
	elem := elemMeta.Value()
	elem, canResolve := resource.ResolveTask(elem)
	if !canResolve {
		return nil, false
	}
	if asSwitch, ok := elem.(*control.SwitchTask); ok {
		return asSwitch, true
	}
	return nil, false
}

func getCaseNode(id string, g *graph.Graph) (*control.CaseTask, bool) {
	elemMeta, ok := g.Get(id)
	if !ok {
		return nil, false
	}

	elem := elemMeta.Value()

	elem, canResolve := resource.ResolveTask(elem)
	if !canResolve {
		return nil, false
	}
	if asCase, ok := elem.(*control.CaseTask); ok {
		return asCase, true
	}
	return nil, false
}
*/
