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

package fetch_test

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/asteris-llc/converge/helpers/fakerenderer"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/file/fetch"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
)

// TestFetchInterface tests that Fetch is properly implemented
func TestFetchInterface(t *testing.T) {
	t.Parallel()

	assert.Implements(t, (*resource.Task)(nil), new(fetch.Fetch))
}

// TestCheck tests the cases Check handles
func TestCheck(t *testing.T) {
	t.Parallel()

	t.Run("hash error", func(t *testing.T) {
		task := fetch.Fetch{
			Source:      "https://github.com/asteris-llc/converge/releases/download/0.2.0/converge_0.2.0_darwin_amd64.tar.gz",
			Destination: "/tmp/converge.tar.gz",
			HashType:    "invalid",
			Hash:        "notarealhashbutstringnonetheless",
		}

		status, err := task.Check(context.Background(), fakerenderer.New())

		assert.EqualError(t, err, fmt.Sprintf("will not attempt file fetch: unsupported hashType %q", task.HashType))
		assert.True(t, status.HasChanges())
	})

	t.Run("fetch new file", func(t *testing.T) {
		t.Run("with checksum", func(t *testing.T) {
			task := fetch.Fetch{
				Source:      "https://github.com/asteris-llc/converge/releases/download/0.2.0/converge_0.2.0_darwin_amd64.tar.gz",
				Destination: "/tmp/converge.tar.gz",
				HashType:    string(fetch.HashMD5),
				Hash:        "notarealhashbutstringnonetheless",
			}

			status, err := task.Check(context.Background(), fakerenderer.New())

			assert.NoError(t, err)
			assert.Equal(t, "<absent>", status.Diffs()["destination"].Original())
			assert.Equal(t, task.Destination, status.Diffs()["destination"].Current())
			assert.True(t, status.HasChanges())
		})

		t.Run("no checksum", func(t *testing.T) {
			task := fetch.Fetch{
				Source:      "https://github.com/asteris-llc/converge/releases/download/0.2.0/converge_0.2.0_darwin_amd64.tar.gz",
				Destination: "/tmp/converge.tar.gz",
			}

			status, err := task.Check(context.Background(), fakerenderer.New())

			assert.NoError(t, err)
			assert.Equal(t, "<absent>", status.Diffs()["destination"].Original())
			assert.Equal(t, task.Destination, status.Diffs()["destination"].Current())
			assert.True(t, status.HasChanges())
		})
	})
}

