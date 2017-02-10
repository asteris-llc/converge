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
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/asteris-llc/converge/helpers/fakerenderer"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/file/fetch"
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

	srcFile, err := ioutil.TempFile("", "unarchive_test.txt")
	require.NoError(t, err)
	defer os.Remove(srcFile.Name())

	destInvalid, err := ioutil.TempFile("", "unarchive_test.txt")
	require.NoError(t, err)
	defer os.Remove(destInvalid.Name())

	srcDir, err := ioutil.TempDir("", "unarchive_srcDir")
	require.NoError(t, err)
	defer os.RemoveAll(srcDir)

	destDir, err := ioutil.TempDir("", "unarchive_destDir")
	require.NoError(t, err)
	defer os.RemoveAll(destDir)

	tmpFetchDir, err := ioutil.TempDir("", "tmpFetchDir")
	require.NoError(t, err)
	defer os.RemoveAll(tmpFetchDir)

	t.Run("error", func(t *testing.T) {
		t.Run("diff error", func(t *testing.T) {
			u := &Unarchive{
				Source:      srcFile.Name(),
				Destination: destInvalid.Name(),
			}

			status, err := u.Check(context.Background(), fakerenderer.New())

			assert.EqualError(t, err, fmt.Sprintf("invalid destination %q, must be directory", u.Destination))
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
				fetchLoc:    destInvalid.Name(),
			}

			u.fetch = fetch.Fetch{
				Source:      u.Source,
				Destination: u.fetchLoc,
				HashType:    u.HashType,
				Hash:        u.Hash,
				Unarchive:   true,
			}

			status, err := u.Check(context.Background(), fakerenderer.New())

			assert.EqualError(t, err, fmt.Sprintf("cannot attempt unarchive: fetch error: invalid destination %q for unarchiving, must be directory", u.fetch.Destination))
			assert.Equal(t, resource.StatusCantChange, status.StatusCode())
			assert.True(t, status.HasChanges())
		})
	})

	t.Run("unarchive", func(t *testing.T) {
		u := &Unarchive{
			Source:      srcFile.Name(),
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

	tmpFetchDir, err := ioutil.TempDir("", "tmpFetchDir")
	require.NoError(t, err)
	defer os.RemoveAll(tmpFetchDir)

	t.Run("error", func(t *testing.T) {
		t.Run("diff error", func(t *testing.T) {
			u := &Unarchive{}

			status, err := u.Apply(context.Background())

			assert.EqualError(t, err, fmt.Sprintf("cannot unarchive: stat %s: no such file or directory", u.Source))
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
				fetchLoc:    tmpFetchDir,
			}

			u.fetch = fetch.Fetch{
				Source:      u.Source,
				Destination: u.fetchLoc,
				HashType:    u.HashType,
				Hash:        u.Hash,
				Unarchive:   true,
			}

			status, err := u.Apply(context.Background())

			assert.EqualError(t, err, "failed to fetch: invalid checksum: encoding/hex: invalid byte: U+006E 'n'")
			assert.Equal(t, resource.StatusFatal, status.StatusCode())
			assert.False(t, status.HasChanges())
		})

		t.Run("evaluateDuplicates error", func(t *testing.T) {
			destDir, err := ioutil.TempDir("", "destDir_unarchive")
			require.NoError(t, err)
			defer os.RemoveAll(destDir)

			fileBDest, err := os.Create(destDir + "/fileB.txt")
			require.NoError(t, err)
			defer os.Remove(fileBDest.Name())

			fileBFetch, err := os.Create("/tmp/fileB.txt")
			require.NoError(t, err)
			defer os.Remove(fileBFetch.Name())

			_, err = fileBFetch.Write([]byte{1})
			require.NoError(t, err)

			// zip fetchDir to use as our unarchive source
			zipFile := "/tmp/unarchive_test_zip.zip"
			err = zipFiles(fileBFetch.Name(), zipFile)
			require.NoError(t, err)

			u := &Unarchive{
				Source:      zipFile,
				Destination: destDir,
				fetchLoc:    tmpFetchDir,
			}

			u.fetch = fetch.Fetch{
				Source:      u.Source,
				Destination: u.fetchLoc,
				HashType:    u.HashType,
				Hash:        u.Hash,
				Unarchive:   true,
			}
			defer os.Remove(u.Source)
			defer os.RemoveAll(u.Destination)

			checksumDest, err := u.getChecksum(destDir + "/fileB.txt")
			require.NoError(t, err)
			checksumFetch, err := u.getChecksum("/tmp/fileB.txt")
			require.NoError(t, err)
			require.NotEqual(t, checksumDest, checksumFetch)

			status, err := u.Apply(context.Background())

			assert.EqualError(t, err, fmt.Sprintf("will not replace, \"/fileB.txt\" exists at %q: checksum mismatch", u.Destination))
			assert.Equal(t, "use the \"force\" option to replace all files with checksum mismatch", status.Messages()[0])
			assert.Equal(t, resource.StatusFatal, status.StatusCode())
			assert.False(t, status.HasChanges())
		})
	})

	t.Run("success", func(t *testing.T) {
		t.Run("empty dest", func(t *testing.T) {
			destDir, err := ioutil.TempDir("", "destDir_unarchive")
			require.NoError(t, err)
			defer os.RemoveAll(destDir)

			fetchDir, err := ioutil.TempDir("", "fetchDir_unarchive")
			require.NoError(t, err)
			defer os.RemoveAll(fetchDir)

			fileBFetch, err := os.Create(fetchDir + "/fileB.txt")
			require.NoError(t, err)
			defer os.Remove(fileBFetch.Name())

			// zip fetchDir to use as our unarchive source
			zipFile := "/tmp/unarchive_test_zip.zip"
			err = zipFiles(fetchDir, zipFile)
			require.NoError(t, err)

			u := &Unarchive{
				Source:      zipFile,
				Destination: destDir,
				fetchLoc:    tmpFetchDir,
			}

			u.fetch = fetch.Fetch{
				Source:      u.Source,
				Destination: u.fetchLoc,
				HashType:    u.HashType,
				Hash:        u.Hash,
				Unarchive:   true,
			}
			defer os.Remove(u.Source)
			defer os.RemoveAll(u.Destination)

			status, err := u.Apply(context.Background())

			assert.NoError(t, err)
			assert.Equal(t, fmt.Sprintf("completed fetch and unarchive %q", u.Source), status.Messages()[0])
			assert.Equal(t, u.Source, status.Diffs()["unarchive"].Original())
			assert.Equal(t, u.Destination, status.Diffs()["unarchive"].Current())
		})

		t.Run("checksum match", func(t *testing.T) {
			destDir, err := ioutil.TempDir("", "destDir_unarchive")
			require.NoError(t, err)
			defer os.RemoveAll(destDir)

			fileBDest, err := os.Create(destDir + "/fileB.txt")
			require.NoError(t, err)
			defer os.Remove(fileBDest.Name())

			fileBFetch, err := os.Create("/tmp/fileB.txt")
			require.NoError(t, err)
			defer os.Remove(fileBFetch.Name())

			// zip fetchDir to use as our unarchive source
			zipFile := "/tmp/unarchive_test_zip.zip"
			err = zipFiles(fileBFetch.Name(), zipFile)
			require.NoError(t, err)

			u := &Unarchive{
				Source:      zipFile,
				Destination: destDir,
				fetchLoc:    tmpFetchDir,
			}

			u.fetch = fetch.Fetch{
				Source:      u.Source,
				Destination: u.fetchLoc,
				HashType:    u.HashType,
				Hash:        u.Hash,
				Unarchive:   true,
			}
			defer os.Remove(u.Source)
			defer os.RemoveAll(u.Destination)

			checksumDest, err := u.getChecksum(destDir + "/fileB.txt")
			require.NoError(t, err)
			checksumFetch, err := u.getChecksum("/tmp/fileB.txt")
			require.NoError(t, err)
			require.Equal(t, checksumDest, checksumFetch)

			status, err := u.Apply(context.Background())

			assert.NoError(t, err)
			assert.Equal(t, fmt.Sprintf("completed fetch and unarchive %q", u.Source), status.Messages()[0])
			assert.Equal(t, u.Source, status.Diffs()["unarchive"].Original())
			assert.Equal(t, u.Destination, status.Diffs()["unarchive"].Current())
		})

		t.Run("checksum mismatch", func(t *testing.T) {
			destDir, err := ioutil.TempDir("", "destDir_unarchive")
			require.NoError(t, err)
			defer os.RemoveAll(destDir)

			fileBDest, err := os.Create(destDir + "/fileB.txt")
			require.NoError(t, err)
			defer os.Remove(fileBDest.Name())

			fileBFetch, err := os.Create("/tmp/fileB.txt")
			require.NoError(t, err)
			defer os.Remove(fileBFetch.Name())

			_, err = fileBFetch.Write([]byte{1})
			require.NoError(t, err)

			// zip fetchDir to use as our unarchive source
			zipFile := "/tmp/unarchive_test_zip.zip"
			err = zipFiles(fileBFetch.Name(), zipFile)
			require.NoError(t, err)

			u := &Unarchive{
				Source:      zipFile,
				Destination: destDir,
				Force:       true,
				fetchLoc:    tmpFetchDir,
			}

			u.fetch = fetch.Fetch{
				Source:      u.Source,
				Destination: u.fetchLoc,
				HashType:    u.HashType,
				Hash:        u.Hash,
				Unarchive:   true,
			}
			defer os.Remove(u.Source)
			defer os.RemoveAll(u.Destination)

			checksumDest, err := u.getChecksum(destDir + "/fileB.txt")
			require.NoError(t, err)
			checksumFetch, err := u.getChecksum("/tmp/fileB.txt")
			require.NoError(t, err)
			require.NotEqual(t, checksumDest, checksumFetch)

			status, testErr := u.Apply(context.Background())

			checksumDest, err = u.getChecksum(destDir + "/fileB.txt")
			require.NoError(t, err)
			checksumFetch, err = u.getChecksum("/tmp/fileB.txt")
			require.NoError(t, err)

			assert.NoError(t, testErr)
			assert.Equal(t, fmt.Sprintf("completed fetch and unarchive %q", u.Source), status.Messages()[0])
			assert.Equal(t, u.Source, status.Diffs()["unarchive"].Original())
			assert.Equal(t, u.Destination, status.Diffs()["unarchive"].Current())
			assert.Equal(t, checksumDest, checksumFetch)
		})

		t.Run("fetch with checksum", func(t *testing.T) {
			destDir, err := ioutil.TempDir("", "destDir_unarchive")
			require.NoError(t, err)
			defer os.RemoveAll(destDir)

			fileBDest, err := os.Create(destDir + "/fileB.txt")
			require.NoError(t, err)
			defer os.Remove(fileBDest.Name())

			fileBFetch, err := os.Create("/tmp/fileB.txt")
			require.NoError(t, err)
			defer os.Remove(fileBFetch.Name())

			// zip fetchDir to use as our unarchive source
			zipFile := "/tmp/unarchive_test_zip.zip"
			err = zipFiles(fileBFetch.Name(), zipFile)
			require.NoError(t, err)

			u := &Unarchive{
				Source:      zipFile,
				Destination: destDir,
				fetchLoc:    tmpFetchDir,
			}

			srcChecksum, err := u.getChecksum(zipFile)
			require.NoError(t, err)

			u.HashType = "sha256"
			u.Hash = srcChecksum

			u.fetch = fetch.Fetch{
				Source:      u.Source,
				Destination: u.fetchLoc,
				HashType:    u.HashType,
				Hash:        u.Hash,
				Unarchive:   true,
			}
			defer os.Remove(u.Source)
			defer os.RemoveAll(u.Destination)

			status, err := u.Apply(context.Background())

			assert.NoError(t, err)
			assert.Equal(t, fmt.Sprintf("completed fetch and unarchive %q", u.Source), status.Messages()[0])
			assert.Equal(t, u.Source, status.Diffs()["unarchive"].Original())
			assert.Equal(t, u.Destination, status.Diffs()["unarchive"].Current())
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

		assert.EqualError(t, err, "cannot unarchive: stat : no such file or directory")
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

		err = setupSetDirsAndContents(u, false)
		require.NoError(t, err)
		defer os.RemoveAll(u.fetchLoc)

		expF := [2]string{u.fetchLoc, "fetchFileA.txt"}
		expD := [1]string{emptyDir}

		evalDups, err := u.setDirsAndContents()

		assert.NoError(t, err)
		assert.False(t, evalDups)
		assert.Equal(t, 2, len(u.fetchContents))
		assert.Equal(t, 1, len(u.destContents))
		assert.True(t, strings.Contains(u.fetchContents[0], expF[0]) || strings.Contains(u.fetchContents[1], expF[0]))
		assert.True(t, strings.Contains(u.fetchContents[0], expF[1]) || strings.Contains(u.fetchContents[1], expF[1]))
		assert.True(t, strings.Contains(u.destContents[0], expD[0]))
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

		err = setupSetDirsAndContents(u, false)
		require.NoError(t, err)
		defer os.RemoveAll(u.fetchLoc)

		expF := [2]string{u.fetchLoc, "fetchFileA.txt"}
		expD := [3]string{destDir, nestedDir, nestedFile.Name()}

		evalDups, err := u.setDirsAndContents()

		assert.NoError(t, err)
		assert.True(t, evalDups)
		assert.Equal(t, 2, len(u.fetchContents))
		assert.Equal(t, 3, len(u.destContents))
		assert.True(t, strings.Contains(u.fetchContents[0], expF[0]) || strings.Contains(u.fetchContents[1], expF[0]))
		assert.True(t, strings.Contains(u.fetchContents[0], expF[1]) || strings.Contains(u.fetchContents[1], expF[1]))
		assert.True(t, strings.Contains(u.destContents[0], expD[0]) || strings.Contains(u.destContents[1], expD[0]) || strings.Contains(u.destContents[2], expD[0]))
		assert.True(t, strings.Contains(u.destContents[0], expD[1]) || strings.Contains(u.destContents[1], expD[1]) || strings.Contains(u.destContents[2], expD[1]))
		assert.True(t, strings.Contains(u.destContents[0], expD[2]) || strings.Contains(u.destContents[1], expD[2]) || strings.Contains(u.destContents[2], expD[2]))
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

		err = setupSetDirsAndContents(u, true)
		require.NoError(t, err)
		defer os.RemoveAll(u.fetchLoc)

		expF := [4]string{u.fetchLoc, "/unarchive_fetch_nest", "fetchFileB.txt", "fetchFileC.txt"}
		expD := [3]string{destDir, nestedDir, nestedFile.Name()}

		evalDups, err := u.setDirsAndContents()

		assert.NoError(t, err)
		assert.True(t, evalDups)
		assert.Equal(t, 4, len(u.fetchContents))
		assert.Equal(t, 3, len(u.destContents))
		assert.True(t, strings.Contains(u.fetchContents[0], expF[0]) || strings.Contains(u.fetchContents[1], expF[0]) || strings.Contains(u.fetchContents[2], expF[0]) || strings.Contains(u.fetchContents[3], expF[0]))
		assert.True(t, strings.Contains(u.fetchContents[0], expF[1]) || strings.Contains(u.fetchContents[1], expF[1]) || strings.Contains(u.fetchContents[2], expF[1]) || strings.Contains(u.fetchContents[3], expF[1]))
		assert.True(t, strings.Contains(u.fetchContents[0], expF[2]) || strings.Contains(u.fetchContents[1], expF[2]) || strings.Contains(u.fetchContents[2], expF[2]) || strings.Contains(u.fetchContents[3], expF[2]))
		assert.True(t, strings.Contains(u.fetchContents[0], expF[3]) || strings.Contains(u.fetchContents[1], expF[3]) || strings.Contains(u.fetchContents[2], expF[3]) || strings.Contains(u.fetchContents[3], expF[3]))
		assert.True(t, strings.Contains(u.destContents[0], expD[0]) || strings.Contains(u.destContents[1], expD[0]) || strings.Contains(u.destContents[2], expD[0]))
		assert.True(t, strings.Contains(u.destContents[0], expD[1]) || strings.Contains(u.destContents[1], expD[1]) || strings.Contains(u.destContents[2], expD[1]))
		assert.True(t, strings.Contains(u.destContents[0], expD[2]) || strings.Contains(u.destContents[1], expD[2]) || strings.Contains(u.destContents[2], expD[2]))
	})
}

// TestEvaluateDuplicates tests evaluateDuplicates for Unarchive
func TestEvaluateDuplicates(t *testing.T) {
	t.Parallel()

	t.Run("no duplicates", func(t *testing.T) {
		destDir, err := ioutil.TempDir("", "destDir_unarchive")
		require.NoError(t, err)
		defer os.RemoveAll(destDir)

		fileA, err := ioutil.TempFile(destDir, "fileA.txt")
		require.NoError(t, err)
		defer os.Remove(fileA.Name())

		fileB, err := ioutil.TempFile(destDir, "fileB.txt")
		require.NoError(t, err)
		defer os.Remove(fileB.Name())

		ddInfo, err := os.Open(destDir)
		require.NoError(t, err)

		fetchDir, err := ioutil.TempDir("", "fetchDir_unarchive")
		require.NoError(t, err)
		defer os.RemoveAll(fetchDir)

		fileC, err := ioutil.TempFile(fetchDir, "fileC.txt")
		require.NoError(t, err)
		defer os.Remove(fileC.Name())

		fileD, err := ioutil.TempFile(fetchDir, "fileD.txt")
		require.NoError(t, err)
		defer os.Remove(fileD.Name())

		fdInfo, err := os.Open(fetchDir)
		require.NoError(t, err)

		u := &Unarchive{
			destDir:  ddInfo,
			fetchDir: fdInfo,
		}

		err = filepath.Walk(u.destDir.Name(), func(path string, f os.FileInfo, err error) error {
			u.destContents = append(u.destContents, path)
			return nil
		})
		require.NoError(t, err)

		err = filepath.Walk(u.fetchDir.Name(), func(path string, f os.FileInfo, err error) error {
			u.fetchContents = append(u.fetchContents, path)
			return nil
		})
		require.NoError(t, err)

		err = u.evaluateDuplicates()

		assert.NoError(t, err)
	})

	t.Run("duplicates", func(t *testing.T) {
		destDir, err := ioutil.TempDir("", "destDir_unarchive")
		require.NoError(t, err)
		defer os.RemoveAll(destDir)

		fileADest, err := os.Create(destDir + "/fileA.txt")
		require.NoError(t, err)
		defer os.Remove(fileADest.Name())

		ddInfo, err := os.Open(destDir)
		require.NoError(t, err)

		fetchDir, err := ioutil.TempDir("", "fetchDir_unarchive")
		require.NoError(t, err)
		defer os.RemoveAll(fetchDir)

		fileAFetch, err := os.Create(fetchDir + "/fileA.txt")
		require.NoError(t, err)
		defer os.Remove(fileAFetch.Name())

		fdInfo, err := os.Open(fetchDir)
		require.NoError(t, err)

		t.Run("checksum match", func(t *testing.T) {
			u := &Unarchive{
				destDir:  ddInfo,
				fetchDir: fdInfo,
			}

			checksumDest, err := u.getChecksum(destDir + "/fileA.txt")
			require.NoError(t, err)
			checksumFetch, err := u.getChecksum(fetchDir + "/fileA.txt")
			require.NoError(t, err)
			require.Equal(t, checksumDest, checksumFetch)

			err = filepath.Walk(u.destDir.Name(), func(path string, f os.FileInfo, err error) error {
				u.destContents = append(u.destContents, path)
				return nil
			})
			require.NoError(t, err)

			err = filepath.Walk(u.fetchDir.Name(), func(path string, f os.FileInfo, err error) error {
				u.fetchContents = append(u.fetchContents, path)
				return nil
			})
			require.NoError(t, err)

			err = u.evaluateDuplicates()

			assert.NoError(t, err)
		})

		t.Run("checksum mismatch", func(t *testing.T) {
			t.Run("different size", func(t *testing.T) {
				fileBDest, err := os.Create(destDir + "/fileB.txt")
				require.NoError(t, err)
				defer os.Remove(fileBDest.Name())

				fileBFetch, err := os.Create(fetchDir + "/fileB.txt")
				require.NoError(t, err)
				defer os.Remove(fileBFetch.Name())

				_, err = fileBFetch.Write([]byte{1})
				require.NoError(t, err)

				dStat, _ := os.Stat(destDir + "/fileB.txt")
				fStat, _ := os.Stat(fetchDir + "/fileB.txt")
				require.NotEqual(t, dStat.Size(), fStat.Size())

				u := &Unarchive{
					Destination: destDir,
					destDir:     ddInfo,
					fetchDir:    fdInfo,
				}

				checksumDest, err := u.getChecksum(destDir + "/fileB.txt")
				require.NoError(t, err)
				checksumFetch, err := u.getChecksum(fetchDir + "/fileB.txt")
				require.NoError(t, err)
				require.NotEqual(t, checksumDest, checksumFetch)

				err = filepath.Walk(u.destDir.Name(), func(path string, f os.FileInfo, err error) error {
					u.destContents = append(u.destContents, path)
					return nil
				})
				require.NoError(t, err)

				err = filepath.Walk(u.fetchDir.Name(), func(path string, f os.FileInfo, err error) error {
					u.fetchContents = append(u.fetchContents, path)
					return nil
				})
				require.NoError(t, err)

				err = u.evaluateDuplicates()

				assert.EqualError(t, err, fmt.Sprintf("will not replace, \"/fileB.txt\" exists at %q: checksum mismatch", u.Destination))
			})

			t.Run("same size", func(t *testing.T) {
				fileBDest, err := os.Create(destDir + "/fileB.txt")
				require.NoError(t, err)
				defer os.Remove(fileBDest.Name())

				_, err = fileBDest.WriteString("a")
				require.NoError(t, err)

				fileBFetch, err := os.Create(fetchDir + "/fileB.txt")
				require.NoError(t, err)
				defer os.Remove(fileBFetch.Name())

				_, err = fileBFetch.WriteString("b")
				require.NoError(t, err)

				dStat, _ := os.Stat(destDir + "/fileB.txt")
				fStat, _ := os.Stat(fetchDir + "/fileB.txt")
				require.Equal(t, dStat.Size(), fStat.Size())

				u := &Unarchive{
					Destination: destDir,
					destDir:     ddInfo,
					fetchDir:    fdInfo,
				}

				checksumDest, err := u.getChecksum(destDir + "/fileB.txt")
				require.NoError(t, err)
				checksumFetch, err := u.getChecksum(fetchDir + "/fileB.txt")
				require.NoError(t, err)
				require.NotEqual(t, checksumDest, checksumFetch)

				err = filepath.Walk(u.destDir.Name(), func(path string, f os.FileInfo, err error) error {
					u.destContents = append(u.destContents, path)
					return nil
				})
				require.NoError(t, err)

				err = filepath.Walk(u.fetchDir.Name(), func(path string, f os.FileInfo, err error) error {
					u.fetchContents = append(u.fetchContents, path)
					return nil
				})
				require.NoError(t, err)

				err = u.evaluateDuplicates()

				assert.EqualError(t, err, fmt.Sprintf("will not replace, \"/fileB.txt\" exists at %q: checksum mismatch", u.Destination))
			})
		})
	})

	t.Run("recurse", func(t *testing.T) {
		destDir, err := ioutil.TempDir("", "destDir_unarchive")
		require.NoError(t, err)
		defer os.RemoveAll(destDir)

		stat, err := os.Stat(destDir)
		require.NoError(t, err)

		err = os.Mkdir(destDir+"/dirA", stat.Mode().Perm())
		require.NoError(t, err)

		err = os.Mkdir(destDir+"/dirA/dirB", stat.Mode().Perm())
		require.NoError(t, err)

		fileBDest, err := os.Create(destDir + "/dirA/dirB/fileB.txt")
		require.NoError(t, err)
		defer os.Remove(fileBDest.Name())

		ddInfo, err := os.Open(destDir)
		require.NoError(t, err)

		fetchDir, err := ioutil.TempDir("", "fetchDir_unarchive")
		require.NoError(t, err)
		defer os.RemoveAll(fetchDir)

		stat, err = os.Stat(fetchDir)
		require.NoError(t, err)

		err = os.Mkdir(fetchDir+"/dirA", stat.Mode().Perm())
		require.NoError(t, err)

		err = os.Mkdir(fetchDir+"/dirA/dirB", stat.Mode().Perm())
		require.NoError(t, err)

		fileBFetch, err := os.Create(fetchDir + "/dirA/dirB/fileB.txt")
		require.NoError(t, err)
		defer os.Remove(fileBFetch.Name())

		fdInfo, err := os.Open(fetchDir)
		require.NoError(t, err)

		t.Run("checksum match", func(t *testing.T) {
			u := &Unarchive{
				Destination: destDir,
				destDir:     ddInfo,
				fetchDir:    fdInfo,
			}

			checksumDest, err := u.getChecksum(destDir + "/dirA/dirB/fileB.txt")
			require.NoError(t, err)
			checksumFetch, err := u.getChecksum(fetchDir + "/dirA/dirB/fileB.txt")
			require.NoError(t, err)
			require.Equal(t, checksumDest, checksumFetch)

			err = filepath.Walk(u.destDir.Name(), func(path string, f os.FileInfo, err error) error {
				u.destContents = append(u.destContents, path)
				return nil
			})
			require.NoError(t, err)

			err = filepath.Walk(u.fetchDir.Name(), func(path string, f os.FileInfo, err error) error {
				u.fetchContents = append(u.fetchContents, path)
				return nil
			})
			require.NoError(t, err)

			err = u.evaluateDuplicates()

			assert.NoError(t, err)
		})

		t.Run("checksum mismatch", func(t *testing.T) {
			fileCDest, err := os.Create(destDir + "/dirA/dirB/fileC.txt")
			require.NoError(t, err)
			defer os.Remove(fileCDest.Name())

			fileCFetch, err := os.Create(fetchDir + "/dirA/dirB/fileC.txt")
			require.NoError(t, err)
			defer os.Remove(fileCFetch.Name())

			_, err = fileCFetch.Write([]byte{2})
			require.NoError(t, err)

			u := &Unarchive{
				Destination: destDir,
				destDir:     ddInfo,
				fetchDir:    fdInfo,
			}

			checksumDest, err := u.getChecksum(destDir + "/dirA/dirB/fileC.txt")
			require.NoError(t, err)
			checksumFetch, err := u.getChecksum(fetchDir + "/dirA/dirB/fileC.txt")
			require.NoError(t, err)
			require.NotEqual(t, checksumDest, checksumFetch)

			err = filepath.Walk(u.destDir.Name(), func(path string, f os.FileInfo, err error) error {
				u.destContents = append(u.destContents, path)
				return nil
			})
			require.NoError(t, err)

			err = filepath.Walk(u.fetchDir.Name(), func(path string, f os.FileInfo, err error) error {
				u.fetchContents = append(u.fetchContents, path)
				return nil
			})
			require.NoError(t, err)

			err = u.evaluateDuplicates()

			assert.EqualError(t, err, fmt.Sprintf("will not replace, \"/dirA/dirB/fileC.txt\" exists at %q: checksum mismatch", u.Destination))
		})
	})
}

// TestCopyToFinalDest tests copyToFinalDest for Unarchive
func TestCopyToFinalDest(t *testing.T) {
	t.Parallel()

	t.Run("no duplicates", func(t *testing.T) {
		destDir, err := ioutil.TempDir("", "destDir_unarchive")
		require.NoError(t, err)
		defer os.RemoveAll(destDir)

		fileA, err := ioutil.TempFile(destDir, "fileA.txt")
		require.NoError(t, err)
		defer os.Remove(fileA.Name())

		fileB, err := ioutil.TempFile(destDir, "fileB.txt")
		require.NoError(t, err)
		defer os.Remove(fileB.Name())

		ddInfo, err := os.Open(destDir)
		require.NoError(t, err)

		fetchDir, err := ioutil.TempDir("", "fetchDir_unarchive")
		require.NoError(t, err)
		defer os.RemoveAll(fetchDir)

		fileC, err := ioutil.TempFile(fetchDir, "fileC.txt")
		require.NoError(t, err)
		defer os.Remove(fileC.Name())

		fileD, err := ioutil.TempFile(fetchDir, "fileD.txt")
		require.NoError(t, err)
		defer os.Remove(fileD.Name())

		fdInfo, err := os.Open(fetchDir)
		require.NoError(t, err)

		u := &Unarchive{
			destDir:  ddInfo,
			fetchDir: fdInfo,
		}

		err = filepath.Walk(u.destDir.Name(), func(path string, f os.FileInfo, err error) error {
			u.destContents = append(u.destContents, path)
			return nil
		})
		require.NoError(t, err)

		err = filepath.Walk(u.fetchDir.Name(), func(path string, f os.FileInfo, err error) error {
			u.fetchContents = append(u.fetchContents, path)
			return nil
		})
		require.NoError(t, err)

		exp := []string{destDir, "fileA.txt", "fileB.txt", "fileC.txt", "fileD.txt"}

		testErr := u.copyToFinalDest()

		act := []string{}
		err = filepath.Walk(u.destDir.Name(), func(path string, f os.FileInfo, err error) error {
			act = append(act, path)
			return nil
		})
		require.NoError(t, err)

		assert.NoError(t, testErr)
		assert.Equal(t, 5, len(act))
		assert.True(t, strings.Contains(act[0], exp[0]) || strings.Contains(act[1], exp[0]) || strings.Contains(act[2], exp[0]) || strings.Contains(act[3], exp[0]) || strings.Contains(act[4], exp[0]))
		assert.True(t, strings.Contains(act[0], exp[1]) || strings.Contains(act[1], exp[1]) || strings.Contains(act[2], exp[1]) || strings.Contains(act[3], exp[1]) || strings.Contains(act[4], exp[1]))
		assert.True(t, strings.Contains(act[0], exp[2]) || strings.Contains(act[1], exp[2]) || strings.Contains(act[2], exp[2]) || strings.Contains(act[3], exp[2]) || strings.Contains(act[4], exp[2]))
		assert.True(t, strings.Contains(act[0], exp[3]) || strings.Contains(act[1], exp[3]) || strings.Contains(act[2], exp[3]) || strings.Contains(act[3], exp[3]) || strings.Contains(act[4], exp[3]))
		assert.True(t, strings.Contains(act[0], exp[4]) || strings.Contains(act[1], exp[4]) || strings.Contains(act[2], exp[4]) || strings.Contains(act[3], exp[4]) || strings.Contains(act[4], exp[4]))
	})

	t.Run("duplicates", func(t *testing.T) {
		destDir, err := ioutil.TempDir("", "destDir_unarchive")
		require.NoError(t, err)
		defer os.RemoveAll(destDir)

		fileADest, err := os.Create(destDir + "/fileA.txt")
		require.NoError(t, err)
		defer os.Remove(fileADest.Name())

		ddInfo, err := os.Open(destDir)
		require.NoError(t, err)

		fetchDir, err := ioutil.TempDir("", "fetchDir_unarchive")
		require.NoError(t, err)
		defer os.RemoveAll(fetchDir)

		fileAFetch, err := os.Create(fetchDir + "/fileA.txt")
		require.NoError(t, err)
		defer os.Remove(fileAFetch.Name())

		fdInfo, err := os.Open(fetchDir)
		require.NoError(t, err)

		t.Run("checksum match", func(t *testing.T) {
			u := &Unarchive{
				destDir:  ddInfo,
				fetchDir: fdInfo,
			}

			checksumDest, err := u.getChecksum(destDir + "/fileA.txt")
			require.NoError(t, err)
			checksumFetch, err := u.getChecksum(fetchDir + "/fileA.txt")
			require.NoError(t, err)
			require.Equal(t, checksumDest, checksumFetch)

			err = filepath.Walk(u.destDir.Name(), func(path string, f os.FileInfo, err error) error {
				u.destContents = append(u.destContents, path)
				return nil
			})
			require.NoError(t, err)

			err = filepath.Walk(u.fetchDir.Name(), func(path string, f os.FileInfo, err error) error {
				u.fetchContents = append(u.fetchContents, path)
				return nil
			})
			require.NoError(t, err)

			exp := []string{destDir, "fileA.txt"}

			testErr := u.copyToFinalDest()

			act := []string{}
			err = filepath.Walk(u.destDir.Name(), func(path string, f os.FileInfo, err error) error {
				act = append(act, path)
				return nil
			})
			require.NoError(t, err)

			assert.NoError(t, testErr)
			assert.Equal(t, 2, len(act))
			assert.True(t, strings.Contains(act[0], exp[0]) || strings.Contains(act[1], exp[0]))
			assert.True(t, strings.Contains(act[0], exp[1]) || strings.Contains(act[1], exp[1]))
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
				Destination: destDir,
				destDir:     ddInfo,
				fetchDir:    fdInfo,
			}

			checksumDest, err := u.getChecksum(destDir + "/fileB.txt")
			require.NoError(t, err)
			checksumFetch, err := u.getChecksum(fetchDir + "/fileB.txt")
			require.NoError(t, err)
			require.NotEqual(t, checksumDest, checksumFetch)

			err = filepath.Walk(u.destDir.Name(), func(path string, f os.FileInfo, err error) error {
				u.destContents = append(u.destContents, path)
				return nil
			})
			require.NoError(t, err)

			err = filepath.Walk(u.fetchDir.Name(), func(path string, f os.FileInfo, err error) error {
				u.fetchContents = append(u.fetchContents, path)
				return nil
			})
			require.NoError(t, err)

			exp := []string{destDir, "fileA.txt", "fileB.txt"}

			testErr := u.copyToFinalDest()

			act := []string{}
			err = filepath.Walk(u.destDir.Name(), func(path string, f os.FileInfo, err error) error {
				act = append(act, path)
				return nil
			})
			require.NoError(t, err)

			checksumDest, err = u.getChecksum(destDir + "/fileB.txt")
			require.NoError(t, err)
			checksumFetch, err = u.getChecksum(fetchDir + "/fileB.txt")
			require.NoError(t, err)

			assert.NoError(t, testErr)
			assert.Equal(t, 3, len(act))
			assert.True(t, strings.Contains(act[0], exp[0]) || strings.Contains(act[1], exp[0]) || strings.Contains(act[2], exp[0]))
			assert.True(t, strings.Contains(act[0], exp[1]) || strings.Contains(act[1], exp[1]) || strings.Contains(act[2], exp[1]))
			assert.True(t, strings.Contains(act[0], exp[2]) || strings.Contains(act[1], exp[2]) || strings.Contains(act[2], exp[2]))
			assert.Equal(t, checksumDest, checksumFetch)
		})
	})

	t.Run("recurse", func(t *testing.T) {
		destDir, err := ioutil.TempDir("", "destDir_unarchive")
		require.NoError(t, err)
		defer os.RemoveAll(destDir)

		ddInfo, err := os.Open(destDir)
		require.NoError(t, err)

		fetchDir, err := ioutil.TempDir("", "fetchDir_unarchive")
		require.NoError(t, err)
		defer os.RemoveAll(fetchDir)

		dirA, err := ioutil.TempDir(fetchDir, "dirA")
		require.NoError(t, err)
		defer os.RemoveAll(dirA)

		dirB, err := ioutil.TempDir(dirA, "dirB")
		require.NoError(t, err)
		defer os.RemoveAll(dirB)

		dirC, err := ioutil.TempDir(dirB, "dirC")
		require.NoError(t, err)
		defer os.RemoveAll(dirC)

		fileC, err := os.Create(dirC + "/fileC.txt")
		require.NoError(t, err)
		defer os.Remove(fileC.Name())

		fdInfo, err := os.Open(fetchDir)
		require.NoError(t, err)

		u := &Unarchive{
			Destination: destDir,
			destDir:     ddInfo,
			fetchDir:    fdInfo,
		}

		err = filepath.Walk(u.destDir.Name(), func(path string, f os.FileInfo, err error) error {
			u.destContents = append(u.destContents, path)
			return nil
		})
		require.NoError(t, err)

		err = filepath.Walk(u.fetchDir.Name(), func(path string, f os.FileInfo, err error) error {
			u.fetchContents = append(u.fetchContents, path)
			return nil
		})
		require.NoError(t, err)

		exp := []string{destDir, "dirA", "dirB", "dirC", "fileC.txt"}

		testErr := u.copyToFinalDest()

		act := []string{}
		err = filepath.Walk(u.destDir.Name(), func(path string, f os.FileInfo, err error) error {
			act = append(act, path)
			return nil
		})
		require.NoError(t, err)

		assert.NoError(t, testErr)
		assert.Equal(t, 5, len(act))
		assert.True(t, strings.Contains(act[0], exp[0]) || strings.Contains(act[1], exp[0]) || strings.Contains(act[2], exp[0]) || strings.Contains(act[3], exp[0]) || strings.Contains(act[4], exp[0]))
		assert.True(t, strings.Contains(act[0], exp[1]) || strings.Contains(act[1], exp[1]) || strings.Contains(act[2], exp[1]) || strings.Contains(act[3], exp[1]) || strings.Contains(act[4], exp[1]))
		assert.True(t, strings.Contains(act[0], exp[2]) || strings.Contains(act[1], exp[2]) || strings.Contains(act[2], exp[2]) || strings.Contains(act[3], exp[2]) || strings.Contains(act[4], exp[2]))
		assert.True(t, strings.Contains(act[0], exp[3]) || strings.Contains(act[1], exp[3]) || strings.Contains(act[2], exp[3]) || strings.Contains(act[3], exp[3]) || strings.Contains(act[4], exp[3]))
		assert.True(t, strings.Contains(act[0], exp[4]) || strings.Contains(act[1], exp[4]) || strings.Contains(act[2], exp[4]) || strings.Contains(act[3], exp[4]) || strings.Contains(act[4], exp[4]))
	})
}

// TestSetFetchLoc tests setFetchLoc for Unarchive
func TestSetFetchLoc(t *testing.T) {
	t.Parallel()

	srcFile, err := ioutil.TempFile("", "unarchive_test.txt")
	require.NoError(t, err)
	defer os.Remove(srcFile.Name())

	u := &Unarchive{
		Source: srcFile.Name(),
	}

	tmp := os.TempDir()
	s := tmp[len(tmp)-1:]
	expected := tmp + "tmpDirFetch"
	if s != "/" {
		expected = tmp + "/tmpDirFetch"
	}

	err = u.setFetchLoc()
	defer os.RemoveAll(u.fetchLoc)

	assert.NoError(t, err)
	assert.Contains(t, u.fetchLoc, expected)
}

// setupSetDirsAndContents performs some setup required to test
// SetDirsAndContents
func setupSetDirsAndContents(u *Unarchive, nested bool) error {

	err := u.setFetchLoc()
	if err != nil {
		return err
	}

	err = fetchApply(u.fetchLoc, nested)

	return err
}

// fetchApply sets up a temporary fetch location with file(s) based on the
// nested flag
func fetchApply(fetchLoc string, nested bool) error {

	if !nested {
		_, err := ioutil.TempFile(fetchLoc, "fetchFileA.txt")
		if err != nil {
			return err
		}
	} else {
		nestedFetchDir, err := ioutil.TempDir(fetchLoc, "unarchive_fetch_nest")
		if err != nil {
			return err
		}

		_, err = ioutil.TempFile(nestedFetchDir, "fetchFileB.txt")
		if err != nil {
			return err
		}

		_, err = ioutil.TempFile(nestedFetchDir, "fetchFileC.txt")
		if err != nil {
			return err
		}
	}

	return nil
}

// zipFiles zips the files in source and places them in destination
func zipFiles(source, destination string) error {
	base := ""

	zipFile, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	w := zip.NewWriter(zipFile)
	defer w.Close()

	f, err := os.Stat(source)
	if err != nil {
		return err
	}

	if f.IsDir() {
		base = filepath.Base(source)
	}

	err = filepath.Walk(source, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(f)
		if err != nil {
			return err
		}

		if base != "" {
			header.Name = filepath.Join(base, strings.TrimPrefix(path, source))
		}

		if f.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := w.CreateHeader(header)
		if err != nil {
			return err
		}

		if f.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(writer, file)
		return err
	})
	if err != nil {
		return err
	}

	return nil
}
