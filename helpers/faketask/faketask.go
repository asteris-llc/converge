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

package faketask

import (
	"errors"

	"github.com/asteris-llc/converge/resource"
)

// FakeTask for testing things that require real tasks
type FakeTask struct {
	Status     string
	WillChange bool
	Error      error
}

// Check returns values set on struct
func (ft *FakeTask) Check(resource.Renderer) (resource.TaskStatus, error) {
	return &resource.Status{Output: []string{ft.Status}, Status: ft.Status, WillChange: ft.WillChange}, ft.Error
}

// Apply returns values set on struct
func (ft *FakeTask) Apply(r resource.Renderer) (resource.TaskStatus, error) {
	return ft.Check(r)
}

// NoOp returns a FakeTask that doesn't have to do anything
func NoOp() *FakeTask {
	return &FakeTask{
		Status:     "all good",
		WillChange: false,
		Error:      nil,
	}
}

// Error returns a FakeTask that will result in an error while checking or
// applying
func Error() *FakeTask {
	return &FakeTask{
		Status:     "error",
		WillChange: false,
		Error:      errors.New("error"),
	}
}

// WillChange returns a FakeTask that will always change
func WillChange() *FakeTask {
	return &FakeTask{
		Status:     "changed",
		WillChange: true,
		Error:      nil,
	}
}

// FakeSwapper is a task that tracks its state so that it can change between
// calls to Apply
type FakeSwapper struct {
	Status     string
	WillChange bool
	Error      error
}

// Check returns values set on struct
func (ft *FakeSwapper) Check(resource.Renderer) (resource.TaskStatus, error) {
	return &resource.Status{Output: []string{ft.Status}, Status: ft.Status, WillChange: ft.WillChange}, ft.Error
}

// Apply negates the current WillChange value set on struct and returns
// configured error
func (ft *FakeSwapper) Apply(r resource.Renderer) (resource.TaskStatus, error) {
	ft.WillChange = !ft.WillChange
	return ft.Check(r)
}

// Swapper creates a new stub swapper with an initial WillChange value of true
func Swapper() *FakeSwapper {
	return &FakeSwapper{
		Status:     "swapper",
		WillChange: true,
		Error:      nil,
	}
}
