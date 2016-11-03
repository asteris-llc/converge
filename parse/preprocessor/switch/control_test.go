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

package control_test

import (
	"github.com/asteris-llc/converge/resource"
	"github.com/stretchr/testify/mock"
	"golang.org/x/net/context"
)

// any makes for less typing compared to mock.Anything
var any = mock.Anything

// This top-level test file provides utility functions and mocks other
// control_test tests.

// MockRenderer allows mocking of a resource.Render
type MockRenderer struct {
	mock.Mock
}

// GetID mocks GetID
func (m *MockRenderer) GetID() string {
	args := m.Called()
	return args.String(0)
}

// Value is a mock value
func (m *MockRenderer) Value() (resource.Value, bool) {
	args := m.Called()
	return args.String(0), args.Bool(1)
}

// Render is a mock render
func (m *MockRenderer) Render(name, content string) (string, error) {
	args := m.Called(name, content)
	return args.String(0), args.Error(1)
}

// defaultMockRenderer takes (any,any) and returns ("",nil)
func defaultMockRenderer() *MockRenderer {
	m := &MockRenderer{}
	m.On("Render", mock.Anything, mock.Anything).Return("", nil)
	m.On("Value").Return("", true)
	return m
}

// MockExecutionController mocks control.ExecutionController
type MockExecutionController struct {
	mock.Mock
}

// ShouldEvaluate mocks ShouldEvaluate
func (m *MockExecutionController) ShouldEvaluate() bool {
	args := m.Called()
	return args.Bool(0)
}

// newMockExecutionController creates a new MockExecutionController that returns
// it's param for ShouldEvaluate
func newMockExecutionController(b bool) *MockExecutionController {
	m := &MockExecutionController{}
	m.On("ShouldEvaluate").Return(b)
	return m
}

// defaultMockExecutionController creates a default that returns false
func defaultMockExecutionController() *MockExecutionController {
	return newMockExecutionController(false)
}

// MockTask mocks resource.Task
type MockTask struct {
	mock.Mock
}

// Apply mocks apply
func (m *MockTask) Apply(context.Context) (resource.TaskStatus, error) {
	args := m.Called()
	return args.Get(0).(resource.TaskStatus), args.Error(1)
}

// Check mocks check
func (m *MockTask) Check(context.Context, resource.Renderer) (resource.TaskStatus, error) {
	args := m.Called(1)
	return args.Get(0).(resource.TaskStatus), args.Error(1)
}

// newMockTask returns a new mock task that returns it's parameters
func newMockTask(tsk resource.TaskStatus, err error) *MockTask {
	t := &MockTask{}
	t.On("Check", any).Return(tsk, err)
	t.On("Apply").Return(tsk, err)
	return t
}
