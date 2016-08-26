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

package resource

import "fmt"

const (
	// StatusNoChange means no changes are necessary
	StatusNoChange int = 0

	// StatusWontChange indicates an acceptable delta that wont be corrected
	StatusWontChange int = iota

	// StatusWillChange indicates an unacceptable delta that will be corrected
	StatusWillChange

	// StatusFatal indicates an unacceptable delta that cannot be corrected
	StatusFatal
)

type badDep struct {
	ID     string
	Status TaskStatus
}

// TaskStatus represents the results of Check called during planning or
// application.
type TaskStatus interface {
	Value() string
	Diffs() map[string]Diff
	StatusCode() int
	Messages() []string
	HasChanges() bool
}

// Status is the default TaskStatus implementation
type Status struct {
	Differences  map[string]Diff
	WarningLevel int
	Output       []string
	WillChange   bool
	Status       string
	FailingDeps  []badDep
}

// Value returns the status value
func (t *Status) Value() string {
	return t.Status
}

// Diffs returns the internal differences
func (t *Status) Diffs() map[string]Diff {
	return t.Differences
}

// StatusCode returns the current warning level
func (t *Status) StatusCode() int {
	return t.WarningLevel
}

// Messages returns the current outpt slice
func (t *Status) Messages() []string {
	return t.Output
}

// HasChanges returns the WillChange value
func (t *Status) HasChanges() bool {
	return t.WillChange
}

// HealthCheck provides a default health check implementation for statuses
func (t *Status) HealthCheck() (status *HealthStatus, err error) {
	status = &HealthStatus{TaskStatus: t, FailingDeps: make(map[string]string)}
	if !t.HasChanges() && len(t.FailingDeps) == 0 {
		return
	}

	// There are changes or failing dependencies so the health check is at least
	// at a warning status.
	status.UpgradeWarning(StatusWarning)

	for _, failingDep := range t.FailingDeps {
		var depMessage string
		if msg := failingDep.Status.Value(); msg != "" {
			depMessage = msg
		} else {
			depMessage = fmt.Sprintf("returned %d", failingDep.Status.StatusCode())
		}

		status.FailingDeps[failingDep.ID] = depMessage
	}
	if t.StatusCode() >= 2 {
		status.UpgradeWarning(StatusError)
	}
	return
}

// FailingDep tracks a new failing dependency
func (t *Status) FailingDep(id string, stat TaskStatus) {
	t.FailingDeps = append(t.FailingDeps, badDep{ID: id, Status: stat})
}

// AddDifference adds a TextDiff to the Differences map
func (t *Status) AddDifference(name, original, current, defaultVal string) {
	t.Differences = AddTextDiff(t.Differences, name, original, current, defaultVal)
}

// Merge takes the current status and adds on any additional messages from another
func (t *Status) Merge(next *Status) {
	// first merge differences
	for key, diff := range next.Differences {
		if _, ok := t.Differences[key]; !ok {
			t.Differences[key] = diff
		}
	}
	// Next Merge WarningLevel such that the higher number takes precedence.
	// This way Willchange overrides Won't change
	t.WarningLevel = max(t.WarningLevel, next.WarningLevel)
	// Merge the Outputs
	t.Output = append(t.Output, next.Output...)
	// Or willchange
	t.WillChange = t.WillChange || next.WillChange
}

// max is used solely for the above function
func max(a, b int) int {
	if b > a {
		return b
	}
	return a
}

// Diff represents a difference
type Diff interface {
	Original() string
	Current() string
	Changes() bool
}

// TextDiff is the default Diff implementation
type TextDiff struct {
	Default string
	Values  [2]string
}

// Original returns the unmodified value of the diff
func (t TextDiff) Original() string {
	if t.Values[0] == "" {
		return t.Default
	}
	return t.Values[0]
}

// Current returns the modified value of the diff
func (t TextDiff) Current() string {
	if t.Values[1] == "" {
		return t.Default
	}
	return t.Values[1]
}

// Changes is true if the Original and Current values differ
func (t TextDiff) Changes() bool {
	return t.Values[0] != t.Values[1]
}

// AnyChanges takes a diff map and returns true if any of the diffs in the map
// have changes.
func AnyChanges(diffs map[string]Diff) bool {
	for _, diffIf := range diffs {
		diff, ok := diffIf.(Diff)
		if !ok {
			panic("invalid conversion")
		}
		if diff.Changes() {
			return true
		}
	}
	return false
}

// AddTextDiff inserts a new TextDiff into a map of names to Diffs
func AddTextDiff(m map[string]Diff, name, original, current, defaultVal string) map[string]Diff {
	if m == nil {
		m = make(map[string]Diff)
	}
	m[name] = TextDiff{Values: [2]string{original, current}, Default: defaultVal}
	return m
}
