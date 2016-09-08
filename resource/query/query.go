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

package query

import (
	"errors"

	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/shell"
)

// Query represents an environmental query
type Query struct {
	CmdGenerator shell.CommandExecutor
	Status       *shell.CommandResults
	Query        string
	hasRun       bool
}

// Check runs the query if it hasn't yet been run
func (q *Query) Check(r resource.Renderer) (resource.TaskStatus, error) {
	if q.hasRun {
		return q, nil
	}
	results, err := q.CmdGenerator.Run(q.Query)
	if err != nil {
		return nil, err
	}
	q.Status = results
	return q, nil
}

// Apply is a nop for queries.  Because HasChanges always returns false this
// should never be executed.
func (q *Query) Apply(resource.Renderer) (resource.TaskStatus, error) {
	return nil, errors.New("query apply called but it should never have changes")
}

// TaskStatus implementation

// Value provides a value for the shell, which is the stdout data from the last
// executed command.
func (q *Query) Value() string {
	return q.Status.Stdout
}

// Diffs is required to implement resource.TaskStatus but there is no mechanism
// for defining diffs for shell operations, so returns a nil map.
func (q *Query) Diffs() map[string]resource.Diff {
	return nil
}

// StatusCode returns the status code of the most recently executed command
func (q *Query) StatusCode() int {
	if q.Status == nil {
		return resource.StatusFatal
	}
	return int(q.Status.ExitStatus)
}

// Messages returns a summary of the first execution of check and/or apply.
// Subsequent runs are surpressed.
func (q *Query) Messages() (messages []string) {
	return
}

// HasChanges returns true if changes are required as determined by the the most
// recent run of check.
func (q *Query) HasChanges() bool {
	return false
}
