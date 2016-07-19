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

package user

import (
	"fmt"
	"os"
	"os/user"
	"strconv"
	"syscall"
)

// Mode monitors the file mode of a file
type User struct {
	username    string
	uid         string
	gid         string
	destination string
}

// Check whether the destination has the right mode
func (u *User) Check() (status string, willChange bool, err error) {
	stat, err := os.Stat(u.destination)
	if err != nil {
		return "", false, err
	}

	statT, ok := stat.Sys().(*syscall.Stat_t)
	if !ok {
		err = fmt.Errorf("file.owner does not currently work on non linux systems\n")
		return "", false, err
	}

	uid := statT.Uid
	actualUser, err := user.LookupId(fmt.Sprintf("%v", uid))
	if err != nil {
		return "", false, err
	}

	return actualUser.Username, actualUser.Username != u.username, nil
}

// Apply the changes in mode
func (u *User) Apply() error {
	uid, _ := strconv.Atoi(u.uid)
	gid, _ := strconv.Atoi(u.gid)
	return os.Chown(u.destination, uid, gid)
}
