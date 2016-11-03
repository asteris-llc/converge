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

	"github.com/asteris-llc/converge/healthcheck"
	"github.com/asteris-llc/converge/helpers/fakerenderer"
	"github.com/asteris-llc/converge/helpers/logging"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/shell"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/net/context"
)

var any = mock.Anything

func Test_Shell_ImplementsTaskInterface(t *testing.T) {
	t.Parallel()
	assert.Implements(t, (*resource.Task)(nil), new(shell.Shell))
	assert.Implements(t, (*healthcheck.Check)(nil), new(shell.Shell))
}

// Check

func Test_Check_WhenRunReturnsError_ReturnsError(t *testing.T) {
	expected := errors.New("test error")
	m := new(MockExecutor)
	m.On("Run", any).Return(&shell.CommandResults{}, expected)
	sh := testShell(m)
	_, actual := sh.Check(context.Background(), fakerenderer.New())
	assert.Error(t, actual)
}

func Test_Check_WhenRunReturnsResults_PrependsResutsToStatus(t *testing.T) {
	firstResult := &shell.CommandResults{}
	expectedResult := &shell.CommandResults{}
	m := &MockExecutor{}
	m.On("Run", any).Return(expectedResult, nil)
	sh := testShell(m)
	sh.Status = firstResult
	_, actual := sh.Check(context.Background(), fakerenderer.New())
	assert.NoError(t, actual)
	assert.Equal(t, expectedResult, sh.Status)
}

func Test_Check_SetsStatusOperationToCheck(t *testing.T) {
	result := &shell.CommandResults{}
	m := resultExecutor(result)
	sh := testShell(m)
	sh.Check(context.Background(), fakerenderer.New())
	assert.Equal(t, "check", result.ResultsContext.Operation)
}

func Test_Check_CallsRunWithCheckStatement(t *testing.T) {
	statement := "test statement"
	m := defaultExecutor()
	sh := &shell.Shell{CheckStmt: statement, CmdGenerator: m}
	sh.Check(context.Background(), fakerenderer.New())
	m.AssertCalled(t, "Run", statement)
}

// Apply

func Test_Apply_WhenRunReturnsError_ReturnsError(t *testing.T) {
	expected := errors.New("test error")
	m := new(MockExecutor)
	m.On("Run", any).Return(&shell.CommandResults{}, expected)
	sh := testShell(m)
	_, actual := sh.Apply(context.Background())
	assert.Error(t, actual)
}

func Test_Apply_WhenRunReturnsResults_PrependsResutsToStatus(t *testing.T) {
	firstResult := &shell.CommandResults{}
	expectedResult := &shell.CommandResults{}
	m := &MockExecutor{}
	m.On("Run", any).Return(expectedResult, nil)
	sh := testShell(m)
	sh.Status = firstResult
	_, actual := sh.Apply(context.Background())
	assert.NoError(t, actual)
	assert.Equal(t, expectedResult, sh.Status)
}

func Test_Apply_SetsStatusOperationToApply(t *testing.T) {
	result := &shell.CommandResults{}
	m := resultExecutor(result)
	sh := testShell(m)
	sh.Apply(context.Background())
	assert.Equal(t, "apply", result.ResultsContext.Operation)
}

func Test_Apply_CallsRunWithApplyStatement(t *testing.T) {
	statement := "test statement"
	m := defaultExecutor()
	sh := &shell.Shell{ApplyStmt: statement, CmdGenerator: m}
	sh.Apply(context.Background())
	m.AssertCalled(t, "Run", statement)
}

// Value

func Test_Value_ReturnsStdoutOfMostRecentStatus(t *testing.T) {
	expected := "good"
	status := &shell.CommandResults{Stdout: "bad"}
	status = status.Cons("", &shell.CommandResults{Stdout: expected})
	sh := &shell.Shell{Status: status}
	assert.Equal(t, expected, sh.Value())
}

// Diffs

func Test_Diffs_ReturnsEmptyMap(t *testing.T) {
	sh := defaultTestShell()
	assert.Equal(t, 0, len(sh.Diffs()))
}

// StatusCode

func Test_StatusCode_WhenNoStatus_ReturnsFatal(t *testing.T) {
	defer logging.HideLogs(t)()

	sh := defaultTestShell()
	assert.Equal(t, resource.StatusFatal, sh.StatusCode())
}

// TestStatusCodeWhenMultipleStatusReturnsMostRecentStatus tests what happens
// when there are multiple status returns
func TestStatusCodeWhenMultipleStatusReturnsMostRecentStatus(t *testing.T) {
	// first result returns 0 (StatusNoChange)
	status := &shell.CommandResults{ExitStatus: 0}
	// second result returns 7 (StatusWillChange)
	status = status.Cons("", &shell.CommandResults{ExitStatus: 7})
	sh := &shell.Shell{Status: status}
	assert.Equal(t, resource.StatusWillChange, sh.StatusCode())
}

// TestStatusCodeWhenExitStatusZero verifies that StatusCode returns
// StatusNoChanges when a shell command has a zero exit code
func TestStatusCodeWhenExitStatusZero(t *testing.T) {
	status := &shell.CommandResults{ExitStatus: 0}
	sh := &shell.Shell{Status: status}
	assert.Equal(t, resource.StatusNoChange, sh.StatusCode())
}

// TestStatusCodeWhenExitStatusNonZero verifies that StatusCode returns
// StatusWillChange when a shell command has a non-zero exit code
func TestStatusCodeWhenExitStatusNonZero(t *testing.T) {
	status := &shell.CommandResults{ExitStatus: 4}
	sh := &shell.Shell{Status: status}
	assert.Equal(t, resource.StatusWillChange, sh.StatusCode())
}

// Shell context

func Test_Messages_Includes_Dir(t *testing.T) {
	sh := defaultTestShell()
	sh.Dir = "/tmp/testing"
	sh.Check(context.Background(), fakerenderer.New())
	assert.Contains(t, sh.Messages(), "dir (/tmp/testing)")
}

func Test_Messages_Includes_Env(t *testing.T) {
	sh := defaultTestShell()
	sh.Env = []string{"VAR=test", "ANOTHER_VAR=test2"}
	sh.Check(context.Background(), fakerenderer.New())
	assert.Contains(t, sh.Messages(), "env (VAR=test ANOTHER_VAR=test2)")
}

// Test Utils

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
	m.On("Run", any).Return(&shell.CommandResults{}, nil)
	return m
}

func resultExecutor(r *shell.CommandResults) *MockExecutor {
	m := new(MockExecutor)
	m.On("Run", any).Return(r, nil)
	return m
}
