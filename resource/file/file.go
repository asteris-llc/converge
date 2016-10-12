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

package file

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strconv"

	"github.com/asteris-llc/converge/resource"
	"github.com/pkg/errors"
)

const defaultPermissions = os.FileMode(0750)

// Action is the calculated action
type Action uint8

const (
	// ActionNone indicates no action is required
	ActionNone Action = iota

	// ActionCreate creates a File resource
	ActionCreate

	// ActionModify deletes a File resource
	ActionModify

	// ActionDelete deletes a File resource
	ActionDelete
)

func (a Action) String() string {
	return string(a)
}

// State is the desired state of the file
type State string

const (
	// StateAbsent indicates the file should be absent
	StateAbsent State = "absent"

	// StateUndefined indicates an undefined state
	StateUndefined State = ""

	// StatePresent indicates the file should be present
	StatePresent State = "present"

	// DefaultState is the default state
	DefaultState State = StatePresent
)

// ValidStates indicates supported states that can be defined
var ValidStates = []State{StateAbsent, StatePresent}

func (s State) String() string {
	return string(s)
}

// Type is the type of file
type Type string

const (
	// TypeDirectory is a directory
	TypeDirectory Type = "directory"

	// TypeFile is a regular file
	TypeFile Type = "file"

	// TypeLink is a hardlink
	TypeLink Type = "hardlink"

	// TypeNone is an undefined type
	TypeNone Type = ""

	// TypeSymlink is a symlink
	TypeSymlink Type = "symlink"

	// DefaultType is the default type
	DefaultType Type = TypeFile
)

var (
	// ValidFileTypes are valid file types
	ValidFileTypes = []Type{TypeDirectory, TypeFile}

	// ValidLinkTypes are valid link types
	ValidLinkTypes = []Type{TypeLink, TypeSymlink}

	// AllTypes are the current FileTypes that are supported
	AllTypes = append(ValidFileTypes, ValidLinkTypes...)
)

func (t Type) String() string {
	return string(t)
}

// File contains information for managing files
type File struct {
	Destination string
	State       State
	Type        Type
	Target      string
	Force       bool    //force replacement of file
	Mode        *uint32 //requested permissions from Prepare
	UserInfo    *user.User
	GroupInfo   *user.Group
	Content     []byte

	action Action //create, delete, modify

	renderer resource.Renderer
}

// New returns a File with UserInfo and GroupInfo fields allocated
func New() *File {
	return &File{
		UserInfo:  &user.User{},
		GroupInfo: &user.Group{},
	}
}

// Apply changes to file resources
func (f *File) Apply() (resource.TaskStatus, error) {

	status, err := f.diff()
	if err != nil || status.Level == resource.StatusFatal {
		return status, errors.Wrap(err, "apply")
	}

	switch f.action {
	case ActionDelete:
		err = f.Delete()
	case ActionCreate, ActionModify:
		err = f.Modify(status)
	}
	if err != nil {
		status.Level = resource.StatusFatal
		status.AddMessage(err.Error())
	}
	return status, err
}

// Check File settings
func (f *File) Check(r resource.Renderer) (resource.TaskStatus, error) {

	f.renderer = r

	status, err := f.diff()
	status.AddMessage(fmt.Sprintf("%s (%s)", f.Destination, f.Type))
	if err != nil || status.Level == resource.StatusFatal {
		return status, errors.Wrap(err, "check")
	}

	return status, err
}

// Compare desired state with an actual file (if present)
func (f *File) diff() (*resource.Status, error) {

	status := &resource.Status{}
	var stat os.FileInfo
	err := f.Validate()

	if err != nil {
		return status, err
	}

	stat, err = os.Lstat(f.Destination)

	if os.IsNotExist(err) { // file not found
		switch f.State {
		case StateAbsent: // if "absent" is set and the file doesn't exist, return with no changes
			return status, nil

		case StatePresent: // file doesn't exist, we need to create it
			f.action = ActionCreate
			actual := &File{
				Destination: "<file does not exist>",
				State:       StateAbsent,
				Type:        f.Type,
				UserInfo:    &user.User{},
				GroupInfo:   &user.Group{},
			}
			err = f.diffFile(actual, status)
			if err != nil {
				status.Level = resource.StatusFatal
				return status, errors.Wrapf(err, "diff")
			}
			return status, nil
		default:
			return status, fmt.Errorf("unknown state %s", f.State)
		}
	} else {
		actual := &File{
			Destination: f.Destination,
			State:       StatePresent,
		}
		switch f.State {
		case StateAbsent: //file exists -> absent
			status.Level = resource.StatusWillChange
			f.action = ActionDelete
			return status, nil
		case StatePresent: //file exists -> modified file
			err = GetFileInfo(actual, stat)
			if err != nil {
				status.Level = resource.StatusFatal
				return status, errors.Wrapf(err, "diff: unable to get file info for %s", f.Destination)
			}
			if actual.Type == TypeFile {
				actual.Content, err = Content(actual.Destination)
				if err != nil {
					return status, errors.Wrap(err, "diff: unable to get content")
				}
			}
			err = f.diffFile(actual, status)
			if err != nil {
				return status, errors.Wrap(err, "diff")
			}

			if status.Level == resource.StatusWillChange {
				f.action = ActionModify
			}
		}
	}
	return status, err
}