// TestApply tests the cases Apply handles
func TestApply(t *testing.T) {
	t.Parallel()

	t.Run("hash error", func(t *testing.T) {
		task := fetch.Fetch{
			Source:      "https://github.com/asteris-llc/converge/releases/download/0.2.0/converge_0.2.0_darwin_amd64.tar.gz",
			Destination: "/tmp/converge.tar.gz",
			HashType:    "invalid",
			Hash:        "notarealhashbutstringnonetheless",
		}
		defer os.Remove(task.Destination)

		status, err := task.Apply(context.Background())

		assert.EqualError(t, err, fmt.Sprintf("will not attempt file fetch: unsupported hashType %q", task.HashType))
		assert.True(t, status.HasChanges())
	})

	t.Run("source error", func(t *testing.T) {
		m := &MockDiff{}
		task := fetch.Fetch{
			Source:      ":test",
			Destination: "/tmp/fetch_test.txt",
			Force:       true,
		}
		defer os.Remove(task.Destination)

		stat := resource.NewStatus()
		stat.RaiseLevel(resource.StatusWillChange)
		m.On("DiffFile", nil, stat).Return(stat, nil)

		status, err := task.Apply(context.Background())

		assert.EqualError(t, err, fmt.Sprintf("could not parse source: parse %s: missing protocol scheme", task.Source))
		assert.False(t, status.HasChanges())
	})

	t.Run("failed to fetch", func(t *testing.T) {
		m := &MockDiff{}
		task := fetch.Fetch{
			Source:      "",
			Destination: "/tmp/fetch_test.txt",
			Force:       true,
		}
		defer os.Remove(task.Destination)

		stat := resource.NewStatus()
		m.On("DiffFile", nil, stat).Return(stat, nil)

		status, err := task.Apply(context.Background())

		assert.EqualError(t, err, "failed to fetch: source path must be a file")
		assert.False(t, status.HasChanges())
	})

	t.Run("with checksum", func(t *testing.T) {
		t.Run("file exists", func(t *testing.T) {
			src, err := ioutil.TempFile("", "fetch_test.txt")
			require.NoError(t, err)
			defer os.Remove(src.Name())

			dest, err := ioutil.TempFile("", "fetch_test.txt")
			require.NoError(t, err)
			defer os.Remove(dest.Name())

			hash, err := getHash(src.Name(), string(fetch.HashMD5))
			require.NoError(t, err)

			m := &MockDiff{}
			task := fetch.Fetch{
				Source:      src.Name(),
				Destination: dest.Name(),
				HashType:    string(fetch.HashMD5),
				Hash:        hex.EncodeToString(hash.Sum(nil)),
			}

			stat := resource.NewStatus()
			m.On("DiffFile", nil, stat).Return(stat, nil)

			status, err := task.Apply(context.Background())

			assert.NoError(t, err)
			assert.Contains(t, status.Messages(), "file exists")
			assert.False(t, status.HasChanges())
		})

		t.Run("force=true", func(t *testing.T) {
			src, err := ioutil.TempFile("", "fetch_test.txt")
			require.NoError(t, err)
			defer os.Remove(src.Name())

			dest, err := ioutil.TempFile("", "fetch_test.txt")
			require.NoError(t, err)
			defer os.Remove(dest.Name())

			hash, err := getHash(src.Name(), string(fetch.HashMD5))
			require.NoError(t, err)

			m := &MockDiff{}
			task := fetch.Fetch{
				Source:      src.Name(),
				Destination: dest.Name(),
				HashType:    string(fetch.HashMD5),
				Hash:        "notarealhashbutstringnonetheless",
				Force:       true,
			}
			defer os.Remove(task.Destination)

			stat := resource.NewStatus()
			m.On("DiffFile", hash, stat).Return(stat, nil)

			status, err := task.Apply(context.Background())

			assert.NoError(t, err)
			assert.Contains(t, status.Messages(), "fetched successfully")
			assert.Equal(t, hex.EncodeToString(hash.Sum(nil)), status.Diffs()["checksum"].Original())
			assert.Equal(t, task.Hash, status.Diffs()["checksum"].Current())
			assert.True(t, status.HasChanges())
		})

		t.Run("force=false", func(t *testing.T) {
			src, err := ioutil.TempFile("", "fetch_test.txt")
			require.NoError(t, err)
			defer os.Remove(src.Name())

			dest, err := ioutil.TempFile("", "fetch_test.txt")
			require.NoError(t, err)
			defer os.Remove(dest.Name())

			hash, err := getHash(src.Name(), string(fetch.HashMD5))
			require.NoError(t, err)

			m := &MockDiff{}
			task := fetch.Fetch{
				Source:      src.Name(),
				Destination: dest.Name(),
				HashType:    string(fetch.HashMD5),
				Hash:        "notarealhashbutstringnonetheless",
			}
			defer os.Remove(task.Destination)

			stat := resource.NewStatus()
			m.On("DiffFile", hash, stat).Return(stat, nil)

			status, err := task.Apply(context.Background())

			assert.EqualError(t, err, "will not attempt fetch: checksum mismatch")
			assert.Contains(t, status.Messages(), "checksum mismatch, use the \"force\" option to replace")
			assert.True(t, status.HasChanges())
		})
	})

	t.Run("no checksum", func(t *testing.T) {
		t.Run("force=true", func(t *testing.T) {
			src, err := ioutil.TempFile("", "fetch_test.txt")
			require.NoError(t, err)
			defer os.Remove(src.Name())

			dest, err := ioutil.TempFile("", "fetch_test.txt")
			require.NoError(t, err)
			defer os.Remove(dest.Name())

			m := &MockDiff{}
			task := fetch.Fetch{
				Source:      src.Name(),
				Destination: dest.Name(),
				Force:       true,
			}
			defer os.Remove(task.Destination)

			stat := resource.NewStatus()
			m.On("DiffFile", nil, stat).Return(stat, nil)

			status, err := task.Apply(context.Background())

			assert.NoError(t, err)
			assert.Contains(t, status.Messages(), "fetched successfully")
			assert.Equal(t, "<force fetch>", status.Diffs()["destination"].Original())
			assert.Equal(t, task.Destination, status.Diffs()["destination"].Current())
			assert.True(t, status.HasChanges())
		})

		t.Run("force=false", func(t *testing.T) {
			src, err := ioutil.TempFile("", "fetch_test.txt")
			require.NoError(t, err)
			defer os.Remove(src.Name())

			dest, err := ioutil.TempFile("", "fetch_test.txt")
			require.NoError(t, err)
			defer os.Remove(dest.Name())

			m := &MockDiff{}
			task := fetch.Fetch{
				Source:      src.Name(),
				Destination: dest.Name(),
			}
			defer os.Remove(task.Destination)

			stat := resource.NewStatus()
			m.On("DiffFile", nil, stat).Return(stat, nil)

			status, err := task.Apply(context.Background())

			assert.NoError(t, err)
			assert.Contains(t, status.Messages(), "file exists")
			assert.False(t, status.HasChanges())
		})
	})

	t.Run("fetch new file", func(t *testing.T) {
		t.Run("with checksum", func(t *testing.T) {
			src, err := ioutil.TempFile("", "fetch_test.txt")
			require.NoError(t, err)
			defer os.Remove(src.Name())

			hash, err := getHash(src.Name(), string(fetch.HashMD5))
			require.NoError(t, err)

			m := &MockDiff{}
			task := fetch.Fetch{
				Source:      src.Name(),
				Destination: "/tmp/fetch_test2.txt",
				HashType:    string(fetch.HashMD5),
				Hash:        hex.EncodeToString(hash.Sum(nil)),
			}
			defer os.Remove(task.Destination)

			stat := resource.NewStatus()
			m.On("DiffFile", hash, stat).Return(stat, nil)

			status, err := task.Apply(context.Background())

			assert.NoError(t, err)
			assert.Contains(t, status.Messages(), "fetched successfully")
			assert.Equal(t, "<absent>", status.Diffs()["destination"].Original())
			assert.Equal(t, task.Destination, status.Diffs()["destination"].Current())
			assert.True(t, status.HasChanges())
		})

		t.Run("no checksum", func(t *testing.T) {
			src, err := ioutil.TempFile("", "fetch_test.txt")
			require.NoError(t, err)
			defer os.Remove(src.Name())

			m := &MockDiff{}
			task := fetch.Fetch{
				Source:      src.Name(),
				Destination: "/tmp/fetch_test2.txt",
			}
			defer os.Remove(task.Destination)

			stat := resource.NewStatus()
			m.On("DiffFile", nil, stat).Return(stat, nil)

			status, err := task.Apply(context.Background())

			assert.NoError(t, err)
			assert.Contains(t, status.Messages(), "fetched successfully")
			assert.Equal(t, "<absent>", status.Diffs()["destination"].Original())
			assert.Equal(t, task.Destination, status.Diffs()["destination"].Current())
			assert.True(t, status.HasChanges())
		})
	})
}

