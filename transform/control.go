package transform

import (
	"context"
	"errors"

	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/helpers/logging"
	"github.com/asteris-llc/converge/render/preprocessor/control"
	"github.com/asteris-llc/converge/resource"
)

// ResolveConditionals will walk the graph and wrap tasks whose parent is a case
// in a conditional resource.  For cases it will look at the parent switch and
func ResolveConditionals(ctx context.Context, g *graph.Graph) (*graph.Graph, error) {
	logger := logging.GetLogger(ctx).WithField("function", "ResolveConditionals")
	logger.Info("resolving conditional macros")
	return g.Transform(ctx, func(id string, out *graph.Graph) error {
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
				targetPreparer := g.Get(targetID).(resource.Task)
				conditionalTarget := targetPreparer
				conditional := &control.ConditionalTask{
					Name: targetID,
					Task: conditionalTarget,
				}
				conditional.SetExecutionController(caseNode)
				out.Add(targetID, conditional)
			}
		}
		return nil
	})
}

func getSwitchNode(id string, g *graph.Graph) (*control.SwitchTask, bool) {
	elem := g.Get(id)
	if elem == nil {
		return nil, false
	}
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
	elem := g.Get(id)
	if elem == nil {
		return nil, false
	}
	elem, canResolve := resource.ResolveTask(elem)
	if !canResolve {
		return nil, false
	}
	if asCase, ok := elem.(*control.CaseTask); ok {
		return asCase, true
	}
	return nil, false
}
