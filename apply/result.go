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
	Ran    bool
	Status resource.TaskStatus
	Err    error

	Plan *plan.Result
}

// Messages returns any result status messages supplied by the task
func (r *Result) Messages() string {
	return r.Status.Value()
}

// Changes returns the fields that changed
func (r *Result) Changes() map[string]resource.Diff {
	return r.Plan.Changes()
}

// HasChanges indicates if this result ran
func (r *Result) HasChanges() bool { return r.Ran }

// Error returns the error assigned to this Result, if any
func (r *Result) Error() error {
	return r.Err
}
