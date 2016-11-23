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

package apply

import (
	"github.com/asteris-llc/converge/plan"
	"github.com/asteris-llc/converge/resource"
)

// Result of application
type Result struct {
	Task   resource.Task
	Status resource.TaskStatus
	Err    error

	Ran       bool
	Plan      *plan.Result
	PostCheck resource.TaskStatus
}

// Messages returns any result status messages supplied by the task
func (r *Result) Messages() []string {
	if r.Status != nil {
		return r.Status.Messages()
	}
	return nil
}

// Changes returns the fields that changed
func (r *Result) Changes() map[string]resource.Diff {
	if r.Status != nil {
		return r.Status.Diffs()
	} else if r.PostCheck != nil {
		return r.PostCheck.Diffs()
	} else if r.Plan != nil {
		return r.Plan.Changes()
	}
	return nil
}

// HasChanges indicates if this result ran
func (r *Result) HasChanges() bool { return r.Ran }

// Error returns the error assigned to this Result, if any
func (r *Result) Error() error { return r.Err }

// Warning returns the warning assigned to this Result, if any
func (r *Result) Warning() string {
	if r.Status != nil {
		return r.Status.Warning()
	}
	return ""
}

// GetStatus returns the current task status
func (r *Result) GetStatus() resource.TaskStatus { return r.Status }

// GetTask returns the task of the embedded plan, if there is one
func (r *Result) GetTask() (resource.Task, bool) {
	if r.Plan != nil {
		return r.Plan.GetTask()
	}
	return nil, false
}
