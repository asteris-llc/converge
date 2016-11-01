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

package pkg_test

import (
	"fmt"
	"os/exec"
	"testing"

	"github.com/asteris-llc/converge/helpers/fakerenderer"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/package"
	"github.com/asteris-llc/converge/resource/package/rpm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// TestPackageInterfaces ensures the correct interfaces are implemented
func TestPackageInterfaces(t *testing.T) {
	t.Parallel()
	assert.Implements(t, (*resource.Task)(nil), new(pkg.Package))
}

// TestPackageState ensures that package state queries work correctly
func TestPackageState(t *testing.T) {
	t.Parallel()
	p := &pkg.Package{Name: "foo"}
	t.Run("when installed", func(t *testing.T) {
		p.PkgMgr = &rpm.YumManager{Sys: newRunner("", makeExitError("", 0))}
		assert.Equal(t, pkg.StatePresent, p.PackageState())
	})
	t.Run("when not installed", func(t *testing.T) {
		p.PkgMgr = &rpm.YumManager{Sys: newRunner("", makeExitError("", 1))}
		assert.Equal(t, pkg.StateAbsent, p.PackageState())
	})
}

// TestCheck ensures Check works correctly
func TestCheck(t *testing.T) {
	t.Parallel()

	// runner := newRunner("", makeExitError("", 0))
	t.Run("when present/present", func(t *testing.T) {
		p := &pkg.Package{State: pkg.StatePresent}
		p.PkgMgr = &rpm.YumManager{newRunner("", nil)}
		status, err := p.Check(fakerenderer.New())
		require.NoError(t, err)
		assert.False(t, status.HasChanges())
	})
	t.Run("when absent/absent", func(t *testing.T) {
		p := &pkg.Package{State: pkg.StateAbsent}
		p.PkgMgr = &rpm.YumManager{newRunner("", makeExitError("", 1))}
		status, err := p.Check(fakerenderer.New())
		require.NoError(t, err)
		assert.False(t, status.HasChanges())
	})
	t.Run("when should be removed", func(t *testing.T) {
		p := &pkg.Package{State: pkg.StateAbsent}
		p.PkgMgr = &rpm.YumManager{newRunner("", nil)}
		status, err := p.Check(fakerenderer.New())
		require.NoError(t, err)
		assert.True(t, status.HasChanges())
	})
	t.Run("when should be installed", func(t *testing.T) {
		p := &pkg.Package{State: pkg.StatePresent}
		p.PkgMgr = &rpm.YumManager{newRunner("", makeExitError("", 1))}
		status, err := p.Check(fakerenderer.New())
		require.NoError(t, err)
		assert.True(t, status.HasChanges())
	})
}

// TestApply ensures Apply works correctly
func TestApply(t *testing.T) {
	t.Parallel()

	t.Run("when present/present", func(t *testing.T) {
		p := &pkg.Package{State: pkg.StatePresent}
		p.PkgMgr = &rpm.YumManager{newRunner("", nil)}
		status, err := p.Apply()
		require.NoError(t, err)
		assert.False(t, status.HasChanges())
	})
	t.Run("when absent/absent", func(t *testing.T) {
		p := &pkg.Package{State: pkg.StateAbsent}
		p.PkgMgr = &rpm.YumManager{newRunner("", makeExitError("", 1))}
		status, err := p.Apply()
		require.NoError(t, err)
		assert.False(t, status.HasChanges())
	})
	t.Run("when should be removed", func(t *testing.T) {
		p := &pkg.Package{State: pkg.StateAbsent}
		p.PkgMgr = &rpm.YumManager{newRunner("", nil)}
		status, err := p.Apply()
		require.NoError(t, err)
		assert.True(t, status.HasChanges())
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
