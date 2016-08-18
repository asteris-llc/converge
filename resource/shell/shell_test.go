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

package shell_test

import (
	"errors"
	"testing"

	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/shell"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var any = mock.Anything

func Test_Shell_ImplementsTaskInterface(t *testing.T) {
	t.Parallel()
	assert.Implements(t, (*resource.Task)(nil), new(shell.Shell))
}

func Test_Check_WhenRunReturnsError_ReturnsError(t *testing.T) {
	expected := errors.New("test error")
	m := new(MockExecutor)
	m.On("Run", any).Return(&shell.CommandResults{}, expected)
	sh := testShell(m)
	_, actual := sh.Check()
	assert.Error(t, actual)
}

func Test_Check_WhenRunReturnsResults_PrependsResutsToStatus(t *testing.T) {
	expected := errors.New("test error")
	m := new(MockExecutor)
	m.On("Run", any).Return(&shell.CommandResults{}, expected)
	sh := testShell(m)
	_, actual := sh.Check()
	assert.Error(t, actual)
}

func testShell(c shell.CommandExecutor) *shell.Shell {
	return &shell.Shell{CmdGenerator: c}
}

func defaultTestShell() *shell.Shell {
	return testShell(defaultExecutor())
}

type MockExecutor struct {
	mock.Mock
}

func (m *MockExecutor) Run(script string) (*shell.CommandResults, error) {
	args := m.Called(script)
	return args.Get(0).(*shell.CommandResults), args.Error(1)
}

func defaultExecutor() *MockExecutor {
	m := new(MockExecutor)
	m.On("Run", any).Return(&shell.CommandResults{})
	return m
}
