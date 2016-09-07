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

// Utilites for managing file resources
package file

import (
	"fmt"
	"os"
	"os/user"
	"strconv"
	"syscall"
)

// Determines the file Type and returns a string {directory, file, symlink..}
func FileType(fi os.FileInfo) (string, error) {

	switch mode := fi.Mode(); {
	case mode.IsRegular():
		return "file", nil
	case mode.IsDir():
		return "directory", nil
	case mode&os.ModeSymlink == os.ModeSymlink:
		return "symlink", nil
	default:
		return "", fmt.Errorf("unsupported filetype for %s", fi.Name())
	}
}

// UnixMode takes a string and converts it to a FileMode suitable for use
// in os.Chmod
func UnixMode(permissions string) (os.FileMode, error) {
	if permissions == "" {
		return defaultPermissions, nil
	}

	mode, err := strconv.ParseUint(permissions, 8, 32)
	if err != nil {
		return os.FileMode(0), fmt.Errorf("%q is not a valid file mode", permissions)
	}
	return os.FileMode(mode), err
}

// Returns the Unix Uid of a File
func FileUid(fi os.FileInfo) int {
	return int(fi.Sys().(*syscall.Stat_t).Uid)
}

// Returns the Unix Gid of a File
func FileGid(fi os.FileInfo) int {
	return int(fi.Sys().(*syscall.Stat_t).Gid)
}

// Returns the Unix username of a File
func FileOwner(fi os.FileInfo) (string, error) {
	uid := FileUid(fi)

	user, err := user.LookupId(strconv.Itoa(uid))
	if err != nil {
		return "", fmt.Errorf("unable to get username for uid %d", uid)
	}

	return user.Username, nil

}

// Returns the Unix groupname of a File
func FileGroup(fi os.FileInfo) (string, error) {
	gid := FileGid(fi)

	group, err := user.LookupGroupId(strconv.Itoa(gid))
	if err != nil {
		return "", fmt.Errorf("unable to get username for gid %d", gid)
	}

	return group.Name, nil

}
