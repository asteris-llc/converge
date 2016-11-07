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

package apt_test

import (
	"context"
	"testing"

	"github.com/asteris-llc/converge/helpers/fakerenderer"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/package"
	"github.com/asteris-llc/converge/resource/package/apt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPreparerInterfaces ensures that the correct interfaces are implemented by
// the preparer
func TestPreparerInterfaces(t *testing.T) {
	t.Parallel()
	assert.Implements(t, (*resource.Resource)(nil), new(apt.Preparer))
}

// estPreparerCreatesPackage tests to make sure the preparer creates valid configurations
func TestPreparerCreatesPackage(t *testing.T) {
	t.Parallel()

	t.Run("when-state-present", func(t *testing.T) {
		p := &apt.Preparer{Name: "test1", State: "present"}
		task, err := p.Prepare(context.Background(), fakerenderer.New())
		require.NoError(t, err)
		asAPT, ok := task.(*pkg.Package)
		require.True(t, ok)
		assert.Equal(t, "present", string(asAPT.State))
	})

	t.Run("when-state-absent", func(t *testing.T) {
		p := &apt.Preparer{Name: "test1", State: "absent"}
		task, err := p.Prepare(context.Background(), fakerenderer.New())
		require.NoError(t, err)
		asAPT, ok := task.(*pkg.Package)
		require.True(t, ok)
		assert.Equal(t, "absent", string(asAPT.State))
	})

	t.Run("when-state-missing", func(t *testing.T) {
		p := &apt.Preparer{Name: "test1"}
		task, err := p.Prepare(context.Background(), fakerenderer.New())
		require.NoError(t, err)
		asAPT, ok := task.(*pkg.Package)
		require.True(t, ok)
		assert.Equal(t, "present", string(asAPT.State))
	})

	t.Run("when-name-null", func(t *testing.T) {
		p := &apt.Preparer{Name: "", State: "present"}
		_, err := p.Prepare(context.Background(), fakerenderer.New())
		require.Error(t, err)
		assert.EqualError(t, err, "package name cannot be empty")
	})

	t.Run("when-name-space", func(t *testing.T) {
		p := &apt.Preparer{Name: " ", State: "present"}
		_, err := p.Prepare(context.Background(), fakerenderer.New())
		require.Error(t, err)
		assert.EqualError(t, err, "package name cannot be empty")
	})

}
