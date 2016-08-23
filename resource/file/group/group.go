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

package group

import (
	"fmt"
	"os"
	"os/user"
	"strconv"
	"syscall"

	"github.com/asteris-llc/converge/resource"
)

// Mode monitors the file mode of a file
type Group struct {
	Group       *user.Group
	Destination string
}

// Check whether the Destination is owned by the user
func (t *Group) Check() (resource.TaskStatus, error) {
	diffs := make(map[string]resource.Diff)
	stat, err := os.Stat(t.Destination)
	if os.IsNotExist(err) {
		diffs[t.Destination] = &FileGroupDiff{Expected: t.Group}
		status := fmt.Sprintf("%q does not exist", t.Destination)
		return &resource.Status{
			Status:       status,
			WarningLevel: resource.StatusFatal,
			WillChange:   false,
			Differences:  diffs,
			Output:       []string{status},
		}, nil
	} else if err != nil {
		return nil, err
	}
	statT, ok := stat.Sys().(*syscall.Stat_t)
	if !ok {
		return nil, fmt.Errorf("file.group does not currently work on non linux systems")
	}
	gid := statT.Gid
	actualGroup, err := user.LookupGroupId(fmt.Sprintf("%v", gid))
	if err != nil {
		return nil, err
	}
	groupDiff := &FileGroupDiff{Expected: t.Group, Actual: actualGroup}
	diffs[t.Destination] = groupDiff
	warningLevel := resource.StatusWontChange
	if groupDiff.Changes() {
		warningLevel = resource.StatusWillChange
	}
	status := fmt.Sprintf("file belongs to group %q should be %q", actualGroup.Name, t.Group.Name)
	return &resource.Status{
		Status:       status,
		WarningLevel: warningLevel,
		WillChange:   groupDiff.Changes(),
		Differences:  diffs,
		Output: []string{
			fmt.Sprintf("%q exist", t.Destination),
			status,
		},
	}, nil
}

// Apply the changes in mode
func (t *Group) Apply() error {
	stat, err := os.Stat(t.Destination)
	if err != nil {
		return err
	}
	statT, ok := stat.Sys().(*syscall.Stat_t)
	if !ok {
		return fmt.Errorf("file.owner does not currently work on non linux systems")
	}
	uid := int(statT.Uid)
	gid, _ := strconv.Atoi(t.Group.Gid)
	return os.Chown(t.Destination, uid, gid)
}

func (t *Group) Validate() error {
	if t.Destination == "" {
		return fmt.Errorf("task requires a %q parameter", "destination")
	}
	if t.Group == nil {
		return fmt.Errorf("could not locate group")
	}
	return nil
}

type FileGroupDiff struct {
	Expected *user.Group
	Actual   *user.Group
}

func (diff *FileGroupDiff) Original() string {
	if diff.Actual != nil {
		return fmt.Sprint(diff.Actual.Name)
	}
	return ""
}

func (diff *FileGroupDiff) Current() string {
	if diff.Expected != nil {
		return fmt.Sprintf(diff.Expected.Name)
	}
	return ""
}

func (diff *FileGroupDiff) Changes() bool {
	if diff.Actual != nil && diff.Expected != nil {
		return diff.Actual.Gid != diff.Expected.Gid
	}
	return false
}
