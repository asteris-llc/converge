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
	"errors"
	"fmt"
	"strings"
)

// Ownership represents a file ownership
type Ownership struct {
	UID *int
	GID *int
}

// OwnershipDiff diffs user and group IDs
type OwnershipDiff struct {
	p    OSProxy
	Path string
	UIDs *[2]int
	GIDs *[2]int
}

// Apply applies the changes in ownership to a file based on a diff
func (d *OwnershipDiff) Apply() error {
	if !d.Changes() {
		return nil
	}
	var newUID *int
	var newGID *int
	oldOwner, err := fileOwnership(d.p, d.Path)
	if err != nil {
		return err
	}
	if d.UIDs != nil {
		newUID = &(*d.UIDs)[1]
	} else {
		newUID = oldOwner.UID
	}
	if d.GIDs != nil {
		newGID = &(*d.GIDs)[1]
	} else {
		newGID = oldOwner.GID
	}
	return d.p.Chown(d.Path, *newUID, *newGID)
}

// showDiffAt shows the UID/GID at a given index
func (d *OwnershipDiff) showDiffAt(idx uint) string {
	var diffStrs []string
	if idx > 1 {
		return "Error executing diff generation on " + d.Path
	}
	if (d.UIDs != nil) && d.UIDs[0] != d.UIDs[1] {
		uid := d.UIDs[idx]
		userName, err := usernameFromUID(d.p, show(uid))
		if err != nil {
			return show(err)
		}
		diffStrs = append(diffStrs, fmt.Sprintf("user: %s (%d)", userName, uid))
	}
	if (d.GIDs != nil) && d.GIDs[0] != d.GIDs[1] {
		gid := d.GIDs[idx]
		groupName, err := groupnameFromGID(d.p, show(gid))
		if err != nil {
			return show(err)
		}
		diffStrs = append(diffStrs, fmt.Sprintf("group: %s (%d)", groupName, gid))
	}
	return strings.Join(diffStrs, "; ")
}

// Original returns the original UID and GID
func (d *OwnershipDiff) Original() string {
	return d.showDiffAt(0)
}

// Current returns the desired UID and GID
func (d *OwnershipDiff) Current() string {
	return d.showDiffAt(1)
}

// Changes returns true if Original() != Current()
func (d *OwnershipDiff) Changes() bool {
	var changes bool
	if d.UIDs != nil {
		changes = changes || (d.UIDs[0] != d.UIDs[1])
	}
	if d.GIDs != nil {
		changes = changes || (d.GIDs[0] != d.GIDs[1])
	}
	return changes
}

// SetProxy sets the internal OS Proxy
func (d *OwnershipDiff) SetProxy(p OSProxy) *OwnershipDiff {
	d.p = p
	return d
}

func fileOwnership(p OSProxy, path string) (*Ownership, error) {
	uid, err := p.GetUID(path)
	if err != nil {
		return nil, err
	}
	gid, err := p.GetGID(path)
	if err != nil {
		return nil, err
	}
	return &Ownership{UID: &uid, GID: &gid}, nil
}

// NewOwnershipDiff creates a new diff
func NewOwnershipDiff(p OSProxy, path string, ownership *Ownership) (*OwnershipDiff, error) {
	var (
		newUID int
		newGID int
		curUID int
		curGID int
	)

	currentOwner, err := fileOwnership(p, path)
	if err != nil {
		return nil, err
	}

	if currentOwner.UID == nil {
		return nil, errors.New("unable to fetch UID for " + path)
	}

	if currentOwner.GID == nil {
		return nil, errors.New("unable to fetch GID for " + path)
	}

	curUID = *(currentOwner.UID)
	curGID = *(currentOwner.GID)

	if ownership.UID == nil {
		newUID = curUID
	} else {
		newUID = *(ownership.UID)
	}

	if ownership.GID == nil {
		newGID = curGID
	} else {
		newGID = *(ownership.GID)
	}

	diff := &OwnershipDiff{Path: path, p: p}

	if curUID != newUID {
		diff.UIDs = &[2]int{curUID, newUID}
	}
	if curGID != newGID {
		diff.GIDs = &[2]int{curGID, newGID}
	}
	return diff, nil
}
