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
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"syscall"

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

	// fetch is used to fetch the file to be unarchived
	fetch fetch.Fetch

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

	// the size in bytes of the fetched/unarchived data
	dataSize int64

	hasApplied bool
}

// Check if changes are needed for unarchive
func (u *Unarchive) Check(ctx context.Context, r resource.Renderer) (resource.TaskStatus, error) {
	status := resource.NewStatus()
	defer os.RemoveAll(u.fetchLoc)

	if u.hasApplied {
		return status, nil
	}

	err := u.Diff(status)
	if err != nil {
		return status, err
	}

	fetchStatus, err := u.fetch.Check(ctx, r)
	if err != nil {
		return fetchStatus, errors.Wrap(err, "cannot attempt unarchive: fetch error")
	}

	status.AddMessage(fmt.Sprintf("fetch and unarchive %q", u.Source))

	return status, nil
}

// Apply changes for unarchive
func (u *Unarchive) Apply(ctx context.Context) (resource.TaskStatus, error) {
	status := resource.NewStatus()
	defer os.RemoveAll(u.fetchLoc)

	err := u.Diff(status)
	if err != nil {
		return status, err
	}

	err = u.setFetchLoc()
	if err != nil {
		return nil, errors.Wrap(err, "error setting fetch location")
	}

	fetchStatus, err := u.fetch.Apply(ctx)
	if err != nil {
		return fetchStatus, err
	}

	evaluateDuplicates, err := u.setDirsAndContents()
	if err != nil {
		return status, err
	}

	mem, err := u.isMemAvailable()
	if !mem || err != nil {
		return status, err
	}

	if !u.Force && evaluateDuplicates {
		err = u.evaluateDuplicates()
		if err != nil {
			status.RaiseLevel(resource.StatusFatal)
			if strings.Contains(err.Error(), "checksum mismatch") {
				status.AddMessage("use the \"force\" option to replace all files with checksum mismatch")
				u.hasApplied = true
			}
			return status, err
		}
	}

	err = u.copyToFinalDest()
	if err != nil {
		status.RaiseLevel(resource.StatusFatal)
		return status, errors.Wrapf(err, "error placing files in %q", u.Destination)
	}

	status.AddMessage(fmt.Sprintf("completed fetch and unarchive %q", u.Source))
	u.hasApplied = true

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

	// set the unarchive destination directory
	u.destDir, err = os.Open(u.Destination)
	if err != nil {
		return false, err
	}

	// walk the destination directory to set the destination contents
	err = filepath.Walk(u.destDir.Name(), func(path string, f os.FileInfo, err error) error {
		u.destContents = append(u.destContents, path)
		return nil
	})
	if err != nil {
		return false, err
	}

	// read the contents of the temporary fetch/unarchive location
	fetchDir := u.fetchLoc
	u.fetchDir, err = os.Open(fetchDir)
	if err != nil {
		return false, err
	}

	// walk the fetch directory to set the fetch contents and determine size
	err = filepath.Walk(u.fetchDir.Name(), func(path string, f os.FileInfo, err error) error {
		u.fetchContents = append(u.fetchContents, path)
		if !f.IsDir() {
			u.dataSize += f.Size()
		}
		return nil
	})
	if err != nil {
		return false, err
	}

	// if there are no files, we do not need to compare checksums with files in
	// the temporary fetch/unarchive location. We check whether the length is 1
	// because the directory is in the contents.
	if len(u.destContents) == 1 {
		return false, nil
	}
	return true, nil
}

