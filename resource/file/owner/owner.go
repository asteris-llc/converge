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
	"strconv"
	"syscall"

	"github.com/asteris-llc/converge/resource"
)

// Owner monitors the file owner of a file
type Owner struct {
	User        *user.User
	Destination string
}

// Check whether the Destination is owned by the user
// 1. If the destination doesn't exist, do nothing
// 2. If this machine is not a linux system, skip
// 3. If expected owner is the actual owner, skip
// 4. If the expected owner and actual owner differ, change
// file to be owned by the actual owner.
func (t *Owner) Check() (resource.TaskStatus, error) {
	diffs := make(map[string]resource.Diff)
	stat, err := os.Stat(t.Destination)

	if os.IsNotExist(err) {
		diffs[t.Destination] = &FileUserDiff{Expected: t.User}
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

	//Cast stat to a statT, this won't work on windows
	statT, ok := stat.Sys().(*syscall.Stat_t)
	if !ok {
		return nil, fmt.Errorf("file.owner does not currently work on non linux systems")
	}

	uid := statT.Uid
	actualUser, err := user.LookupId(fmt.Sprintf("%v", uid))
	if err != nil {
		return nil, err
	}

	// Assume no changes are needed
	ownerDiff := &FileUserDiff{Expected: t.User, Actual: actualUser}
	diffs[t.Destination] = ownerDiff
	warningLevel := resource.StatusWontChange
	// If changes are needed update warning level
	if ownerDiff.Changes() {
		warningLevel = resource.StatusWillChange
	}

	status := fmt.Sprintf("owner of file %q is %q should be %q", t.Destination, actualUser.Username, t.User.Username)
	return &resource.Status{
		Status:       status,
		WarningLevel: warningLevel,
		WillChange:   ownerDiff.Changes(),
		Differences:  diffs,
		Output: []string{
			fmt.Sprintf("%q exist", t.Destination),
			status,
		},
	}, nil
}

// Change the owner of the file
func (o *Owner) Apply() error {
	uid, _ := strconv.Atoi(o.User.Uid)
	gid, _ := strconv.Atoi(o.User.Gid)
	return os.Chown(o.Destination, uid, gid)
}

func (t *Owner) Validate() error {
	if t.Destination == "" {
		return fmt.Errorf("task requires a %q parameter", "destination")
	}
	return nil
}

type FileUserDiff struct {
	Expected *user.User
	Actual   *user.User
}

func (diff *FileUserDiff) Original() string {
	return fmt.Sprint(diff.Actual.Username)
}

func (diff *FileUserDiff) Current() string {
	return fmt.Sprintf(diff.Expected.Username)
}

func (diff *FileUserDiff) Changes() bool {
	return diff.Actual.Username != diff.Expected.Username
}
