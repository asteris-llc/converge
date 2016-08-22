package resource

import (
	"context"

	"github.com/asteris-llc/converge/graph"
)

// Check defines the interface for a health check
type Check interface {
	FailingDep(string, TaskStatus)
	HealthCheck() (*HealthStatus, error)
}

// CheckGraph walks a graph and runs health checks on each health-checkable node
func CheckGraph(ctx context.Context, in *graph.Graph) (*graph.Graph, error) {
	return in.Transform(ctx, func(id string, out *graph.Graph) error {
		asCheck, ok := out.Get(id).(Check)
		if !ok {
			return nil
		}
		for _, dep := range out.Dependencies(id) {
			depStatus, ok := out.Get(dep).(TaskStatus)
			if !ok {
				continue
			}
			if isFailingStatus(depStatus) {
				asCheck.FailingDep(dep, depStatus)
			}
		}
		out.Add(id, asCheck)
		return nil
	})
}

func isFailingStatus(stat TaskStatus) bool {
	if check, ok := stat.(Check); ok {
		return healthCheckOK(check)
	}
	return stat.Changes()
}

func healthCheckOK(c Check) bool {
	status, err := c.HealthCheck()
	if err != nil {
		return false
	}
	return status.ShouldDisplay()
}
