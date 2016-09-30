package load

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
	logger.Debug("resolving conditional macros")
	return g.Transform(ctx, func(id string, out *graph.Graph) error {
		switchNode, ok := g.Get(id).(*control.SwitchPreparer)
		if !ok {
			return nil
		}
		for _, edge := range g.DownEdges(id) {
			caseID, ok := edge.Target().(string)
			if !ok {
				logger.Error("graph node was not a string as expected")
				return errors.New("invalid node")
			}
			caseNode, ok := g.Get(caseID).(*control.CasePreparer)
			if !ok {
				continue
			}
			switchNode.Cases = append(switchNode.Cases, caseNode)
			for _, caseEdge := range g.DownEdges(edge.Target().(string)) {
				targetID, ok := caseEdge.Target().(string)
				if !ok {
					logger.Error("graph node was not a string as expected")
					return errors.New("invalid node")
				}
				conditionalTarget, ok := g.Get(targetID).(resource.Resource)
				if !ok {
					logger.Infof(
						"unexpected type for node at %s: %T",
						caseEdge.Target().(string),
						g.Get(caseEdge.Target().(string)),
					)
				}
				conditional := &control.ConditionalPreparer{
					Resource: conditionalTarget,
					ShouldEvaluate: func() bool {
						return caseNode.ShouldEvaluate()
					},
				}
				out.Add(targetID, conditional)
			}
		}
		return nil
	})
}
