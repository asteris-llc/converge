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

// ThunkTask represents an abstract task over a thunk, used when we need to
// serialized a thunked value.
type ThunkTask struct {
	Name        string
	ThunkedType Task
}

// Check returns a task status with thunk information
func (t *ThunkTask) Check(Renderer) (TaskStatus, error) {
	return t.ToStatus(), nil
}

// Apply returns a task status with thunk information
func (t *ThunkTask) Apply() (TaskStatus, error) {
	return t.ToStatus(), nil
}

// ToStatus converts a ThunkStatus to a *Status
func (t *ThunkTask) ToStatus() *Status {
	return &Status{
		Level:  StatusWillChange,
		Output: []string{fmt.Sprintf("%s depends on external node execution", t.Name)},
	}
}

// NewThunkedTask generates a ThunkTask from a PrepareThunk
func NewThunkedTask(name string, thunkedType Task) *ThunkTask {
	return &ThunkTask{
		ThunkedType: thunkedType,
		Name:        name,
	}
}