// diffFile computes the difference between desired and actual state
func (f *File) diffFile(actual *File, status *resource.Status) error {
	if f.Destination != actual.Destination {
		status.AddDifference("destination", actual.Destination, f.Destination, "")
	}

	if f.State != actual.State {
		status.AddDifference("state", actual.State.String(), f.State.String(), "")
	}

	if f.Type != actual.Type && f.Type != TypeLink {
		switch f.Force {
		case true:
			status.AddDifference("type", actual.Type.String(), f.Type.String(), "")
		default:
			status.Level = resource.StatusCantChange
			return fmt.Errorf("please set force=true to change the type of file")
		}

	}

	if f.Type == TypeLink || f.Type == TypeSymlink {
		err := f.diffLink(actual, status)
		if err != nil {
			status.Level = resource.StatusFatal
			return err
		}
	}

	f.diffMode(actual, status)

	// determine if the file owner needs to be changed
	user, userChanges, err := desiredUser(f.UserInfo, actual.UserInfo)
	if err != nil {
		status.Level = resource.StatusFatal
		return err
	}

	f.UserInfo = user

	if userChanges {
		if user.Username != actual.UserInfo.Username {
			status.AddDifference("username", actual.UserInfo.Username, user.Username, "")
		}

		if user.Uid != actual.UserInfo.Uid {
			status.AddDifference("uid", actual.UserInfo.Uid, user.Uid, "")
		}
	}

	// determine if the file owner needs to be changed
	group, groupChanges, err := desiredGroup(f.GroupInfo, actual.GroupInfo)
	if err != nil {
		status.Level = resource.StatusFatal
		return err
	}
	f.GroupInfo = group

	if groupChanges == true {
		if f.GroupInfo.Name != actual.GroupInfo.Name {
			status.AddDifference("group", actual.GroupInfo.Name, f.GroupInfo.Name, "")
		}

		if f.GroupInfo.Gid != actual.GroupInfo.Gid {
			status.AddDifference("gid", actual.GroupInfo.Gid, f.GroupInfo.Gid, "")
		}
	}

	// only check content on file types
	if f.Type == TypeFile {
		fHash := hash(f.Content)
		actualHash := hash(actual.Content)

		if fHash != actualHash {
			status.AddDifference("content", string(actual.Content), string(f.Content), "")
		}
	}

	if resource.AnyChanges(status.Differences) {
		status.Level = resource.StatusWillChange
	}

	return nil
}

// generates diffs for hard and symlinks
func (f *File) diffLink(actual *File, status *resource.Status) error {
	var err error
	switch f.Type {
	case TypeSymlink:
		if f.Target != actual.Target {
			status.AddDifference("target", actual.Target, f.Target, "")
			return err
		}

	case TypeLink:
		tgt, err := os.Stat(f.Target)
		if os.IsNotExist(err) || err != nil {
			status.Level = resource.StatusFatal
			return errors.Wrap(err, "hardlink target lookup")
		}

		fi, err := os.Stat(f.Destination)

		if os.IsNotExist(err) { // create link
			status.AddDifference("hardlink", f.Destination, f.Target, "")
			return nil
		}
		if err == nil {
			same := os.SameFile(tgt, fi)
			if !same {
				actual.Type = TypeLink
				status.AddDifference("hardlink", f.Destination, f.Target, "")
				return nil
			}
		}
	}
	return err
}

// keep existing permissions if new ones are not requested
func (f *File) diffMode(actual *File, status *resource.Status) {

	if f.Type == TypeSymlink {
		return
	}

	switch f.Mode {
	case nil:
		switch actual.Mode {
		case nil: // use default perms if nothing is set
			m := ModeType(uint32(defaultPermissions), f.Type)
			f.Mode = &m
			status.AddDifference("permissions", os.FileMode(0).String(), os.FileMode(*f.Mode).String(), "")
		default:
			f.Mode = actual.Mode
		}
	default: // compare to file permissions
		switch actual.Mode {
		case nil:
			status.AddDifference("permissions", os.FileMode(0).String(), os.FileMode(*f.Mode).String(), "")
		default:
			if *actual.Mode != *f.Mode {
				status.AddDifference("permissions", os.FileMode(*actual.Mode).String(), os.FileMode(*f.Mode).String(), "")
			}
		}
	}
}

