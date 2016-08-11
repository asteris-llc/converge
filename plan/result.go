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
	Task       resource.Task
	Status     string
	WillChange bool
}

// Fields returns the fields that will change based on this result
func (r *Result) Fields() map[string][2]string {
	return map[string][2]string{
		"state": [2]string{r.Status, "<unknown>"},
	}
}

// HasChanges indicates if this result will change
func (r *Result) HasChanges() bool { return r.WillChange }

// Error always returns nil since plan results cannot fail
func (r *Result) Error() error { return nil }
