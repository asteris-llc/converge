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

// Package file utilites for managing file resources
package file

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"strconv"
	"syscall"
)

// Type determines the file type and returns a string {directory, file, symlink..}
func Type(fi os.FileInfo) (string, error) {

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

// UnixMode takes a string and converts it to a FileMode. If string is not set,
// return -1
func UnixMode(permissions string) (os.FileMode, error) {
	if permissions == "" {
		return os.FileMode(0), nil
	}

	mode, err := strconv.ParseUint(permissions, 8, 32)
	if err != nil {
		return os.FileMode(0), fmt.Errorf("%q is not a valid file mode", permissions)
	}
	return os.FileMode(mode), err
}

// UID returns the Unix Uid of a File
func UID(fi os.FileInfo) int {
	return int(fi.Sys().(*syscall.Stat_t).Uid)
}

// GID returns the Unix Gid of a File
func GID(fi os.FileInfo) int {
	return int(fi.Sys().(*syscall.Stat_t).Gid)
}

// Owner Returns the Unix username of a File
func UserInfo(fi os.FileInfo) (*user.User, error) {
	uid := UID(fi)

	user, err := user.LookupId(strconv.Itoa(uid))
	if err != nil {
		return nil, fmt.Errorf("unable to get username for uid %d", uid)
	}

	return user, nil

}

// Group returns the Unix groupname of a File
func GroupInfo(fi os.FileInfo) (*user.Group, error) {
	gid := GID(fi)

	group, err := user.LookupGroupId(strconv.Itoa(gid))
	if err != nil {
		return &user.Group{}, fmt.Errorf("unable to get group name for gid %d", gid)
	}

	return group, nil

}

//given two users, decide which one to use
//returns true if a change is required
func desiredUser(f, actual *user.User) (userInfo *user.User, changed bool, err error) {
	switch f.Username {
	case "":
		if actual.Username == "" { // if neither is set, use the effective uid of the process
			userInfo, err = user.LookupId(strconv.Itoa(os.Geteuid()))
			if err != nil {
				return &user.User{}, true, fmt.Errorf("unable to set default username %s", err)
			}
			return userInfo, true, err
		}
		if actual.Username != "" {
			return &user.User{Username: actual.Username, Uid: actual.Uid}, true, err
		}
	default:
		userInfo, err = user.Lookup(f.Username)
		if err != nil {
			return userInfo, true, fmt.Errorf("unable to get user information for username %s:", f.Username, err)
		}
		changed = true
	}
	return userInfo, changed, err
}

//given two users, decide which one to use
//returns true if a change is required
func desiredGroup(f, actual *user.Group) (groupInfo *user.Group, changed bool, err error) {
	switch f.Name {
	case "":
		if actual.Name == "" { // if neither is set, use the effective uid of the process
			groupInfo, err = user.LookupGroupId(strconv.Itoa(os.Getegid()))
			if err != nil {
				return &user.Group{}, true, fmt.Errorf("unable to set default group %s", err)
			}
			changed = true
		}
		if actual.Name != "" { //if we didn't request a group, use the file's information
			return &user.Group{Name: actual.Name, Gid: actual.Gid}, false, nil
		}
	default: //we asked to set a group on the file
		groupInfo, err = user.LookupGroup(f.Name)
		if err != nil {
			return groupInfo, true, fmt.Errorf("unable to get user information for username %s:", f.Name, err)
		}
		changed = true

	}
	return groupInfo, changed, err
}

// Content reads a file's contents
func Content(filename string) ([]byte, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("unable to open %s: %s", filename, err)
	}
	return b, err
}

// SameLink checks if two files are the same inode
func SameFile(file1, file2 string) bool {
	fi1, err := os.Lstat(file1)
	if err != nil {
		return false
	}
	fi2, err := os.Lstat(file2)
	if err != nil {
		return false
	}
	return os.SameFile(fi1, fi2)

}
