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
	"strconv"
	"strings"

	"github.com/asteris-llc/converge/resource"
	"github.com/pkg/errors"
)

const defaultPermissions = os.FileMode(0750)

const defaultState = "present"

var validStates = []string{"present", "absent"}

const defaultType = "file"

var validFileTypes = []string{"directory", "file"}
var validLinkTypes = []string{"hardlink", "symlink"}

// File contains information for managing files
type File struct {
	Destination   string
	State         string
	Type          string
	Target        string
	Force         bool //force replacement of symlinks, etc.
	FileMode      os.FileMode
	UserInfo      *user.User
	GroupInfo     *user.Group
	Content       []byte
	action        string //create, delete, modify
	modifyContent bool   // does content need to be changed
}

// Apply  changes to file resources
func (f *File) Apply() error {
	var err error
	switch f.action {
	case "delete":
		err = f.Delete()
	case "create":
		err = f.Create()
	case "modify":
		err = f.Modify()
	}
	if err != nil {
		return errors.Wrapf(err, "%s on %s failed: %s", f.action, f.Destination)
	}
	return err
}

// Check File settings
func (f *File) Check() (resource.TaskStatus, error) {
	status := &resource.Status{Status: f.Destination}
	status.Output = append(status.Output, fmt.Sprintf("%s (%s)", f.Destination, f.Type))
	var actual *File
	// Get information about the current file
	stat, err := os.Lstat(f.Destination) //link aware

	if os.IsNotExist(err) { //file not found
		switch f.State {
		case "absent": // if "absent" is set and the file doesn't exist, return with no changes
			status.WillChange = false
			status.WarningLevel = resource.StatusNoChange
			return status, nil
		case "present": //file doesn't exist, we need to create it
			actual = &File{Destination: "<file does not exist>", State: "absent", UserInfo: &user.User{}, GroupInfo: &user.Group{}}
			status.WillChange = true
			status.WarningLevel = resource.StatusWillChange
			f.diffFile(actual, status)
			f.action = "create"
		}
	} else { //file exists
		actual = &File{Destination: f.Destination, State: "present"}
		switch f.State {
		case "absent": //remove file
			status.WillChange = true
			status.AddDifference("destination", actual.Destination, "<removed>", "")
			status.AddDifference("state", actual.State, f.State, "")
			f.action = "delete"
		case "present": //modify file
			err = GetFileInfo(actual, stat)
			if err != nil {
				status.WarningLevel = resource.StatusFatal
				return status, fmt.Errorf("unable to get file info for %s: %s", f.Destination, err)
			}
			actual.Content, _ = Content(actual.Destination)
			f.diffFile(actual, status)
			if status.WillChange {
				f.action = "modify"
			}

		}
	}

	return status, nil
}

// Validate runs checks against a File resource
func (f *File) Validate() error {
	var err error
	if f.Destination == "" {
		return fmt.Errorf("file requires a destination parameter")
	}

	err = f.validateState()
	if err != nil {
		return err
	}

	err = f.validateType()
	if err != nil {
		return err
	}

	// links should have a target
	err = f.validateTarget()
	if err != nil {
		return err
	}

	err = f.validateUser()
	if err != nil {
		return err
	}

	err = f.validateGroup()
	if err != nil {
		return err
	}

	return err
}

// Validate the state or set default value
func (f *File) validateState() error {
	var err error

	switch f.State {
	case "": //nothing set, use default
		f.State = defaultState
		return nil
	default:
		for _, s := range validStates {
			if f.State == s {
				return nil
			}
		}
		return fmt.Errorf("state should be one of %s, got %q", strings.Join(validStates, ", "), f.State)
	}
	return err
}

// Validate the type or set default value
func (f *File) validateType() error {
	var allTypes []string
	allTypes = append(allTypes, validFileTypes...)
	allTypes = append(allTypes, validLinkTypes...)
	switch f.Type {
	case "": //use default if not set
		f.Type = defaultType
		return nil
	default:
		for _, t := range allTypes {
			if f.Type == t {
				return nil
			}
		}
		return fmt.Errorf("type should be one of %s, got %q", strings.Join(allTypes, ", "), f.Type)
	}
	return nil

}

// A target needs to be set if you are creating a link
func (f *File) validateTarget() error {

	switch f.Target {
	case "":
		if f.Type == "symlink" || f.Type == "hardlink" {
			return fmt.Errorf("must define a target if you are using a %q", f.Type)
		}
		return nil
	default:
		// is target set for a file or directory type?
		if f.Type == "symlink" || f.Type == "hardlink" {
			return nil
		}
		return fmt.Errorf("cannot define target on a type of %q: target: %q", f.Type, f.Target)
	}
	return fmt.Errorf("unknown combination of type %q and target %q", f.Type, f.Target)
}

