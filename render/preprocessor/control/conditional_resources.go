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

package control

import (
	"github.com/asteris-llc/converge/resource"
	"github.com/pkg/errors"
)

// ConditionalTask represents a task that may or may not be executed. It's
// evaluation is determined by it's parent control-structure predicate.
type ConditionalTask struct {
	resource.Task
	Name       string
	controller EvaluationController
}

// EvaluationController represents an interface for a thing that can control
// conditional execution (e.g. a CasePreparer or CaseTask)
type EvaluationController interface {
	ShouldEvaluate() bool
}

// SetExecutionController sets the private execution controller
func (c *ConditionalTask) SetExecutionController(ctrl EvaluationController) {
	c.controller = ctrl
}

// Apply will conditionally apply a task
func (c *ConditionalTask) Apply() (resource.TaskStatus, error) {
	if c.controller.ShouldEvaluate() {
		return c.Task.Apply()
	}
	return &resource.Status{}, nil
}

// Check will conditionally check a task
func (c *ConditionalTask) Check(r resource.Renderer) (resource.TaskStatus, error) {
	if c == nil {
		return &resource.Status{}, errors.New("conditional task is nil")
	}
	if c.controller.ShouldEvaluate() {
		return c.Task.Check(r)
	}
	return &resource.Status{}, nil
}
