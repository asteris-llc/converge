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

package unarchive

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/asteris-llc/converge/helpers/fakerenderer"
	"github.com/asteris-llc/converge/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
)

// TestUnarchiveInterface tests that Unarchive is properly implemented
func TestUnarchiveInterface(t *testing.T) {
	t.Parallel()

	assert.Implements(t, (*resource.Task)(nil), new(Unarchive))
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
		u := &Unarchive{
			Source:      src.Name(),
			Destination: destInvalid.Name(),
		}

		status, err := u.Check(context.Background(), fakerenderer.New())

		assert.EqualError(t, err, fmt.Sprintf("invalid destination %q, must be directory", u.Destination))
		assert.Equal(t, resource.StatusCantChange, status.StatusCode())
		assert.True(t, status.HasChanges())
	})

	t.Run("unarchive", func(t *testing.T) {
		u := &Unarchive{
			Source:      src.Name(),
			Destination: "/tmp",
		}

		status, err := u.Check(context.Background(), fakerenderer.New())

		assert.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("fetch and unarchive %q", u.Source), status.Messages()[0])
		assert.Equal(t, u.Source, status.Diffs()["unarchive"].Original())
		assert.Equal(t, u.Destination, status.Diffs()["unarchive"].Current())
		assert.Equal(t, resource.StatusWillChange, status.StatusCode())
		assert.True(t, status.HasChanges())
	})
}

