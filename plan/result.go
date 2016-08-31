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

package plan

import "github.com/asteris-llc/converge/resource"

// Result is the result of planning execution
type Result struct {
	Task   resource.Task
	Status resource.TaskStatus
	Err    error
}

// Messages returns any message values supplied by the task
func (r *Result) Messages() []string { return r.Status.Messages() }

// Changes returns the fields that will change based on this result
func (r *Result) Changes() map[string]resource.Diff { return r.Status.Diffs() }

// HasChanges indicates if this result will change
func (r *Result) HasChanges() bool { return r.Status.HasChanges() }

// Error returns the error assigned to this Result, if any
func (r *Result) Error() error { return r.Err }

// GetStatus returns the current task status
func (r *Result) GetStatus() resource.TaskStatus { return r.Status }

// GetTask returns the embedded task
func (r *Result) GetTask() (resource.Task, bool) { return r.Task, true }
