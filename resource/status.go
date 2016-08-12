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
	StatusCode() int
	Messages() []string
	Changes() bool
}

type Status struct {
	Differences  map[string]Diff
	WarningLevel int
	Output       []string
	WillChange   bool
}

func (t *Status) Diffs() map[string]Diff {
	return t.Differences
}

func (t *Status) StatusCode() int {
	return t.WarningLevel
}

func (t *Status) Messages() []string {
	return t.Output
}

func (t *Status) Changes() bool {
	return t.WillChange
}

func NewStatus(status string, willChange bool, err error) (TaskStatus, error) {
	return &Status{
		Output:     []string{status},
		WillChange: willChange,
	}, err
}

type Diff interface {
	Original() string
	Current() string
	Changes() bool
}

type TextDiff [2]string

func (t TextDiff) Original() string {
	if t[0] == "" {
		return "<unknown>"
	}
	return t[0]
}

func (t TextDiff) Current() string {
	if t[1] == "" {
		return "<unknown>"
	}
	return t[1]
}

func (t TextDiff) Changes() bool {
	return t[0] == t[1]
}
