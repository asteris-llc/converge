package resource

import (
	"bytes"
	"errors"
	"fmt"
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
	TaskStatus
	WarningLevel HealthStatusCode
	DisplayLevel *HealthStatusCode
	FailingDeps  map[string]string
}

// HasFailingDeps returns true of FailingDeps is not empty
func (h *HealthStatus) HasFailingDeps() bool {
	return len(h.FailingDeps) > 0
}

// UpgradeWarning will increase the warning level to at least `level`, but will
// not decrease it if it's already higher.
func (h *HealthStatus) UpgradeWarning(level HealthStatusCode) {
	if h.WarningLevel < level {
		h.WarningLevel = level
	}
}

// IsWarning returns true if the warning level is StatusWarning
func (h *HealthStatus) IsWarning() bool {
	return h.WarningLevel == StatusWarning
}

// IsError returns true if the warning level is StatusError
func (h *HealthStatus) IsError() bool {
	return h.WarningLevel == StatusError
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

// Messages returns health status messages
func (h *HealthStatus) Messages() []string {
	var messages []string
	if statusCode := h.TaskStatus.StatusCode(); statusCode != 0 {
		messages = append(messages, fmt.Sprintf("Check returned %d", statusCode))
	}
	if h.TaskStatus.HasChanges() {
		messages = append(messages, "Check indicates changes are required")
	}
	if depCount := len(h.FailingDeps); depCount > 0 {
		messages = append(messages, fmt.Sprintf("%d failing dependencies", depCount))
	}
	return messages
}

// Changes returns changes from the underlying TaskStatus diffs
func (h *HealthStatus) Changes() map[string]Diff {
	return h.Diffs()
}

// HasChanges returns true if the status indicates that there are changes
func (h *HealthStatus) HasChanges() bool {
	return h.TaskStatus.HasChanges()
}

// Error returns nil
func (h *HealthStatus) Error() error {
	var msg bytes.Buffer
	var hasError bool
	if h.ShouldDisplay() {
		msg.WriteString("required changes detected")
		hasError = true
	}
	if len(h.FailingDeps) > 0 {
		hasError = true
		if h.ShouldDisplay() {
			msg.WriteString("; ")
		}
		msg.WriteString("required dependencies are failing")
	}
	if hasError {
		return errors.New(msg.String())
	}
	return nil
}
