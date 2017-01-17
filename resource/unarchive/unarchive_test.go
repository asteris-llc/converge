// Copyright © 2017 Asteris, LLC
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

package unarchive_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/asteris-llc/converge/helpers/fakerenderer"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/unarchive"
	"github.com/fgrid/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
)

// TestUnarchiveInterface tests that Unarchive is properly implemented
func TestUnarchiveInterface(t *testing.T) {
	t.Parallel()

	assert.Implements(t, (*resource.Task)(nil), new(unarchive.Unarchive))
}

// TestCheck tests the cases Check handles
func TestCheck(t *testing.T) {
	t.Parallel()

	src, err := ioutil.TempFile("", "unarchive_test.txt")
	require.NoError(t, err)
	defer os.Remove(src.Name())

	destInvalid, err := ioutil.TempFile("", "unarchive_test.txt")
	require.NoError(t, err)
	defer os.Remove(destInvalid.Name())

	t.Run("error", func(t *testing.T) {
		u := unarchive.Unarchive{
			Source:      src.Name(),
			Destination: destInvalid.Name(),
		}

		status, err := u.Check(context.Background(), fakerenderer.New())

		assert.Error(t, err)
		assert.True(t, status.HasChanges())
	})

	t.Run("unarchive", func(t *testing.T) {
		u := unarchive.Unarchive{
			Source:      src.Name(),
			Destination: "/tmp",
		}

		status, err := u.Check(context.Background(), fakerenderer.New())

		assert.NoError(t, err)
		assert.Equal(t, u.Source, status.Diffs()["unarchive"].Original())
		assert.Equal(t, u.Destination, status.Diffs()["unarchive"].Current())
		assert.True(t, status.HasChanges())
	})
}

// TestApply tests the cases Apply handles
func TestApply(t *testing.T) {
	t.Parallel()

	t.Run("error", func(t *testing.T) {
		u := unarchive.Unarchive{}

		status, err := u.Apply(context.Background())

		assert.EqualError(t, err, "cannot unarchive: stat : no such file or directory")
		assert.True(t, status.HasChanges())
	})
}

// TestDiff tests Diff for Unarchive
func TestDiff(t *testing.T) {
	t.Parallel()

	src, err := ioutil.TempFile("", "unarchive_test.txt")
	require.NoError(t, err)
	defer os.Remove(src.Name())

	destInvalid, err := ioutil.TempFile("", "unarchive_test.txt")
	require.NoError(t, err)
	defer os.Remove(destInvalid.Name())

	tempFileDir := strings.Split(uuid.NewV4().String(), "-")
	fakeFileDir := strings.Join(tempFileDir[0:], "")

	t.Run("source does not exist", func(t *testing.T) {
		u := unarchive.Unarchive{
			Source:      fakeFileDir,
			Destination: "/tmp",
		}
		status := resource.NewStatus()

		err := u.Diff(status)

		assert.EqualError(t, err, fmt.Sprintf("cannot unarchive: stat %s: no such file or directory", u.Source))
		assert.True(t, status.HasChanges())
	})

	t.Run("destination is not directory", func(t *testing.T) {
		u := unarchive.Unarchive{
			Source:      src.Name(),
			Destination: destInvalid.Name(),
		}
		status := resource.NewStatus()

		err := u.Diff(status)

		assert.EqualError(t, err, fmt.Sprintf("invalid destination \"%s\", must be directory", u.Destination))
		assert.True(t, status.HasChanges())
	})

	t.Run("destination does not exist", func(t *testing.T) {
		u := unarchive.Unarchive{
			Source:      src.Name(),
			Destination: fakeFileDir,
		}
		status := resource.NewStatus()

		err := u.Diff(status)

		assert.EqualError(t, err, fmt.Sprintf("destination \"%s\" does not exist", u.Destination))
		assert.True(t, status.HasChanges())
	})

	t.Run("unarchive", func(t *testing.T) {
		u := unarchive.Unarchive{
			Source:      src.Name(),
			Destination: "/tmp",
		}
		status := resource.NewStatus()

		err := u.Diff(status)

		assert.NoError(t, err)
		assert.Equal(t, u.Source, status.Diffs()["unarchive"].Original())
		assert.Equal(t, u.Destination, status.Diffs()["unarchive"].Current())
		assert.True(t, status.HasChanges())
	})

}
