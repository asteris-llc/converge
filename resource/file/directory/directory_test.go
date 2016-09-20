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

package directory_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/asteris-llc/converge/helpers/fakerenderer"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/file/directory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDirectoryCheck(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "converge-directory-check")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	t.Run("already-exists", func(t *testing.T) {
		dir := directory.Directory{Destination: tmpDir}

		plan, err := dir.Check(fakerenderer.New())
		require.NoError(t, err)

		assert.False(t, plan.HasChanges())
		assert.Equal(
			t,
			[]string{fmt.Sprintf("%q already exists", tmpDir)},
			plan.Messages(),
		)
		assert.Equal(t, resource.StatusNoChange, plan.StatusCode())
		assert.Empty(t, plan.Diffs())
	})

	t.Run("new", func(t *testing.T) {
		dest := path.Join(tmpDir, "x")
		dir := directory.Directory{Destination: dest}

		plan, err := dir.Check(fakerenderer.New())
		require.NoError(t, err)

		assert.True(t, plan.HasChanges())
		assert.Equal(t, resource.StatusWillChange, plan.StatusCode())

		diffs := plan.Diffs()
		if diff := diffs[dest]; assert.NotNil(t, diff) {
			assert.Equal(t, "<absent>", diff.Original())
			assert.Equal(t, "<present>", diff.Current())
			assert.True(t, diff.Changes())
		}
	})

	t.Run("new-2-deep", func(t *testing.T) {
		dest := path.Join(tmpDir, "x", "y")
		dir := directory.Directory{
			Destination: dest,
			CreateAll:   true,
		}

		plan, err := dir.Check(fakerenderer.New())
		require.NoError(t, err)

		assert.True(t, plan.HasChanges())
		assert.Equal(t, resource.StatusWillChange, plan.StatusCode())

		// make sure all the diffs are going from absent -> present and that
		// we're creating all of them except for tmpDir
		diffs := plan.Diffs()
		diffDest := dest
		for diffDest != tmpDir {
			if diff := diffs[diffDest]; assert.NotNil(t, diff) {
				assert.Equal(t, "<absent>", diff.Original(), diffDest)
				assert.Equal(t, "<present>", diff.Current(), diffDest)
				assert.True(t, diff.Changes())
			}
			diffDest = path.Dir(diffDest)
		}
	})

	t.Run("new-2-deep-no-create-all", func(t *testing.T) {
		dest := path.Join(tmpDir, "x", "y")
		dir := directory.Directory{
			Destination: dest,
			CreateAll:   false,
		}

		plan, err := dir.Check(fakerenderer.New())
		require.NoError(t, err)

		assert.True(t, plan.HasChanges())
		assert.Equal(
			t,
			[]string{fmt.Sprintf("%q does not exist and will not be created (enable create_all to do this)", path.Dir(dest))},
			plan.Messages(),
		)
		assert.Equal(t, resource.StatusCantChange, plan.StatusCode())
		assert.Empty(t, plan.Diffs())
	})

	t.Run("file", func(t *testing.T) {
		dest := path.Join(tmpDir, "file")
		require.NoError(t, ioutil.WriteFile(dest, []byte("test"), 777))
		defer os.Remove(dest)

		dir := directory.Directory{Destination: dest}
		plan, err := dir.Check(fakerenderer.New())
		require.NoError(t, err)

		assert.True(t, plan.HasChanges())
		assert.Equal(
			t,
			[]string{fmt.Sprintf("%q already exists and is not a directory", dest)},
			plan.Messages(),
		)
		assert.Equal(t, resource.StatusCantChange, plan.StatusCode())
		assert.Empty(t, plan.Diffs())
	})
}

func TestDirectoryApply(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "converge-directory-apply")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	t.Run("one-level", func(t *testing.T) {
		dest := path.Join(tmpDir, "one-level")
		dir := directory.Directory{Destination: dest}

		apply, err := dir.Apply()
		require.NoError(t, err)

		assert.Equal(
			t,
			[]string{fmt.Sprintf("%q exists", dest)},
			apply.Messages(),
		)
		assert.Equal(t, resource.StatusWillChange, apply.StatusCode())
	})

	t.Run("two-levels", func(t *testing.T) {
		dest := path.Join(tmpDir, "two-levels", "second")
		dir := directory.Directory{
			Destination: dest,
			CreateAll:   true,
		}

		apply, err := dir.Apply()
		require.NoError(t, err)

		assert.Equal(
			t,
			[]string{fmt.Sprintf("%q exists", dest)},
			apply.Messages(),
		)
		assert.Equal(t, resource.StatusWillChange, apply.StatusCode())
	})

	t.Run("error", func(t *testing.T) {
		dest := path.Join(tmpDir, "file")
		require.NoError(t, ioutil.WriteFile(dest, []byte("test"), 777))
		defer os.Remove(dest)

		dir := directory.Directory{Destination: dest}

		_, err := dir.Apply()
		require.Error(t, err)
	})
}
