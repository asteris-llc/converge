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
	"testing"

	"github.com/asteris-llc/converge/helpers/fakerenderer"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/package/rpm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
)

// TestPackageInterfaces ensures the correct interfaces are implemented
func TestPackageInterfaces(t *testing.T) {
	t.Parallel()
	assert.Implements(t, (*resource.Task)(nil), new(rpm.Package))
}

// TestPackageState ensures that package state queries work correctly
func TestPackageState(t *testing.T) {
	t.Parallel()
	p := &rpm.Package{Name: "foo"}
	t.Run("when installed", func(t *testing.T) {
		p.PkgMgr = &rpm.YumManager{Sys: newRunner("", makeExitError("", 0))}
		assert.Equal(t, rpm.StatePresent, p.PackageState())
	})
	t.Run("when not installed", func(t *testing.T) {
		p.PkgMgr = &rpm.YumManager{Sys: newRunner("", makeExitError("", 1))}
		assert.Equal(t, rpm.StateAbsent, p.PackageState())
	})
}

// TestCheck ensures Check works correctly
func TestCheck(t *testing.T) {
	t.Parallel()

	// runner := newRunner("", makeExitError("", 0))
	t.Run("when present/present", func(t *testing.T) {
		p := &rpm.Package{State: rpm.StatePresent}
		p.PkgMgr = &rpm.YumManager{newRunner("", nil)}
		status, err := p.Check(context.Background(), fakerenderer.New())
		require.NoError(t, err)
		assert.False(t, status.HasChanges())
	})
	t.Run("when absent/absent", func(t *testing.T) {
		p := &rpm.Package{State: rpm.StateAbsent}
		p.PkgMgr = &rpm.YumManager{newRunner("", makeExitError("", 1))}
		status, err := p.Check(context.Background(), fakerenderer.New())
		require.NoError(t, err)
		assert.False(t, status.HasChanges())
	})
	t.Run("when should be removed", func(t *testing.T) {
		p := &rpm.Package{State: rpm.StateAbsent}
		p.PkgMgr = &rpm.YumManager{newRunner("", nil)}
		status, err := p.Check(context.Background(), fakerenderer.New())
		require.NoError(t, err)
		assert.True(t, status.HasChanges())
	})
	t.Run("when should be installed", func(t *testing.T) {
		p := &rpm.Package{State: rpm.StatePresent}
		p.PkgMgr = &rpm.YumManager{newRunner("", makeExitError("", 1))}
		status, err := p.Check(context.Background(), fakerenderer.New())
		require.NoError(t, err)
		assert.True(t, status.HasChanges())
	})
}

// TestApply ensures Apply works correctly
func TestApply(t *testing.T) {
	t.Parallel()

	t.Run("when present/present", func(t *testing.T) {
		p := &rpm.Package{State: rpm.StatePresent}
		p.PkgMgr = &rpm.YumManager{newRunner("", nil)}
		status, err := p.Apply(context.Background())
		require.NoError(t, err)
		assert.False(t, status.HasChanges())
	})
	t.Run("when absent/absent", func(t *testing.T) {
		p := &rpm.Package{State: rpm.StateAbsent}
		p.PkgMgr = &rpm.YumManager{newRunner("", makeExitError("", 1))}
		status, err := p.Apply(context.Background())
		require.NoError(t, err)
		assert.False(t, status.HasChanges())
	})
	t.Run("when should be removed", func(t *testing.T) {
		p := &rpm.Package{State: rpm.StateAbsent}
		p.PkgMgr = &rpm.YumManager{newRunner("", nil)}
		status, err := p.Apply(context.Background())
		require.NoError(t, err)
		assert.True(t, status.HasChanges())
	})
}
