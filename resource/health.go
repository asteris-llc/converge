package resource

import (
	"context"
	"errors"

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
		task, err := unboxNode(out.Get(id))
		if err != nil {
			return err
		}
		asCheck, ok := task.(Check)
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
		status, err := asCheck.HealthCheck()
		if err != nil {
			return err
		}
		out.Add(id, status)
		return nil
	})
}

func unboxNode(i interface{}) (TaskStatus, error) {
	type statusWrapper interface {
		GetStatus() TaskStatus
	}

	if wrapper, ok := i.(statusWrapper); ok {
		return wrapper.GetStatus(), nil
	} else if taskStatus, ok := i.(TaskStatus); ok {
		return taskStatus, nil
	}
	return nil, errors.New("[FATAL] cannot get task status from node")
}

func isFailingStatus(stat TaskStatus) bool {
	if check, ok := stat.(Check); ok {
		return healthCheckOK(check)
	}
	return stat.HasChanges()
}

func healthCheckOK(c Check) bool {
	status, err := c.HealthCheck()
	if err != nil {
		return false
	}
	return status.ShouldDisplay()
}