func (f *File) validateUser() error {
	if f.UserInfo.Username != "" {
		u, err := user.Lookup(f.UserInfo.Username)
		if err != nil {
			return fmt.Errorf("unable to get user information for username %s:", f.UserInfo.Username, err)
		}
		f.UserInfo.Username = u.Username
	}
	return nil
}

//if a group is provided, make sure it exists on the system
func (f *File) validateGroup() error {
	if f.GroupInfo.Name != "" {
		g, err := user.LookupGroup(f.GroupInfo.Name)
		if err != nil {
			return fmt.Errorf("unable to get user information for username %s:", f.GroupInfo.Name, err)
		}
		f.GroupInfo = g
	}
	return nil
}

// func (f *File) validateGroup() error {
// 	if f.GroupInfo.Name = "" {
// 		g, err := user.LookupGroupId(strconv.Itoa(os.Getegid()))
// 		if err != nil {
// 			return fmt.Errorf("unable to set default group %s", err)
// 		}
// 		f.GroupInfo.Name = g.Name
// 	}
// 	return nil
// }

// GetFileInfo populates a File struct with data from a file on the system
func GetFileInfo(f *File, stat os.FileInfo) error {
	var err error

	if f.Destination == "" {
		f.Destination = stat.Name()
	}

	if f.State == "" {
		f.State = "present"
	}

	f.Type, err = Type(stat)
	if err != nil {
		return fmt.Errorf("error determining type of %s : %s", f.Destination, err)
	}

	// follow symlinks
	if f.Type == "symlink" {
		f.Target, err = os.Readlink(f.Destination)
		if err != nil {
			return fmt.Errorf("error determining target of symlink %s : %s", f.Destination, err)
		}
	}

	f.FileMode = stat.Mode() & os.ModePerm //strip out higher order bits from perms

	f.UserInfo, err = UserInfo(stat)
	if err != nil {
		return fmt.Errorf("error determining owner of %s : %s", f.Destination, err)
	}

	f.GroupInfo, err = GroupInfo(stat)
	if err != nil {
		return fmt.Errorf("error determining group of %s : %s", f.Destination, err)
	}
	return err
}

// Compute the difference between desired and actual state
func (f *File) diffFile(actual *File, status *resource.Status) {

	if f.State != actual.State {
		status.AddDifference("state", actual.State, f.State, "")
	}

	if f.Type != actual.Type {
		status.AddDifference("type", actual.Type, f.Type, "")
	}

	if f.Target != actual.Target {
		status.AddDifference("target", actual.Target, f.Target, "")
	}

	if f.Type == "hardlink" && !SameFile(f.Destination, actual.Target) {
		status.AddDifference("hardlink inode", actual.Target, f.Target, "")
	}

	if f.FileMode != 0 && f.FileMode != actual.FileMode {
		status.AddDifference("permissions", actual.FileMode.String(), f.FileMode.String(), "")
	}
	// determine if the file owner needs to be changed
	desired, userChanges, err := desiredUser(f.UserInfo, actual.UserInfo)
	if err != nil {
		status.WarningLevel = resource.StatusFatal
	}

	if userChanges {
		if desired.Username != actual.UserInfo.Username {
			status.AddDifference("username", actual.UserInfo.Username, desired.Username, "")
		}

		if desired.Uid != actual.UserInfo.Uid {
			status.AddDifference("uid", actual.UserInfo.Uid, desired.Uid, "")
		}
	}

	// determine if the file owner needs to be changed
	desiredGrp, groupChanges, err := desiredGroup(f.GroupInfo, actual.GroupInfo)
	if err != nil {
		status.WarningLevel = resource.StatusFatal
	}

	if groupChanges == true {
		if desiredGrp.Name != actual.GroupInfo.Name {
			status.AddDifference("group", actual.GroupInfo.Name, desiredGrp.Name, "")
		}

		if desiredGrp.Gid != actual.GroupInfo.Gid {
			status.AddDifference("gid", actual.GroupInfo.Gid, desiredGrp.Gid, "")
		}
	}

	fHash := hash(f.Content)
	actualHash := hash(actual.Content)

	if fHash != actualHash {
		status.AddDifference("content", string(actual.Content), string(f.Content), "")
		f.modifyContent = true
	}

	if resource.AnyChanges(status.Differences) {
		status.WillChange = true
		status.WarningLevel = resource.StatusWillChange
	}
}