// TestDiffFile tests DiffFile
func TestDiffFile(t *testing.T) {
	t.Parallel()

	t.Run("fetch new file", func(t *testing.T) {
		t.Run("with checksum", func(t *testing.T) {
			task := fetch.Fetch{
				Source:      "https://github.com/asteris-llc/converge/releases/download/0.2.0/converge_0.2.0_darwin_amd64.tar.gz",
				Destination: "/tmp/converge.tar.gz",
				HashType:    string(fetch.HashMD5),
				Hash:        "notarealhashbutstringnonetheless",
			}

			status, err := task.Check(context.Background(), fakerenderer.New())

			assert.NoError(t, err)
			assert.Equal(t, "<absent>", status.Diffs()["destination"].Original())
			assert.Equal(t, task.Destination, status.Diffs()["destination"].Current())
			assert.True(t, status.HasChanges())
		})

		t.Run("no checksum", func(t *testing.T) {
			task := fetch.Fetch{
				Source:      "https://github.com/asteris-llc/converge/releases/download/0.2.0/converge_0.2.0_darwin_amd64.tar.gz",
				Destination: "/tmp/converge.tar.gz",
			}

			status, err := task.Check(context.Background(), fakerenderer.New())

			assert.NoError(t, err)
			assert.Equal(t, "<absent>", status.Diffs()["destination"].Original())
			assert.Equal(t, task.Destination, status.Diffs()["destination"].Current())
			assert.True(t, status.HasChanges())
		})
	})

	t.Run("destination error-directory", func(t *testing.T) {
		t.Run("with checksum", func(t *testing.T) {
			task := fetch.Fetch{
				Source:      "https://github.com/asteris-llc/converge/releases/download/0.2.0/converge_0.2.0_darwin_amd64.tar.gz",
				Destination: "/tmp",
				HashType:    string(fetch.HashMD5),
				Hash:        "notarealhashbutstringnonetheless",
			}

			status, err := task.Check(context.Background(), fakerenderer.New())

			assert.EqualError(t, err, fmt.Sprintf("invalid destination \"%s\", cannot be directory", task.Destination))
			assert.True(t, status.HasChanges())
		})

		t.Run("no checksum", func(t *testing.T) {
			task := fetch.Fetch{
				Source:      "https://github.com/asteris-llc/converge/releases/download/0.2.0/converge_0.2.0_darwin_amd64.tar.gz",
				Destination: "/tmp",
			}

			status, err := task.Check(context.Background(), fakerenderer.New())

			assert.EqualError(t, err, fmt.Sprintf("invalid destination \"%s\", cannot be directory", task.Destination))
			assert.True(t, status.HasChanges())
		})
	})

	t.Run("checksums differ", func(t *testing.T) {
		t.Run("force=true", func(t *testing.T) {
			src, err := ioutil.TempFile("", "fetch_test.txt")
			require.NoError(t, err)
			defer os.Remove(src.Name())

			dest, err := ioutil.TempFile("", "fetch_test.txt")
			require.NoError(t, err)
			defer os.Remove(dest.Name())

			hash, err := getHash(src.Name(), string(fetch.HashMD5))
			require.NoError(t, err)

			task := fetch.Fetch{
				Source:      src.Name(),
				Destination: dest.Name(),
				HashType:    string(fetch.HashMD5),
				Hash:        "notarealhashbutstringnonetheless",
				Force:       true,
			}

			status, err := task.DiffFile(resource.NewStatus(), hash)

			assert.NoError(t, err)
			assert.Contains(t, status.Messages(), "checksum mismatch")
			assert.Equal(t, hex.EncodeToString(hash.Sum(nil)), status.Diffs()["checksum"].Original())
			assert.Equal(t, task.Hash, status.Diffs()["checksum"].Current())
			assert.True(t, status.HasChanges())
		})

		t.Run("force=false", func(t *testing.T) {
			src, err := ioutil.TempFile("", "fetch_test.txt")
			require.NoError(t, err)
			defer os.Remove(src.Name())

			dest, err := ioutil.TempFile("", "fetch_test.txt")
			require.NoError(t, err)
			defer os.Remove(dest.Name())

			hash, err := getHash(src.Name(), string(fetch.HashMD5))
			require.NoError(t, err)

			task := fetch.Fetch{
				Source:      src.Name(),
				Destination: dest.Name(),
				HashType:    string(fetch.HashMD5),
				Hash:        "notarealhashbutstringnonetheless",
			}

			status, err := task.DiffFile(resource.NewStatus(), hash)

			assert.EqualError(t, err, "will not attempt fetch: checksum mismatch")
			assert.Contains(t, status.Messages(), "checksum mismatch, use the \"force\" option to replace")
			assert.True(t, status.HasChanges())
		})
	})

	t.Run("file exists", func(t *testing.T) {
		t.Run("force=true", func(t *testing.T) {
			src, err := ioutil.TempFile("", "fetch_test.txt")
			require.NoError(t, err)
			defer os.Remove(src.Name())

			dest, err := ioutil.TempFile("", "fetch_test.txt")
			require.NoError(t, err)
			defer os.Remove(dest.Name())

			task := fetch.Fetch{
				Source:      src.Name(),
				Destination: dest.Name(),
				Force:       true,
			}

			status, err := task.DiffFile(resource.NewStatus(), nil)

			assert.NoError(t, err)
			assert.Contains(t, status.Messages(), "file exists, will fetch due to \"force\"")
			assert.True(t, status.HasChanges())
		})

		t.Run("force=false", func(t *testing.T) {
			src, err := ioutil.TempFile("", "fetch_test.txt")
			require.NoError(t, err)
			defer os.Remove(src.Name())

			dest, err := ioutil.TempFile("", "fetch_test.txt")
			require.NoError(t, err)
			defer os.Remove(dest.Name())

			hash, err := getHash(src.Name(), string(fetch.HashMD5))
			require.NoError(t, err)

			task := fetch.Fetch{
				Source:      src.Name(),
				Destination: dest.Name(),
				HashType:    string(fetch.HashMD5),
				Hash:        hex.EncodeToString(hash.Sum(nil)),
			}

			status, err := task.DiffFile(resource.NewStatus(), hash)

			assert.NoError(t, err)
			assert.Contains(t, status.Messages(), "file exists")
			assert.False(t, status.HasChanges())
		})
	})
}

func getHash(path, hashType string) (hash.Hash, error) {
	var h hash.Hash
	switch hashType {
	case string(fetch.HashMD5):
		h = md5.New()
	case string(fetch.HashSHA1):
		h = sha1.New()
	case string(fetch.HashSHA256):
		h = sha256.New()
	case string(fetch.HashSHA512):
		h = sha512.New()
	default:
		err := fmt.Errorf("unsupported hashType %q", hashType)
		return nil, err
	}

	f, err := os.Open(path)
	if err != nil {
		err = fmt.Errorf("Failed to open file for checksum: %s", err)
		return nil, err
	}
	defer f.Close()
	_, err = io.Copy(h, f)
	return h, err
}

// MockDiff is a mock implementation for file diffs
type MockDiff struct {
	mock.Mock
}

// DiffFile mocks the fetch DiffFile
func (m *MockDiff) DiffFile(r resource.Status, hsh hash.Hash) (*resource.Status, error) {
	args := m.Called(r, hsh)
	return args.Get(0).(*resource.Status), args.Error(1)
}
