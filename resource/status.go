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

// StatusLevel will be used as a level in Status. It indicates if a resource
// needs to be changed, as well as fatal conditions.
type StatusLevel uint32

const (
	// StatusNoChange means no changes are necessary. This status signals that
	// execution of dependent resources can continue.
	StatusNoChange StatusLevel = iota

	// StatusWontChange indicates an acceptable delta that wont be corrected.
	// This status signals that execution of dependent resources can continue.
	StatusWontChange

	// StatusWillChange indicates an unacceptable delta that will be corrected.
	// This status signals that execution of dependent resources can continue.
	StatusWillChange

	// StatusCantChange indicates an unacceptable delta that can't be corrected.
	// This is just like StatusFatal except the user will see that the resource
	// needs to change, but can't because of the condition specified in your
	// messaging. This status halts execution of dependent resources.
	StatusCantChange

	// StatusFatal indicates an error. This is just like StatusCantChange except
	// it does not imply that there are changes to be made. This status halts
	// execution of dependent resources.
	StatusFatal
)

func (l StatusLevel) String() string {
	switch l {
	case StatusNoChange:
		return "no change"

	case StatusWontChange:
		return "won't change"

	case StatusWillChange:
		return "will change"

	case StatusCantChange:
		return "can't change"

	case StatusFatal:
		return "fatal"
	}

	return "invalid status level"
}

type badDep struct {
	ID     string
	Status TaskStatus
}

// TaskStatus represents the results of Check called during planning or
// application.
type TaskStatus interface {
	Diffs() map[string]Diff
	StatusCode() StatusLevel
	Messages() []string
	HasChanges() bool
}

// Status is the default TaskStatus implementation
type Status struct {
	// Differences contains the things that will change as a part of this
	// Status. This will be used almost exclusively in the Check phase of
	// operations on resources. Use `NewStatus` to get a Status with this
	// initialized properly.
	Differences map[string]Diff

	// Output is the human-consumable fields on this struct. Output will be
	// returned as the Status' messages
	Output []string

	// Level indicates the change level of the status. Level is a gradation (see
	// the Status* contsts above.)
	Level StatusLevel

	failingDeps []badDep
}

// NewStatus returns a Status with all fields initialized
func NewStatus() *Status {
	return &Status{
		Differences: map[string]Diff{},
	}
}

// Diffs returns the internal differences
func (t *Status) Diffs() map[string]Diff {
	return t.Differences
}

// StatusCode returns the current warning level
func (t *Status) StatusCode() StatusLevel {
	return t.Level
}

// Messages returns the current output slice
func (t *Status) Messages() []string {
	return t.Output
}

// HasChanges returns the WillChange value
func (t *Status) HasChanges() bool {
	if t.Level == StatusWillChange || t.Level == StatusCantChange {
		return true
	}

	for _, diff := range t.Diffs() {
		if diff.Changes() {
			return true
		}
	}

	return false
}

// HealthCheck provides a default health check implementation for statuses
func (t *Status) HealthCheck() (status *HealthStatus, err error) {
	status = &HealthStatus{TaskStatus: t, FailingDeps: make(map[string]string)}
	if !t.HasChanges() && len(t.failingDeps) == 0 {
		return
	}

	// There are changes or failing dependencies so the health check is at least
	// at a warning status.
	status.UpgradeWarning(StatusWarning)

	for _, failingDep := range t.failingDeps {
		status.FailingDeps[failingDep.ID] = fmt.Sprintf("returned %d", failingDep.Status.StatusCode())
	}
	if t.StatusCode() >= 2 {
		status.UpgradeWarning(StatusError)
	}
	return
}

// FailingDep tracks a new failing dependency
func (t *Status) FailingDep(id string, stat TaskStatus) {
	t.failingDeps = append(t.failingDeps, badDep{ID: id, Status: stat})
}

// AddDifference adds a TextDiff to the Differences map
func (t *Status) AddDifference(name, original, current, defaultVal string) {
	t.Differences = AddTextDiff(t.Differences, name, original, current, defaultVal)
}

// AddMessage adds a human-readable message(s) to the output
func (t *Status) AddMessage(message ...string) {
	t.Output = append(t.Output, message...)
}

// RaiseLevel raises the status level to the given level
func (t *Status) RaiseLevel(level StatusLevel) {
	if level > t.Level {
		t.Level = level
	}
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