func hash(b []byte) string {
	sha := sha256.Sum256(b)
	return hex.EncodeToString(sha[:])
}

// Create a file from File information
func (f *File) Create() error {
	var err error
	switch f.Type {
	case "file":
		err = ioutil.WriteFile(f.Destination, f.Content, f.FileMode)
		if err != nil {
			return fmt.Errorf("unable to write file %s: %s", f.Destination, err)
		}

	case "directory":
		err = os.MkdirAll(f.Destination, f.FileMode)
		if err != nil {
			return fmt.Errorf("unable to create directory %s: %s", f.Destination, err)
		}
	case "symlink":
		err := os.Symlink(f.Target, f.Destination)
		if err != nil {
			return fmt.Errorf("unable to create symlink %s: %s", f.Destination, err)
		}
	case "hardlink":
		err := os.Link(f.Target, f.Destination)
		if err != nil {
			return fmt.Errorf("unable to create hardlink %s: %s", f.Destination, err)
		}
	}

	tgtUser, _, err := desiredUser(f.UserInfo, &user.User{})
	if err != nil {
		return fmt.Errorf("unable to get file owner information %s: %s", f.Destination, err)
	}
	tgtGroup, _, err := desiredGroup(f.GroupInfo, &user.Group{})
	if err != nil {
		return fmt.Errorf("unable to get file group information %s: %s", f.Destination, err)
	}
	uid, _ := strconv.Atoi(tgtUser.Uid)
	gid, _ := strconv.Atoi(tgtGroup.Gid)
	err = os.Chown(f.Destination, uid, gid)
	if err != nil {
		return fmt.Errorf("unable to change permissions on file %s: %s", f.Destination, err)
	}

	return err
}

// Delete a file Listed by the File Resource
func (f *File) Delete() error {
	var err error
	switch f.Type {
	case "directory":
		err = os.RemoveAll(f.Destination)
	default:
		err = os.Remove(f.Destination)
	}
	return err
}

// Modify a file based on a file resource
func (f *File) Modify() error {
	var err error

	actual := &File{Destination: f.Destination}
	stat, err := os.Lstat(f.Destination)
	if err != nil {
		return fmt.Errorf("apply: unable to get information about %s: %s", f.Destination, err)
	}
	err = GetFileInfo(actual, stat)
	if err != nil {
		return fmt.Errorf("apply: unable to get information about %s: %s", f.Destination, err)
	}

	// if the file type changes, delete and recreate
	if f.Type != actual.Type {
		err := f.Delete()
		if err != nil {
			return fmt.Errorf("apply: unable to recreate %s: %s", f.Destination, err)
		}
		err = f.Create()
		if err != nil {
			return fmt.Errorf("apply: unable to recreate %s: %s", f.Destination, err)
		}
		return nil

		if f.Target != actual.Target {

		}
	}

	if f.modifyContent && f.Type == "file" {
		err = ioutil.WriteFile(f.Destination, f.Content, f.FileMode)
		if err != nil {
			return fmt.Errorf("unable to write file %s: %s", f.Destination, err)
		}
	}

	//only modify gid/uid of a file if it has been requested
	tgtUser, userChanges, err := desiredUser(f.UserInfo, actual.UserInfo)
	if err != nil {
		return fmt.Errorf("unable to get file owner information %s: %s", f.Destination, err)
	}
	tgtGroup, groupChanges, err := desiredGroup(f.GroupInfo, actual.GroupInfo)
	if err != nil {
		return fmt.Errorf("unable to get file group information %s: %s", f.Destination, err)
	}

	if userChanges || groupChanges {
		uid, _ := strconv.Atoi(tgtUser.Uid)
		gid, _ := strconv.Atoi(tgtGroup.Gid)
		err = os.Chown(f.Destination, uid, gid)
		if err != nil {
			return fmt.Errorf("unable to change ownership on %s: %s", f.Destination, err)
		}
	}

	if f.FileMode != actual.FileMode {
		err = os.Chmod(f.Destination, f.FileMode)
		if err != nil {
			return fmt.Errorf("unable to change permissions on %s: %s", f.Destination, err)
		}
	}

	if f.Target != actual.Target {
		switch f.Type {
		case "symlink":
			err := os.Symlink(f.Target, f.Destination)
			if err != nil {
				return fmt.Errorf("unable to create symlink %s: %s", f.Destination, err)
			}
		case "hardlink":
			if !SameFile(f.Destination, actual.Target) {
				err := os.Link(f.Target, f.Destination)
				if err != nil {
					return fmt.Errorf("unable to create link %s: %s", f.Destination, err)
				}
			}
		}
	}

	return err
}
