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
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"

	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/file/fetch"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

// Unarchive manages unarchive
type Unarchive struct {

	// the source
	Source string `export:"source"`

	// the destination
	Destination string `export:"destination"`

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
	}

	fetchStatus, err := fetch.Apply(ctx)
	if err != nil {
		return fetchStatus, err
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

// setFetchLoc sets the location for the fetch destination
func (u *Unarchive) setFetchLoc() error {
	if u.fetchLoc != "" {
		return nil
	}

	base := filepath.Base(u.Source)
	checksum, err := u.getChecksum(nil)
	if err != nil {
		return errors.Wrap(err, "failed to get checksum of source")
	}
	u.fetchLoc = "/var/run/converge/cache/" + checksum + "/" + base

	return nil
}

// getChecksum obtains the checksum of the destination
// Defaults to sha256 if no hash type is specified
func (u *Unarchive) getChecksum(hsh hash.Hash) (string, error) {
	if hsh == nil {
		hsh = sha256.New()
	}

	file, err := os.Open(u.Source)
	if err != nil {
		return "", errors.Wrap(err, "failed to open file for checksum")
	}
	defer file.Close()

	if _, err := io.Copy(hsh, file); err != nil {
		return "", errors.Wrap(err, "failed to hash")
	}

	return hex.EncodeToString(hsh.Sum(nil)), nil
}