// isMemAvailable determines whether adequate memory exists in both the
// temporary fetch/unarchive location and the destination based on u.dataSize
func (u *Unarchive) isMemAvailable() (bool, error) {
	var (
		destStat syscall.Statfs_t
		tmpStat  syscall.Statfs_t
	)

	// determine available space in temporary fetch location
	err := syscall.Statfs(os.TempDir(), &tmpStat)
	if err != nil {
		return false, err
	}
	tmpFetchAvailable := tmpStat.Bavail * uint64(tmpStat.Bsize)

	// determine available space in destination
	err = syscall.Statfs(u.destDir.Name(), &destStat)
	if err != nil {
		return false, err
	}
	destAvailable := destStat.Bavail * uint64(destStat.Bsize)

	if strings.HasPrefix(u.destDir.Name(), os.TempDir()) {
		if destAvailable < 2*uint64(u.dataSize) {
			return false, fmt.Errorf("not enough memory in %q for fetch and unarchive", os.TempDir())
		}
	}
	if tmpFetchAvailable < uint64(u.dataSize) {
		return false, fmt.Errorf("not enough memory in %q for fetch", os.TempDir())
	}
	if destAvailable < uint64(u.dataSize) {
		return false, fmt.Errorf("not enough memory in %q for unarchive", u.destDir.Name())
	}

	return true, nil
}

// evaluateDuplicates evaluates whether identical files exist in u.Destination
// and the temporary fetch/unarchive location
func (u *Unarchive) evaluateDuplicates() error {
	// determine which directory has fewer items in order to minimize operations
	dirA := u.destDir.Name()
	dirB := u.fetchDir.Name()
	filesA := u.destContents
	filesB := u.fetchContents
	if len(u.fetchContents) < len(u.destContents) {
		dirA = u.fetchDir.Name()
		dirB = u.destDir.Name()
		filesA = u.fetchContents
		filesB = u.destContents
	}

	// for each item in filesA, determine if it also exists in filesB
	// compare the checksums for the files - if mismatch, return an error
	for _, fA := range filesA {
		for _, fB := range filesB {
			fileA := strings.TrimPrefix(fA, dirA)
			fileB := strings.TrimPrefix(fB, dirB)

			faStat, err := os.Stat(fA)
			if err != nil {
				return err
			}
			fbStat, err := os.Stat(fB)
			if err != nil {
				return err
			}

			if !faStat.IsDir() && !fbStat.IsDir() && fileA == fileB {

				if faStat.Size() != fbStat.Size() {
					return fmt.Errorf("will not replace, %q exists at %q: checksum mismatch", fileA, u.Destination)
				}

				checkA, err := u.getChecksum(fA)
				if err != nil {
					return err
				}

				checkB, err := u.getChecksum(fB)
				if err != nil {
					return err
				}

				if checkA != checkB {
					return fmt.Errorf("will not replace, %q exists at %q: checksum mismatch", fileA, u.Destination)
				}

				break
			}
		}
	}

	return nil
}

// copyToFinalDest copies the fetched and unarchived files from their temporary
// directory to the final destination
func (u *Unarchive) copyToFinalDest() error {
	// for each item in the fetchDir, mkdir or copy to the final destination
	for _, file := range u.fetchContents {
		src, err := os.Open(file)
		if err != nil {
			return err
		}
		defer src.Close()

		fileName := strings.TrimPrefix(file, u.fetchDir.Name())

		fStat, err := os.Stat(file)
		if err != nil {
			return err
		}

		if fileName != "" {
			if fStat.IsDir() {
				err = os.Mkdir(u.destDir.Name()+fileName, fStat.Mode().Perm())
				if err != nil {
					if !os.IsNotExist(err) {
						continue
					}
					return err
				}
			} else {
				// get src []byte
				srcData, err := ioutil.ReadAll(src)
				if err != nil {
					return err
				}

				// get src FileInfo
				srcInfo, err := src.Stat()
				if err != nil {
					return err
				}

				err = ioutil.WriteFile(u.destDir.Name()+fileName, srcData, srcInfo.Mode().Perm())
				if err != nil {
					return err
				}
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

	dir, err := ioutil.TempDir("", "tmpDirFetch")
	if err != nil {
		return errors.Wrap(err, "failed to set temporary fetch location")
	}

	u.fetchLoc = dir
	u.fetch.Destination = u.fetchLoc

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
