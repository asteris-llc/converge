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
	LookupGroupId(string) (*user.Group, error)
	LookupGroup(string) (*user.Group, error)
	LookupId(string) (*user.User, error)
	Lookup(string) (*user.User, error)
}

// Owner represents the ownership mode of a file or directory
type Owner struct {
	Destination string `export:"destination"`
	Username    string `export:"username"`
	UID         string `export:"uid"`
	Group       string `export:"group"`
	GID         string `export:"gid"`
	Recursive   bool   `export:"recursive"`
	executor    OSProxy
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
	status := &resource.Status{}
	newOwner, err := o.getNewOwner()
	if err != nil {
		return nil, err
	}
	w := &fileWalker{Status: status, NewOwner: newOwner, Executor: o.executor}

	if o.Recursive {
	}

	if err := w.CheckFile(o.Destination, nil, nil); err != nil {
		return nil, err
	}
	return status, nil
}

// Apply applies the ownership to the file
func (o *Owner) Apply(context.Context) (resource.TaskStatus, error) {
	return &resource.Status{}, nil
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
