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

	srcDir, err := ioutil.TempDir("", "unarchive_srcDir")
	require.NoError(t, err)
	defer os.RemoveAll(srcDir)

	srcFile, err := ioutil.TempFile("", "unarchive_file.txt")
	require.NoError(t, err)
	defer os.Remove(srcFile.Name())

	destDir, err := ioutil.TempDir("", "unarchive_destDir")
	require.NoError(t, err)
	defer os.RemoveAll(destDir)

	t.Run("error", func(t *testing.T) {
		t.Run("diff error", func(t *testing.T) {
			u := &Unarchive{}

			status, err := u.Apply(context.Background())

			assert.EqualError(t, err, "cannot unarchive: stat : no such file or directory")
			assert.Equal(t, resource.StatusCantChange, status.StatusCode())
			assert.True(t, status.HasChanges())
		})

		t.Run("setFetchLoc error", func(t *testing.T) {
			u := &Unarchive{
				Source:      srcDir,
				Destination: destDir,
			}

			status, err := u.Apply(context.Background())

			assert.EqualError(t, err, fmt.Sprintf("error setting fetch location: failed to get checksum of source: failed to hash: read %s: is a directory", srcDir))
			assert.Equal(t, resource.StatusCantChange, status.StatusCode())
			assert.True(t, status.HasChanges())
		})

		t.Run("fetch error", func(t *testing.T) {
			u := &Unarchive{
				Source:      srcFile.Name(),
				Destination: destDir,
				HashType:    string(HashMD5),
				Hash:        "notarealhashbutstringnonetheless",
				Force:       false,
			}

			status, err := u.Apply(context.Background())

			assert.EqualError(t, err, "failed to fetch: invalid checksum: encoding/hex: invalid byte: U+006E 'n'")
			assert.Equal(t, resource.StatusFatal, status.StatusCode())
			assert.False(t, status.HasChanges())
		})
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

	t.Run("dest not exist", func(t *testing.T) {
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

		assert.EqualError(t, err, fmt.Sprintf("open %s: no such file or directory", u.Destination))
		assert.False(t, evalDups)
		assert.True(t, os.IsNotExist(exists))
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

	destDir, err := ioutil.TempDir("", "destDir_unarchive")
	require.NoError(t, err)
	defer os.RemoveAll(destDir)
	ddInfo, err := os.Open(destDir)
	require.NoError(t, err)

	fetchDir, err := ioutil.TempDir("", "fetchDir_unarchive")
	require.NoError(t, err)
	defer os.RemoveAll(fetchDir)
	fdInfo, err := os.Open(fetchDir)
	require.NoError(t, err)

	t.Run("no duplicates", func(t *testing.T) {
		u := &Unarchive{
			destContents:  []string{"fileA.txt", "fileB.txt"},
			destDir:       ddInfo,
			fetchContents: []string{"fileC.txt", "fileD.txt"},
			fetchDir:      fdInfo,
		}

		err = u.evaluateDuplicates()

		assert.NoError(t, err)
	})

	t.Run("duplicates", func(t *testing.T) {
		fileADest, err := os.Create(destDir + "/fileA.txt")
		require.NoError(t, err)
		defer os.Remove(fileADest.Name())

		fileAFetch, err := os.Create(fetchDir + "/fileA.txt")
		require.NoError(t, err)
		defer os.Remove(fileAFetch.Name())

		t.Run("checksum match", func(t *testing.T) {
			u := &Unarchive{
				destContents:  []string{"fileA.txt"},
				destDir:       ddInfo,
				fetchContents: []string{"fileA.txt"},
				fetchDir:      fdInfo,
			}

			err = u.evaluateDuplicates()

			assert.NoError(t, err)
		})

		t.Run("checksum mismatch", func(t *testing.T) {
			fileBDest, err := os.Create(destDir + "/fileB.txt")
			require.NoError(t, err)
			defer os.Remove(fileBDest.Name())

			fileBFetch, err := os.Create(fetchDir + "/fileB.txt")
			require.NoError(t, err)
			defer os.Remove(fileBFetch.Name())

			_, err = fileBFetch.Write([]byte{1})
			require.NoError(t, err)

			u := &Unarchive{
				Destination:   destDir,
				destContents:  []string{"fileA.txt", "fileB.txt", "fileC.txt"},
				destDir:       ddInfo,
				fetchContents: []string{"fileA.txt", "fileB.txt"},
				fetchDir:      fdInfo,
			}

			err = u.evaluateDuplicates()

			assert.EqualError(t, err, fmt.Sprintf("will not replace, file \"fileB.txt\" exists at %q: checksum mismatch", u.Destination))
		})
	})
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
