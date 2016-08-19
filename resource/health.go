package resource

import (
	"context"

	"github.com/asteris-llc/converge/graph"
)

const (
	// StatusHealthy indicates a passing health check
	StatusHealthy HealthStatusCode = iota
	// StatusWarning indicates a change is needed
	StatusWarning
	// StatusError indicates that a module is non-functional
	StatusError
)

const defaultDisplayLevel = StatusWarning

// HealthStatusCode is a status indicator for health level.  It should be one of:
//   StatusHealth
//   StatusWarning
//   StatusError
type HealthStatusCode int

// HealthStatus contains a status level, display threshold, and status message
type HealthStatus struct {
	WarningLevel HealthStatusCode
	DisplayLevel *HealthStatusCode
	Message      string
}

// NewHealthStatus will return a new health status with a warning level of
// `Healthy` and a DisplayHealthStatus of `StatusWarning`
func NewHealthStatus() *HealthStatus {
	lvl := defaultDisplayLevel
	return &HealthStatus{DisplayLevel: &lvl}
}

// FatalHealthStatus is shorthand for a fatal status case
func FatalHealthStatus(msg string) *HealthStatus {
	return &HealthStatus{WarningLevel: StatusError, Message: msg}
}

// UpgradeWarning will increase the warning level to at least `level`, but will
// not decrease it if it's already higher.
func (h *HealthStatus) UpgradeWarning(level HealthStatusCode) {
	if h.WarningLevel < level {
		h.WarningLevel = level
	}
}

// ShouldDisplay returns true if the warning level is at least the display level
func (h *HealthStatus) ShouldDisplay() bool {
	var threshold HealthStatusCode
	if h.DisplayLevel == nil {
		threshold = defaultDisplayLevel
	} else {
		threshold = *h.DisplayLevel
	}
	return h.WarningLevel >= threshold
}

// Check defines the interface for a health check
type Check interface {
	FailingDep(string, TaskStatus)
	HealthCheck() (*HealthStatus, error)
}

// CheckGraph walks a graph and runs health checks on each health-checkable node
func CheckGraph(ctx context.Context, in *graph.Graph) (*graph.Graph, error) {
	out := graph.New()
	for _, vertexID := range in.Vertices() {
		vertex := in.Get(vertexID)
		check, ok := vertex.(Check)
		if !ok {
			continue
		}

		for _, ancestor := range in.Descendents(vertexID) {
			asTask, ok := in.Get(ancestor).(TaskStatus)
			if !ok {
				continue
			}
			if asTask.Changes() {
				check.FailingDep(ancestor, asTask)
			}
		}
		status, err := check.HealthCheck()
		if err != nil {
			return nil, err
		}
		out.Add(vertexID, status)
	}
	return out, nil
}
