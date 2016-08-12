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

type Diff struct {
	Original string
	Current  string
}

const (
	// ErrOk status represents no error
	ErrOk int = 0

	// ErrInfo status represents a delta that will not change
	ErrInfo int = iota

	// ErrWarning status represents a delta that will cause a change
	ErrWarning

	// ErrError status represents a severe delta that can be corrected
	ErrError

	// ErrFatal status represents an irrecoverable delta
	ErrFatal
)

type TaskStatus interface {
	Diffs() map[string]Diff
	Err() error
	Status() int
	Messages() []string
	WillChange() bool
}

type Status struct {
	Changes      map[string]Diff
	WarningLevel int
	Output       []string
	HasChanges   bool
	ErrorStatus  error
}

func (t *Status) Diffs() map[string]Diff {
	return t.Changes
}

func (t *Status) Error() error {
	return t.ErrorStatus
}

func (t *Status) Status() int {
	return t.WarningLevel
}

func (t *Status) Messages() []string {
	return t.Output
}

func (t *Status) WillChange() bool {
	return t.HasChanges
}

func NewStatus(status string, willChange bool, err error) *Status {
	return &Status{
		Output:      []string{status},
		HasChanges:  willChange,
		ErrorStatus: err,
	}
}
