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
	"golang.org/x/net/context"
)

// FakeTask for testing things that require real tasks
type FakeTask struct {
	Status string
	Level  resource.StatusLevel
	Error  error
}

// Check returns values set on struct
func (ft *FakeTask) Check(context.Context, resource.Renderer) (resource.TaskStatus, error) {
	return &resource.Status{Output: []string{ft.Status}, Level: ft.Level}, ft.Error
}

// Apply returns values set on struct
func (ft *FakeTask) Apply(context.Context) (resource.TaskStatus, error) {
	return &resource.Status{Output: []string{ft.Status}, Level: ft.Level}, ft.Error
}

// NoOp returns a FakeTask that doesn't have to do anything
func NoOp() *FakeTask {
	return &FakeTask{
		Status: "all good",
		Level:  resource.StatusWontChange,
		Error:  nil,
	}
}

// Error returns a FakeTask that will result in an error while checking or
// applying
func Error() *FakeTask {
	return &FakeTask{
		Status: "error",
		Level:  resource.StatusFatal,
		Error:  errors.New("error"),
	}
}

// WillChange returns a FakeTask that will always change
func WillChange() *FakeTask {
	return &FakeTask{
		Status: "changed",
		Level:  resource.StatusWillChange,
		Error:  nil,
	}
}

// NilTask always return (nil, error) tuple on Check/Apply calls
type NilTask struct {
}

// Check always raise error
func (*NilTask) Check(context.Context, resource.Renderer) (resource.TaskStatus, error) {
	return nil, errors.New("check error")
}

// Apply always raise error
func (*NilTask) Apply(context.Context) (resource.TaskStatus, error) {
	return nil, errors.New("apply error")
}

// NilAndError return a FakeTask that will simulate `return nil, err` case
func NilAndError() resource.Task {
	return &NilTask{}
}

// FakeSwapper is a task that tracks its state so that it can change between
// calls to Apply
type FakeSwapper struct {
	Status     string
	WillChange bool
	Error      error
}

// Check returns values set on struct
func (ft *FakeSwapper) Check(context.Context, resource.Renderer) (resource.TaskStatus, error) {
	return &resource.Status{Output: []string{ft.Status}, Level: ft.level()}, ft.Error
}

// Apply negates the current WillChange value set on struct and returns
// configured error
func (ft *FakeSwapper) Apply(context.Context) (resource.TaskStatus, error) {
	ft.WillChange = !ft.WillChange
	return &resource.Status{Output: []string{ft.Status}, Level: ft.level()}, ft.Error
}

func (ft *FakeSwapper) level() resource.StatusLevel {
	if ft.WillChange {
		return resource.StatusWillChange
	}

	return resource.StatusNoChange
}

// Swapper creates a new stub swapper with an initial WillChange value of true
func Swapper() *FakeSwapper {
	return &FakeSwapper{
		Status:     "swapper",
		WillChange: true,
		Error:      nil,
	}
}