// GetFileInfo populates a File struct with data from a file on the system
func GetFileInfo(f *File, fi os.FileInfo) error {
	var err error

	if f.State == StateUndefined {
		f.State = StatePresent
	}

	f.Type, err = GetType(fi)
	if err != nil {
		return errors.Wrapf(err, "error determining type")
	}

	// set perms & higher bits for file type
	// in POSIX permissions are not set for symlinks
	if f.Type != TypeSymlink {
		f.Mode = new(uint32)
		*f.Mode = ModeType(uint32(fi.Mode()), f.Type)
	}

	// follow symlinks
	if f.Type == TypeSymlink {
		f.Target, err = os.Readlink(f.Destination)
		if err != nil {
			return errors.Wrapf(err, "error determining target of symlink")
		}
	}

	f.UserInfo, err = UserInfo(fi)
	if err != nil {
		return errors.Wrapf(err, "error determining owner")
	}

	f.GroupInfo, err = GroupInfo(fi)
	if err != nil {
		return errors.Wrapf(err, "error determining group")
	}

	return err

}

// hash is used to compare file content
func hash(b []byte) string {
	sha := sha256.Sum256(b)
	return hex.EncodeToString(sha[:])
}

// Delete a file Listed by the File Resource
func (f *File) Delete() error {
	var err error
	_, err = os.Lstat(f.Destination)

	if os.IsNotExist(err) {
		return nil
	}

	switch f.Type {
	case TypeDirectory:
		err = os.RemoveAll(f.Destination)
	default:
		err = os.Remove(f.Destination)
	}
	return err
}

// Modify a file based on a file resource
func (f *File) Modify(status *resource.Status) error {
	var err error

	if status.HasDifference("type") || status.HasDifference("hardlink") && f.action == ActionModify {
		switch f.Force {
		case true:
			// if the file type changes, delete and recreate
			err = f.Delete()
			if err != nil {
				return errors.Wrapf(err, "modify: unable to recreate")
			}
			f.action = ActionCreate
		default:
			return errors.New("modify: please set force=true to change file")
		}
	}

	switch f.Type {
	case TypeDirectory:
		err = os.MkdirAll(f.Destination, os.FileMode(*f.Mode))
		if err != nil {
			return errors.Wrap(err, "modify: unable to create directory")
		}

	case TypeFile:
		if _, ok := status.Difference("content"); ok || f.action == ActionCreate {

			// parent directory check
			d := filepath.Dir(f.Destination)
			_, err := os.Stat(d)

			if os.IsNotExist(err) {
				switch f.Force {
				case true:
					err = os.MkdirAll(d, os.FileMode(*f.Mode))
					if err != nil {
						return errors.Wrap(err, "modify: unable to create directory")
					}
				default:
					return errors.Wrapf(err, "modify: parent directory is missing (set force = \"true\" to create it)")
				}
			}

			err = ioutil.WriteFile(f.Destination, f.Content, os.FileMode(*f.Mode))
			if err != nil {
				return errors.Wrapf(err, "modify")
			}
		}

	case TypeLink:
		if _, ok := status.Difference(TypeLink.String()); ok {
			err := os.Link(f.Target, f.Destination)
			if err != nil {
				return errors.Wrapf(err, "modify: unable to create hardlink %s -> %s", f.Destination, f.Target)
			}
		}

	case TypeSymlink:
		if _, ok := status.Difference("target"); ok {
			_, err := os.Stat(f.Destination)
			if err == nil { //symlink already exists
				switch f.Force {
				case true:
					err := f.Delete()
					if err != nil {
						return errors.Wrapf(err, "modify: unable to delete existing symlink")
					}
				default:
					return errors.Wrap(err, "modify: symlink already exists, set force=true to replace")
				}
			}
			err = os.Symlink(f.Target, f.Destination)
			if err != nil {
				return errors.Wrapf(err, "modify: unable to create")
			}
		}
	}

	_, changeUser := status.Difference("uid")
	_, changeGroup := status.Difference("gid")

	if changeUser || changeGroup {
		uid, err := strconv.Atoi(f.UserInfo.Uid)
		if err != nil {
			return errors.Wrapf(err, "modify: uid")
		}
		gid, err := strconv.Atoi(f.GroupInfo.Gid)
		if err != nil {
			return errors.Wrapf(err, "modify: gid")
		}
		err = os.Lchown(f.Destination, uid, gid)
		if err != nil {
			return errors.Wrapf(err, "modify: owner/group")
		}
	}

	if _, ok := status.Difference("permissions"); ok {
		err = os.Chmod(f.Destination, os.FileMode(*f.Mode).Perm())
		if err != nil {
			return errors.Wrapf(err, "modify: permissions")
		}
	}
	return nil
}
