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
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
	"strings"

	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/file/fetch"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

// Hash type for Unarchive
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

// Unarchive manages unarchive
type Unarchive struct {

	// the source
	Source string `export:"source"`

	// the destination
	Destination string `export:"destination"`

	// hash function used to generate the checksum hash of the source; value is
	// available for lookup if set in the hcl
	HashType string `export:"hash_type"`

	// the checksum hash of the source; value is available for lookup if set in
	// the hcl
	Hash string `export:"hash"`

	// whether a file from the unarchived source will replace a file in the
	// destination if it already exists
	Force bool `export:"force"`

	// the destination directory
	destDir *os.File

	// the files within the destination directory
	destContents []string

	// the intermediate directory containing fetched and unarchived file(s)
	fetchDir *os.File

	// the files within the intermediate fetch directory
	fetchContents []string

	// the location of the fetched file
	fetchLoc string
}

// Check if changes are needed for unarchive
func (u *Unarchive) Check(ctx context.Context, r resource.Renderer) (resource.TaskStatus, error) {
	status := resource.NewStatus()

	err := u.Diff(status)
	if err != nil {
		return status, err
	}

	err = u.setFetchLoc()
	if err != nil {
		status.RaiseLevel(resource.StatusCantChange)
		return status, errors.Wrap(err, "error setting fetch location")
	}

	fetch := fetch.Fetch{
		Source:      u.Source,
		Destination: u.fetchLoc,
		HashType:    u.HashType,
		Hash:        u.Hash,
		Unarchive:   true,
	}

	fetchStatus, err := fetch.Check(ctx, r)
	if err != nil {
		return fetchStatus, errors.Wrap(err, "cannot attempt unarchive: fetch error")
	}

	status.AddMessage(fmt.Sprintf("fetch and unarchive %q", u.Source))

	return status, nil
}

// Apply changes for unarchive
func (u *Unarchive) Apply(ctx context.Context) (resource.TaskStatus, error) {
	status := resource.NewStatus()

	err := u.Diff(status)
	if err != nil {
		return status, err
	}

	err = u.setFetchLoc()
	if err != nil {
		status.RaiseLevel(resource.StatusCantChange)
		return status, errors.Wrap(err, "error setting fetch location")
	}

	fetch := fetch.Fetch{
		Source:      u.Source,
		Destination: u.fetchLoc,
		HashType:    u.HashType,
		Hash:        u.Hash,
		Unarchive:   true,
	}

	fetchStatus, err := fetch.Apply(ctx)
	if err != nil {
		return fetchStatus, err
	}

	evaluateDuplicates, err := u.setDirsAndContents()
	if err != nil {
		return status, err
	}

	if u.Force == false && evaluateDuplicates {
		err = u.evaluateDuplicates()
		if err != nil {
			status.RaiseLevel(resource.StatusFatal)
			if strings.Contains(err.Error(), "checksum mismatch") {
				status.AddMessage("use the \"force\" option to replace all files with checksum mismatch")
			}
			return status, err
		}
	}

	status.AddMessage(fmt.Sprintf("completed fetch and unarchive %q", u.Source))

	return status, nil
}

// Diff evaluates the differences for unarchive
func (u *Unarchive) Diff(status *resource.Status) error {
	_, err := os.Stat(u.Source)
	if os.IsNotExist(err) {
		status.RaiseLevel(resource.StatusCantChange)
		return errors.Wrap(err, "cannot unarchive")
	}

	stat, err := os.Stat(u.Destination)
	if err == nil {
		if !stat.IsDir() {
			status.RaiseLevel(resource.StatusCantChange)
			return fmt.Errorf("invalid destination %q, must be directory", u.Destination)
		}
	} else if os.IsNotExist(err) {
		status.RaiseLevel(resource.StatusCantChange)
		return fmt.Errorf("destination %q does not exist", u.Destination)
	}

	status.AddDifference("unarchive", u.Source, u.Destination, "")
	status.RaiseLevelForDiffs()

	return nil
}

