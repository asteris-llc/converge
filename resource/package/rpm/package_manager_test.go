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

package rpm_test

import (
	"fmt"
	"os/exec"
	"testing"

	"github.com/asteris-llc/converge/resource/package/rpm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestYumInstalledVersion validates that installation status is successfully
// validated
func TestYumInstalledVersion(t *testing.T) {
	t.Parallel()

	t.Run("when installed", func(t *testing.T) {
		expected := "foo-0.1.2.3"
		runner := newRunner(expected, nil)
		y := &rpm.YumManager{Sys: runner}
		result, found := y.InstalledVersion("foo1")
		assert.True(t, found)
		assert.Equal(t, expected, string(result))
	})

	t.Run("when not installed", func(t *testing.T) {
		expected := ""
		y := &rpm.YumManager{Sys: newRunner("", makeExitError("", 1))}
		result, found := y.InstalledVersion("foo1")
		assert.False(t, found)
		assert.Equal(t, expected, string(result))
	})
}

// TestYumInstallPackage validates that we successfully ask yum to install a
// package
func TestYumInstallPackage(t *testing.T) {
	t.Parallel()

	t.Run("when installed", func(t *testing.T) {
		pkg := "foo1"
		runner := newRunner("", nil)
		y := &rpm.YumManager{Sys: runner}
		_, err := y.InstallPackage(pkg)
		assert.NoError(t, err)
		runner.AssertNumberOfCalls(t, "Run", 1)
	})

	t.Run("when not installed", func(t *testing.T) {
		pkg := "foo1"
		runner := newRunner("", makeExitError("", 1))
		y := &rpm.YumManager{Sys: runner}
		y.InstallPackage(pkg)
		runner.AssertNumberOfCalls(t, "Run", 2)
	})

	t.Run("when installation error", func(t *testing.T) {
		pkg := "foo1"
		runner := newRunner("", makeExitError("", 1))
		y := &rpm.YumManager{Sys: runner}
		_, err := y.InstallPackage(pkg)
		assert.Error(t, err)
		runner.AssertNumberOfCalls(t, "Run", 2)
	})
}

// MockRunner mocks out SysCaller
type MockRunner struct {
	mock.Mock
}

// Run mocks out Run
func (m *MockRunner) Run(cmd string) ([]byte, error) {
	args := m.Called(1)
	return args.Get(0).([]byte), args.Error(1)
}

// newRunner creates a new MockRunner that returns the output string and error
func newRunner(output string, err error) *MockRunner {
	m := &MockRunner{}
	m.On("Run", mock.Anything).Return([]byte(output), err)
	return m
}

// makeExitError generates a new ExitError
func makeExitError(stderr string, exitCode uint32) error {
	cmd := fmt.Sprintf("echo %q 1>&2; exit %d", stderr, exitCode)
	_, err := exec.Command("/bin/bash", "-c", cmd).Output()
	return err
}

// queryString generates an RPM query string
func queryString(pkg string) string {
	return "rpm -q " + pkg
}

// installString generates a yum install string
func installString(pkg string) string {
	return "yum install -y " + pkg
}

// removeString generates a yum remove string
func removeString(pkg string) string {
	return "yum remove -y " + pkg
}