// TestApply tests the cases Apply handles
func TestApply(t *testing.T) {
	t.Parallel()

	t.Run("error", func(t *testing.T) {
		u := &Unarchive{}

		status, err := u.Apply(context.Background())

		assert.EqualError(t, err, "cannot unarchive: stat : no such file or directory")
		assert.Equal(t, resource.StatusCantChange, status.StatusCode())
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

	t.Run("source does not exist", func(t *testing.T) {
		u := &Unarchive{
			Source:      "",
			Destination: "/tmp",
		}
		status := resource.NewStatus()

		err := u.Diff(status)

		assert.EqualError(t, err, fmt.Sprintf("cannot unarchive: stat %s: no such file or directory", u.Source))
		assert.Equal(t, resource.StatusCantChange, status.StatusCode())
		assert.True(t, status.HasChanges())
	})

	t.Run("destination is not directory", func(t *testing.T) {
		u := &Unarchive{
			Source:      src.Name(),
			Destination: destInvalid.Name(),
		}
		status := resource.NewStatus()

		err := u.Diff(status)

		assert.EqualError(t, err, fmt.Sprintf("invalid destination \"%s\", must be directory", u.Destination))
		assert.Equal(t, resource.StatusCantChange, status.StatusCode())
		assert.True(t, status.HasChanges())
	})

	t.Run("destination does not exist", func(t *testing.T) {
		u := &Unarchive{
			Source:      src.Name(),
			Destination: "",
		}
		status := resource.NewStatus()

		err := u.Diff(status)

		assert.EqualError(t, err, fmt.Sprintf("destination \"%s\" does not exist", u.Destination))
		assert.Equal(t, resource.StatusCantChange, status.StatusCode())
		assert.True(t, status.HasChanges())
	})

	t.Run("unarchive", func(t *testing.T) {
		u := &Unarchive{
			Source:      src.Name(),
			Destination: "/tmp",
		}
		status := resource.NewStatus()

		err := u.Diff(status)

		assert.NoError(t, err)
		assert.Equal(t, u.Source, status.Diffs()["unarchive"].Original())
		assert.Equal(t, u.Destination, status.Diffs()["unarchive"].Current())
		assert.Equal(t, resource.StatusWillChange, status.StatusCode())
		assert.True(t, status.HasChanges())
	})
}

// TestSetDirsAndContents tests setDirsAndContents for Unarchive
func TestSetDirsAndContents(t *testing.T) {
	t.Parallel()

	srcFile, err := ioutil.TempFile("", "unarchive_test.zip")
	require.NoError(t, err)
	defer os.Remove(srcFile.Name())

	t.Run("create destination", func(t *testing.T) {
		notExistDir := "/tmp/unarchive_test12345678"
		_, err := os.Stat(notExistDir)
		require.True(t, os.IsNotExist(err))

		u := &Unarchive{
			Source:      srcFile.Name(),
			Destination: notExistDir,
		}
		defer os.RemoveAll(u.Destination)

		evalDups, err := u.setDirsAndContents()

		_, exists := os.Stat(notExistDir)

		assert.NoError(t, err)
		assert.False(t, evalDups)
		assert.False(t, os.IsNotExist(exists))
	})

	t.Run("empty dest", func(t *testing.T) {
		emptyDir, err := ioutil.TempDir("", "unarchive_empty")
		require.NoError(t, err)
		defer os.RemoveAll(emptyDir)

		u := &Unarchive{
			Source:      srcFile.Name(),
			Destination: emptyDir,
		}
		defer os.RemoveAll(u.Destination)

		_, tempFetchLoc, err := setupSetDirsAndContents(u, false)
		require.NoError(t, err)
		defer os.RemoveAll(tempFetchLoc)

		expected := []string(nil)

		evalDups, err := u.setDirsAndContents()

		assert.NoError(t, err)
		assert.False(t, evalDups)
		assert.Equal(t, 0, len(u.destContents))
		assert.Equal(t, expected, u.fetchContents)
	})

	t.Run("fetch dir", func(t *testing.T) {
		destDir, err := ioutil.TempDir("", "unarchive_dir")
		require.NoError(t, err)
		defer os.RemoveAll(destDir)

		nestedDir, err := ioutil.TempDir(destDir, "unarchive_nest_dir")
		require.NoError(t, err)
		defer os.RemoveAll(nestedDir)

		nestedFile, err := ioutil.TempFile(destDir, "unarchive_nest_file")
		require.NoError(t, err)
		defer os.Remove(nestedFile.Name())

		u := &Unarchive{
			Source:      srcFile.Name(),
			Destination: destDir,
		}
		defer os.RemoveAll(u.Destination)

		_, tempFetchLoc, err := setupSetDirsAndContents(u, false)
		require.NoError(t, err)
		defer os.RemoveAll(tempFetchLoc)

		expected := [1]string{"fetchFileA.txt"}

		evalDups, err := u.setDirsAndContents()

		assert.NoError(t, err)
		assert.True(t, evalDups)
		assert.Equal(t, 1, len(u.fetchContents))
		assert.Contains(t, u.fetchContents[0], expected[0])
	})

	t.Run("nested fetch dir", func(t *testing.T) {
		destDir, err := ioutil.TempDir("", "unarchive_dir")
		require.NoError(t, err)
		defer os.RemoveAll(destDir)

		nestedDir, err := ioutil.TempDir(destDir, "unarchive_nest_dir")
		require.NoError(t, err)
		defer os.RemoveAll(nestedDir)

		nestedFile, err := ioutil.TempFile(destDir, "unarchive_nest_file")
		require.NoError(t, err)
		defer os.Remove(nestedFile.Name())

		u := &Unarchive{
			Source:      srcFile.Name(),
			Destination: destDir,
		}
		defer os.RemoveAll(u.Destination)

		_, tempFetchLoc, err := setupSetDirsAndContents(u, true)
		require.NoError(t, err)
		defer os.RemoveAll(tempFetchLoc)

		expected := [2]string{"fetchFileB.txt", "fetchFileC.txt"}

		evalDups, err := u.setDirsAndContents()

		assert.NoError(t, err)
		assert.True(t, evalDups)
		assert.Equal(t, 2, len(u.fetchContents))
		assert.True(t, strings.Contains(u.fetchContents[0], expected[0]) || strings.Contains(u.fetchContents[1], expected[0]))
		assert.True(t, strings.Contains(u.fetchContents[0], expected[1]) || strings.Contains(u.fetchContents[1], expected[1]))
	})
}

// TestEvaluateDuplicates tests evaluateDuplicates for Unarchive
func TestEvaluateDuplicates(t *testing.T) {
	t.Parallel()

}

// setupSetDirsAndContents performs some setup required to test
// SetDirsAndContents
func setupSetDirsAndContents(u *Unarchive, nested bool) (*resource.Status, string, error) {
	status := resource.NewStatus()

	err := u.setFetchLoc()
	if err != nil {
		return status, "", err
	}

	modifyFetchLocForTest(u)

	f := stubFetch{}
	status, tempFetchLoc, err := f.Apply(context.Background(), u.fetchLoc, nested)

	// set the fetchLoc again based on the tempFetchLoc
	u.fetchLoc = tempFetchLoc

	return status, tempFetchLoc, err
}

// modifyFetchLocForTest changes the fetchLoc to reside in /tmp
func modifyFetchLocForTest(u *Unarchive) {
	info := strings.Split(u.fetchLoc, "/")
	u.fetchLoc = "/tmp/" + info[len(info)-1]
}

// stubFetch stubs the Fetch resource
type stubFetch struct{}

func (f stubFetch) Apply(ctx context.Context, fetchLoc string, nested bool) (*resource.Status, string, error) {
	status := resource.NewStatus()

	info := strings.Split(fetchLoc, "/")

	tempFetchLoc, err := ioutil.TempDir("", info[len(info)-1])
	if err != nil {
		return status, tempFetchLoc, err
	}

	if !nested {
		_, err := ioutil.TempFile(tempFetchLoc, "fetchFileA.txt")
		if err != nil {
			return status, tempFetchLoc, err
		}
	} else {
		nestedFetchDir, err := ioutil.TempDir(tempFetchLoc, "unarchive_fetch_nest")
		if err != nil {
			return status, tempFetchLoc, err
		}

		_, err = ioutil.TempFile(nestedFetchDir, "fetchFileB.txt")
		if err != nil {
			return status, tempFetchLoc, err
		}

		_, err = ioutil.TempFile(nestedFetchDir, "fetchFileC.txt")
		if err != nil {
			return status, tempFetchLoc, err
		}
	}

	status.RaiseLevel(resource.StatusWillChange)
	status.AddDifference("destination", "<absent>", fetchLoc, "")
	status.AddMessage("fetched successfully")

	return status, tempFetchLoc, nil
}
