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
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

// Hash type for Fetch
type Hash string

const (
	// HashMD5 indicates hash type md5
	HashMD5 Hash = "md5"

	// HashSHA1 indicates hash type sha1
	HashSHA1 Hash = "sha1"

	// HashSHA256 indicates hash type sha256
	HashSHA256 Hash = "sha256"

	// HashSHA512 indicates hash type sha512
	HashSHA512 Hash = "sha512"
)

// Fetch gets a file and makes it available on disk
type Fetch struct {
	// location of the file to fetch
	Source string `export:"source"`

	// destination for the fetched file
	Destination string `export:"destination"`

	// hash function used to generate the checksum hash; value is available for
	// lookup if set in the hcl
	HashType string `export:"hash_type"`

	// the checksum hash; value is available for lookup if set in the hcl
	Hash string `export:"hash"`

	// whether the file will be fetched if it already exists
	Force bool `export:"force"`

	hasApplied bool
}

// Check if the the file is on disk, and the hashes are availabe
func (f *Fetch) Check(context.Context, resource.Renderer) (resource.TaskStatus, error) {
	var (
		hsh    hash.Hash
		err    error
		status = resource.NewStatus()
	)

	if f.hasApplied {
		return status, nil
	}

	if f.Hash != "" {
		hsh, err = f.getHash()
		if err != nil {
			status.RaiseLevel(resource.StatusCantChange)
			return status, errors.Wrap(err, "will not attempt file fetch")
		}
	}

	status, err = f.DiffFile(status, hsh)
	if err != nil {
		return status, err
	}

	return status, nil
}

// Apply fetches the file
func (f *Fetch) Apply(context.Context) (resource.TaskStatus, error) {
	var (
		hsh      hash.Hash
		err      error
		status   = resource.NewStatus()
		checksum = ""
	)

	if f.Hash != "" {
		hsh, err = f.getHash()
		if err != nil {
			status.RaiseLevel(resource.StatusCantChange)
			return status, errors.Wrap(err, "will not attempt file fetch")
		}
	}

	stat, err := f.DiffFile(status, hsh)
	if err != nil {
		return stat, err
	} else if !resource.AnyChanges(stat.Differences) {
		return status, nil
	}

	u, err := url.Parse(f.Source)
	if err != nil {
		status.RaiseLevel(resource.StatusFatal)
		return status, errors.Wrap(err, "could not parse source")
	}

	values := u.Query()
	values.Set("archive", "false")
	if f.Hash != "" {
		checksum = fmt.Sprintf("checksum=%s:%s", f.HashType, f.Hash)
	}
	source := u.String() + "?" + values.Encode() + checksum

	pwd, err := os.Getwd()
	if err != nil {
		status.RaiseLevel(resource.StatusFatal)
		return status, errors.Wrap(err, "failed to get working directory")
	}

	client := &getter.Client{
		Src:  source,
		Dst:  f.Destination,
		Pwd:  pwd,
		Mode: getter.ClientModeFile,
	}
	if err := client.Get(); err != nil {
		status.RaiseLevel(resource.StatusFatal)
		return status, errors.Wrap(err, "failed to fetch")
	}
	status.AddMessage("fetched successfully")
	f.hasApplied = true

	return status, nil
}

// DiffFile evaluates the differences of the file to be fetched and the current
// state of the system
func (f *Fetch) DiffFile(status *resource.Status, hsh hash.Hash) (*resource.Status, error) {
	// verify the destination is not a directory
	stat, err := os.Stat(f.Destination)
	if err == nil {
		if stat.IsDir() {
			status.RaiseLevel(resource.StatusCantChange)
			return status, fmt.Errorf("invalid destination %q, cannot be directory", f.Destination)
		}
	} else if os.IsNotExist(err) {
		status.RaiseLevel(resource.StatusWillChange)
		status.AddDifference("destination", "<absent>", f.Destination, "")
		return status, nil
	}

	// file exists, evaluate what needs to change
	if hsh != nil {
		actual, err := f.getChecksum(hsh)
		if err != nil {
			status.RaiseLevel(resource.StatusFatal)
			return status, err
		}

		// evaluate the checksums
		if actual == f.Hash {
			status.AddMessage("file exists")
		} else if f.Force {
			status.AddDifference("checksum", actual, f.Hash, "")
			status.AddMessage("checksum mismatch")
			status.RaiseLevel(resource.StatusWillChange)
		} else {
			status.AddMessage("checksum mismatch, use the \"force\" option to replace")
			status.RaiseLevel(resource.StatusCantChange)
			return status, errors.New("will not attempt fetch: checksum mismatch")
		}
	} else {
		if f.Force {
			status.AddDifference("destination", "<force fetch>", f.Destination, "")
			status.AddMessage("file exists, will fetch due to \"force\"")
			status.RaiseLevel(resource.StatusWillChange)
		} else {
			status.AddMessage("file exists")
		}
	}

	return status, nil
}

// getHash returns a new hash based on the f.HashType
func (f *Fetch) getHash() (hash.Hash, error) {
	switch f.HashType {
	case string(HashMD5):
		return md5.New(), nil
	case string(HashSHA1):
		return sha1.New(), nil
	case string(HashSHA256):
		return sha256.New(), nil
	case string(HashSHA512):
		return sha512.New(), nil
	default:
		return nil, fmt.Errorf("unsupported hashType %q", f.HashType)
	}
}

// checksum obtains the checksum of the destination
func (f *Fetch) getChecksum(hsh hash.Hash) (string, error) {
	file, err := os.Open(f.Destination)
	if err != nil {
		return "", errors.Wrap(err, "failed to open file for checksum")
	}
	defer file.Close()

	if _, err := io.Copy(hsh, file); err != nil {
		return "", errors.Wrap(err, "failed to hash")
	}

	return hex.EncodeToString(hsh.Sum(nil)), nil
}
