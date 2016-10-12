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

	"github.com/pkg/errors"
)

// ModeType sets the higher-order bits of an os.FileMode based on the file type
func ModeType(mode uint32, filetype Type) uint32 {
	m := os.FileMode(mode)
	switch filetype {
	case TypeFile:
		m &^= os.ModeType //clear bits if they are set
	case TypeDirectory:
		m |= os.ModeDir
	case TypeSymlink:
		m |= os.ModeSymlink
	}
	return uint32(m)
}

// GetType determines the file type and returns a Type
func GetType(fi os.FileInfo) (Type, error) {
	switch mode := fi.Mode(); {
	case mode.IsRegular():
		return TypeFile, nil
	case mode.IsDir():
		return TypeDirectory, nil
	case mode&os.ModeSymlink == os.ModeSymlink:
		return TypeSymlink, nil
	default:
		return TypeNone, fmt.Errorf("unsupported filetype for %s", fi.Name())
	}
}

// UID returns the Unix Uid of a File
func UID(fi os.FileInfo) (int, error) {
	if stat, ok := fi.Sys().(*syscall.Stat_t); ok && stat != nil {
		return int(stat.Uid), nil
	}
	return 0, fmt.Errorf("UID: stat.Sys failed")
}

// GID returns the Unix Gid of a File
func GID(fi os.FileInfo) (int, error) {
	if stat, ok := fi.Sys().(*syscall.Stat_t); ok && stat != nil {
		return int(stat.Gid), nil
	}
	return 0, fmt.Errorf("GID: stat.Sys failed")
}

// UserInfo Returns the (unix user.User, error) of a file
func UserInfo(fi os.FileInfo) (*user.User, error) {
	uid, err := UID(fi)
	if err != nil {
		return nil, fmt.Errorf("unable to get uid for file")
	}

	user, err := user.LookupId(strconv.Itoa(uid))
	if err != nil {
		return nil, fmt.Errorf("unable to get username for uid %d", uid)
	}

	return user, nil

}

// GroupInfo returns the (Unix user.Group of a File, error)
func GroupInfo(fi os.FileInfo) (*user.Group, error) {
	gid, err := GID(fi)
	if err != nil {
		return nil, fmt.Errorf("unable to get gid for file")
	}

	group, err := user.LookupGroupId(strconv.Itoa(gid))
	if err != nil {
		return &user.Group{}, fmt.Errorf("unable to get group name for gid %d", gid)
	}

	return group, nil

}

//given two users, decide which one to use
//returns true if a change is required
func desiredUser(f, actual *user.User) (userInfo *user.User, changed bool, err error) {
	if f == nil || f.Username == "" {
		if actual == nil || actual.Username == "" { // if neither is set, use the effective uid of the process
			userInfo, err = user.LookupId(strconv.Itoa(os.Geteuid()))
			if err != nil {
				return nil, false, errors.Wrapf(err, "unable to set default username %s", f.Name)
			}
			return userInfo, true, err
		}
		if actual.Username != "" {
			return &user.User{Username: actual.Username, Uid: actual.Uid}, false, err
		}
	}

	userInfo, err = user.Lookup(f.Username)
	if err != nil {
		return userInfo, false, errors.Wrapf(err, "unable to get user information")
	}
	if f.Username != actual.Username {
		changed = true
	}
	return userInfo, changed, err
}

//given two groups, decide which one to use
//returns true if a change is required
func desiredGroup(f, actual *user.Group) (groupInfo *user.Group, changed bool, err error) {
	if f == nil || f.Name == "" {
		if actual == nil || actual.Name == "" { // if neither is set, use the effective gid of the process
			groupInfo, err = user.LookupGroupId(strconv.Itoa(os.Getegid()))
			if err != nil {
				return nil, false, errors.Wrapf(err, "unable to set default group")
			}
			return groupInfo, true, nil
		}
		if actual.Name != "" { //if we didn't request a group, use the file's information
			return &user.Group{Name: actual.Name, Gid: actual.Gid}, false, nil
		}
	}

	groupInfo, err = user.LookupGroup(f.Name)
	if err != nil {
		return groupInfo, false, errors.Wrapf(err, "unable to get information")
	}

	if f.Name != actual.Name {
		changed = true
	}

	return groupInfo, changed, err
}

// Content reads a file's contents
func Content(filename string) ([]byte, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, errors.Wrapf(err, "content")
	}
	return b, err
}

// SameFile checks if two files are the same inode
func SameFile(file1, file2 string) (bool, error) {
	fi1, err := os.Lstat(file1)
	if err != nil {
		return false, err
	}
	fi2, err := os.Lstat(file2)
	if err != nil {
		return false, err
	}

	return os.SameFile(fi1, fi2), nil
}
