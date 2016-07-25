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
	"syscall"
)

// Mode monitors the file mode of a file
type Owner struct {
	Username    string
	Uid         int
	Gid         int
	Destination string
}

// Check whether the Destination has the right mode
func (o *Owner) Check() (status string, willChange bool, err error) {
	stat, err := os.Stat(o.Destination)
	if err != nil {
		return "", false, err
	}

	statT, ok := stat.Sys().(*syscall.Stat_t)
	if !ok {
		return "", false, fmt.Errorf("file.owner does not currently work on non linux systems\n")
	}

	Uid := statT.Uid
	actualUser, err := user.LookupId(fmt.Sprintf("%v", Uid))
	if err != nil {
		return "", false, err
	}

	return actualUser.Username, actualUser.Username != o.Username, nil
}

// Apply the changes in mode
func (o *Owner) Apply() error {
	return os.Chown(o.Destination, o.Uid, o.Gid)
}
