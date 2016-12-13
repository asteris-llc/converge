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
	"strings"
)

// Ownership represents a file ownership
type Ownership struct {
	UID int
	GID int
}

// OwnershipDiff diffs user and group IDs
type OwnershipDiff struct {
	path string
	p    OSProxy
	UIDs *[2]int
	GIDs *[2]int
}

// showDiffAt shows the UID/GID at a given index
func (d *OwnershipDiff) showDiffAt(idx uint) string {
	var diffStrs []string
	if idx > 1 {
		return "Error executing diff generation on " + d.path
	}
	if nil != d.UIDs {
		uid := d.UIDs[idx]
		userName, err := usernameFromUID(d.p, show(uid))
		if err != nil {
			return show(err)
		}
		diffStrs = append(diffStrs, fmt.Sprintf("user: %s (%d)", userName, uid))
	}
	if nil != d.GIDs {
		gid := d.GIDs[idx]
		groupName, err := groupnameFromGID(d.p, show(gid))
		if err != nil {
			return show(err)
		}
		diffStrs = append(diffStrs, fmt.Sprintf("group: %s (%d)", groupName, gid))
	}
	return strings.Join(diffStrs, ";")
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
	if nil != d.UIDs {
		changes = changes || (d.UIDs[0] != d.UIDs[1])
	}
	if nil != d.GIDs {
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
	return &Ownership{UID: uid, GID: gid}, nil
}

// NewOwnershipDiff creates a new diff
func NewOwnershipDiff(path string, p OSProxy, ownership *Ownership) (*OwnershipDiff, error) {
	currentOwner, err := fileOwnership(p, path)
	if err != nil {
		return nil, err
	}
	diff := &OwnershipDiff{path: path, p: p}
	if currentOwner.UID != ownership.UID {
		diff.UIDs = &[2]int{currentOwner.UID, ownership.UID}
	}
	if currentOwner.GID != ownership.GID {
		diff.GIDs = &[2]int{currentOwner.GID, ownership.GID}
	}
	return diff, nil
}
