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

package fetch

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"net/url"
	"os"

	"github.com/asteris-llc/converge/resource"
	"github.com/hashicorp/go-getter"
)

// Fetch gets a file and makes it available on disk
type Fetch struct {
	resource.TaskStatus
	Source      string
	Destination string
	HashType    string
	Hash        string
}

// Check if the the file is on disk, and the hashes are availabe
func (t *Fetch) Check(resource.Renderer) (resource.TaskStatus, error) {
	source := t.Destination
	hashType := t.HashType
	shouldBe := t.Hash

	// Determine the hashtypes, supports "md5", "sha1", "sha256", "sha512"
	var h hash.Hash
	switch hashType {
	case "md5":
		h = md5.New()
	case "sha1":
		h = sha1.New()
	case "sha256":
		h = sha256.New()
	case "sha512":
		h = sha512.New()
	default:
		err := fmt.Errorf("unsupported hashType: %s", hashType)
		t.TaskStatus = &resource.Status{
			Level:  resource.StatusFatal,
			Output: []string{err.Error()},
		}
		return t, err

	}

	// destination doesn't exist -> apply
	// destination is a directory -> throw error
	if stat, err := os.Stat(source); err == nil {
		if stat.IsDir() {
			err := fmt.Errorf("%q is a directory. cannot checksum a directory", source)
			t.TaskStatus = &resource.Status{
				Level:  resource.StatusCantChange,
				Output: []string{err.Error()},
			}
			return t, err
		}
	} else if os.IsNotExist(err) {
		t.TaskStatus = &resource.Status{
			Level:  resource.StatusWillChange,
			Output: []string{fmt.Sprintf("%q does not exist. will be fetched", source)},
		}
		return t, nil
	}

	f, err := os.Open(source)
	if err != nil {
		err = fmt.Errorf("Failed to open file for checksum: %s", err)
		t.TaskStatus = &resource.Status{
			Level:  resource.StatusFatal,
			Output: []string{err.Error()},
		}
		return t, err
	}
	defer f.Close()

	if _, err := io.Copy(h, f); err != nil {
		err = fmt.Errorf("Failed to hash: %s", err)
		t.TaskStatus = &resource.Status{
			Level:  resource.StatusFatal,
			Output: []string{err.Error()},
		}
		return t, err
	}

	// no hash given -> do nothing
	actual := h.Sum(nil)
	if shouldBe == "" {
		t.TaskStatus = &resource.Status{
			Level:  resource.StatusWontChange,
			Output: []string{"No hash given, won't change file", fmt.Sprintf("actual %s hash is %q", hashType, actual)},
		}
		return t, nil
	}

	diffs := make(map[string]resource.Diff)
	diffs[source] = &ChecksumDiff{
		Actual:   checksum{hashType: hashType, hash: actual},
		Expected: checksum{hashType: hashType, hash: []byte(shouldBe)},
	}

	// hash is checkted here
	if !bytes.Equal(actual, []byte(shouldBe)) {
		t.TaskStatus = &resource.Status{
			Level:       resource.StatusWillChange,
			Differences: diffs,
			Output: []string{
				fmt.Sprintf(
					"Checksums did not match.\nExpected: %s\nGot: %s",
					hex.EncodeToString([]byte(shouldBe)),
					hex.EncodeToString(actual)),
			},
		}
		return t, nil
	}
	t.TaskStatus = &resource.Status{
		Level:       resource.StatusNoChange,
		Differences: diffs,
		Output:      []string{"Checksums matched"},
	}
	return t, nil
}

// Apply fetches the file
func (t *Fetch) Apply() (resource.TaskStatus, error) {
	u, err := url.Parse(t.Source)
	if err != nil {
		err = fmt.Errorf("could not parse source: %s", err)
		t.TaskStatus = &resource.Status{
			Level:  resource.StatusFatal,
			Output: []string{err.Error()},
		}
		return t, err
	}
	q := u.Query()
	q.Set("archive", "")
	if t.Hash != "" && t.HashType != "" {
		q.Set("checksum", fmt.Sprintf("%s:%s", t.HashType, t.Hash))
	}

	source := u.String()

	pwd, err := os.Getwd()
	if err != nil {
		err = fmt.Errorf("failed to fetch file: %s", err)
		t.TaskStatus = &resource.Status{
			Level:  resource.StatusFatal,
			Output: []string{err.Error()},
		}
		return t, err
	}
	client := &getter.Client{
		Src:  source,
		Dst:  t.Destination,
		Pwd:  pwd,
		Mode: getter.ClientModeFile,
	}
	if err := client.Get(); err != nil {
		err = fmt.Errorf("failed to fetch file: %s", err)
		t.TaskStatus = &resource.Status{
			Level:  resource.StatusFatal,
			Output: []string{err.Error()},
		}
		return t, err
	}
	t.TaskStatus = &resource.Status{
		Level:  resource.StatusNoChange,
		Output: []string{fmt.Sprintf("file %q fetched to location %q", t.Source, t.Destination)},
	}
	return t, err

}

type checksum struct {
	hashType string
	hash     []byte
}

// ChecksumDiff shows a diff of the file checksum
type ChecksumDiff struct {
	Actual   checksum
	Expected checksum
}

// Original shows the original checksum
func (diff *ChecksumDiff) Original() string {
	return fmt.Sprintf("hash_type: %q, hash_type: %q", diff.Actual.hashType, diff.Actual.hash)
}

// Current shows the current checksum
func (diff *ChecksumDiff) Current() string {
	return fmt.Sprintf("hash_type: %q, hash_type: %q", diff.Expected.hashType, diff.Expected.hash)
}

// Changes returns true if the expected checksum differs from the current
func (diff *ChecksumDiff) Changes() bool {
	return !(diff.Actual.hashType == diff.Expected.hashType && bytes.Equal(diff.Actual.hash, diff.Expected.hash))
}