// setDirsAndContents sets the Unarchive fields of unarchive destination and its
// contents, and the temporary fetch/unarchive destination and its contents. A
// bool indicating whether duplicates need to be evaluated between the unarchive
// destination and the temporary fetch/unarchive destination.
func (u *Unarchive) setDirsAndContents() (bool, error) {
	var err error

	// create the unarchive destination directory if it does not exist
	u.destDir, err = os.Open(u.Destination)
	if os.IsNotExist(err) {
		err = os.Mkdir(u.Destination, 1755)
		if err != nil {
			return false, err
		}
		return false, nil
	} else if err != nil {
		return false, err
	}

	// read the file names in the directory indicated by u.Destination
	u.destContents, err = u.destDir.Readdirnames(0)
	if err != nil {
		return false, errors.Wrapf(err, "could not read files from %q", u.Destination)
	}

	// if there are no files, we do not need to compare checksums with files in
	// the temporary fetch/unarchive location
	if len(u.destContents) == 0 {
		return false, nil
	}

	// read the contents of the temporary fetch/unarchive location
	fetchDir := u.fetchLoc
	u.fetchDir, err = os.Open(fetchDir)
	if err != nil {
		return false, err
	}
	fetchDirContents, err := u.fetchDir.Readdir(0)
	if err != nil {
		return false, errors.Wrapf(err, "could not read dir %q", fetchDir)
	}

	// if one directory is within the temporary fetch/unarchive location, we need
	// to use this as the directory to read file names
	if len(fetchDirContents) == 1 && fetchDirContents[0].IsDir() {
		fetchDir = u.fetchLoc + "/" + fetchDirContents[0].Name()
	}

	// read the file names in the temporary fetch/unarchive location
	u.fetchDir, err = os.Open(fetchDir)
	u.fetchContents, err = u.fetchDir.Readdirnames(0)
	if err != nil {
		return false, errors.Wrapf(err, "could not read files from %q", fetchDir)
	}

	return true, nil
}

// evaluateDuplicates evaluates whether identical files exist in u.Destination
// and the temporary fetch/unarchive location
func (u *Unarchive) evaluateDuplicates() error {
	// determine which directory has fewer items in order to minimize operations
	filesA := u.destContents
	filesB := u.fetchContents
	dirA := u.destDir.Name()
	dirB := u.fetchDir.Name()
	if len(u.fetchContents) < len(u.destContents) {
		filesA = u.fetchContents
		filesB = u.destContents
		dirA = u.fetchDir.Name()
		dirB = u.destDir.Name()
	}

	// for each item in filesA, determine if it also exists in filesB
	// compare the checksums for the files - if mismatch, return an error
	for _, fileA := range filesA {
		for _, fileB := range filesB {

			if fileA == fileB {
				checkA, err := u.getChecksum(dirA + "/" + fileA)
				if err != nil {
					return err
				}

				checkB, err := u.getChecksum(dirB + "/" + fileB)
				if err != nil {
					return err
				}

				if checkA != checkB {
					return fmt.Errorf("will not replace, file %q exists at %q: checksum mismatch", fileA, u.Destination)
				}

				break
			}
		}
	}

	return nil
}

// setFetchLoc sets the location for the fetch destination
func (u *Unarchive) setFetchLoc() error {
	if u.fetchLoc != "" {
		return nil
	}

	checksum, err := u.getChecksum(u.Source)
	if err != nil {
		return errors.Wrap(err, "failed to get checksum of source")
	}
	u.fetchLoc = "/var/run/converge/cache/" + checksum

	return nil
}

// getChecksum obtains the checksum of the provided file
func (u *Unarchive) getChecksum(f string) (string, error) {
	hsh := u.getHash()

	file, err := os.Open(f)
	if err != nil {
		return "", errors.Wrap(err, "failed to open file for checksum")
	}
	defer file.Close()

	if _, err := io.Copy(hsh, file); err != nil {
		return "", errors.Wrap(err, "failed to hash")
	}

	return hex.EncodeToString(hsh.Sum(nil)), nil
}

// getHash returns a new hash based on the u.HashType
// default hash is sha256
func (u *Unarchive) getHash() hash.Hash {
	switch u.HashType {
	case string(HashMD5):
		return md5.New()
	case string(HashSHA1):
		return sha1.New()
	case string(HashSHA256):
		return sha256.New()
	case string(HashSHA512):
		return sha512.New()
	default:
		return sha256.New()
	}
}
