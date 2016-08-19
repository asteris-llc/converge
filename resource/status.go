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

import (
	"bytes"
	"fmt"
)

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
	Changes() bool
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

func convertMap(t Diff) Diff {
	return t
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

// Changes returns the WillChange value
func (t *Status) Changes() bool {
	return t.WillChange
}

// HealthCheck provides a default health check implementation for statuses
func (t *Status) HealthCheck() (status *HealthStatus, err error) {
	status = new(HealthStatus)
	var diffBuffer bytes.Buffer
	if !t.Changes() {
		return
	}
	for diffName, diff := range t.Differences {
		diffBuffer.WriteString(fmt.Sprintf("%s: %s => %s\n", diffName, diff.Original(), diff.Current()))
	}
	for _, dep := range t.FailingDeps {
		diffBuffer.WriteString(fmt.Sprintf("failing dependency: %s (%s)\n", dep.ID, dep.Status.Value()))
	}
	if len(t.FailingDeps) > 0 {
		status.UpgradeWarning(StatusWarning)
	}
	status.Message = diffBuffer.String()
	if t.StatusCode() < 2 {
		status.UpgradeWarning(StatusWarning)
	} else {
		status.UpgradeWarning(StatusError)
	}
	return
}

// FailingDep tracks a new failing dependency
func (t *Status) FailingDep(id string, stat TaskStatus) {
	t.FailingDeps = append(t.FailingDeps, badDep{ID: id, Status: stat})
}

// AddDifference adds a TextDiff to the Differences map
func (t *Status) AddDifference(name, current, original string) {
	return
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
func AddTextDiff(m map[string]Diff, name, original, current string) map[string]Diff {
	if m == nil {
		m = make(map[string]Diff)
	}
	m[name] = TextDiff{Values: [2]string{original, current}}
	return m
}
