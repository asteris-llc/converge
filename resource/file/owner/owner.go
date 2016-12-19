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

package owner

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"

	"github.com/asteris-llc/converge/resource"
	"golang.org/x/net/context"
)

// OSProxy is an intermediary used for interfacing with the underlying OS or
// test mocks.
type OSProxy interface {
	Walk(string, filepath.WalkFunc) error
	Chown(string, int, int) error
	GetUID(string) (int, error)
	GetGID(string) (int, error)
	LookupGroupID(string) (*user.Group, error)
	LookupGroup(string) (*user.Group, error)
	LookupID(string) (*user.User, error)
	Lookup(string) (*user.User, error)
}

// Owner represents the ownership mode of a file or directory
type Owner struct {
	// the path to the file that should change; or if `recursive` is set, the path
	// to the root of the filesystem to recursively change.
	Destination string `export:"destination"`

	// the username of the user that should be given ownership of the file
	Username string `export:"username"`

	// the uid of the user that should be given ownership of the file
	UID string `export:"uid"`

	// the group name of the group that should be given ownership of the file
	Group string `export:"group"`

	// the gid of the group that should be given ownership of the file
	GID string `export:"gid"`

	// if true, and `destination` is a directory, apply changes recursively
	Recursive bool `export:"recursive"`

	HideDetails bool
	executor    OSProxy
	differences []*OwnershipDiff
}

type fileWalker struct {
	Status   *resource.Status
	NewOwner *Ownership
	Executor OSProxy
}

// SetOSProxy sets the internal OS Proxy
func (o *Owner) SetOSProxy(p OSProxy) *Owner {
	o.executor = p
	return o
}

// Check checks the ownership status of the file
func (o *Owner) Check(context.Context, resource.Renderer) (resource.TaskStatus, error) {
	status := &resource.Status{Differences: make(map[string]resource.Diff)}
	newOwner, err := o.getNewOwner()
	if err != nil {
		return nil, err
	}
	w := &fileWalker{Status: status, NewOwner: newOwner, Executor: o.executor}
	if o.Recursive {
		status.AddMessage("Recursively updating permissions in " + o.Destination)
		err = o.executor.Walk(o.Destination, w.CheckFile)
	} else {
		err = w.CheckFile(o.Destination, nil, nil)
	}
	if err != nil {
		return nil, err
	}
	o.copyDiffs(status)

	if o.HideDetails && o.Recursive && status.HasChanges() {
		status.AddMessage("reporting abridged changes; add \"verbose = true\" to see all changes")
		status.Differences = make(map[string]resource.Diff)
		status.AddDifference(o.Destination, "*", fmt.Sprintf("user: %s (%s); group: %s (%s)", o.Username, o.UID, o.Group, o.GID), "")
	}

	return status, nil
}

// Apply applies the ownership to the file
func (o *Owner) Apply(context.Context) (resource.TaskStatus, error) {
	showDetails := !(o.Recursive && o.HideDetails)
	status := resource.NewStatus()
	if o.Recursive {
		status.AddMessage("Recursively updating permissions in " + o.Destination)
	}
	for _, diff := range o.differences {
		if showDetails {
			status.Differences[diff.Path] = diff
		}
		if err := diff.Apply(); err != nil {
			return nil, err
		}
	}
	if !showDetails {
		status.AddMessage("reporting abridged changes; add \"verbose = true\" to see all changes")
		status.Differences = make(map[string]resource.Diff)
		status.AddDifference(o.Destination, "*", fmt.Sprintf("user: %s (%s); group: %s (%s)", o.Username, o.UID, o.Group, o.GID), "")
	}
	return status, nil
}

func (o *Owner) copyDiffs(status *resource.Status) {
	for _, diff := range status.Differences {
		o.differences = append(o.differences, diff.(*OwnershipDiff))
	}
}

func (o *Owner) getNewOwner() (*Ownership, error) {
	owner := &Ownership{}

	if o.UID != "" {
		uid, err := strconv.Atoi(o.UID)
		if err != nil {
			return nil, err
		}
		owner.UID = &uid
	}

	if o.GID != "" {
		gid, err := strconv.Atoi(o.GID)
		if err != nil {
			return nil, err
		}
		owner.GID = &gid
	}
	return owner, nil
}

func (w *fileWalker) CheckFile(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	diff, err := NewOwnershipDiff(w.Executor, path, w.NewOwner)
	if err != nil {
		return err
	}
	if diff.Changes() {
		w.Status.Differences[path] = diff
	}
	return nil
}
