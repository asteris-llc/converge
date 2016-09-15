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
	Destination string
	State       string
	Type        string
	Target      string
	Force       bool        //force replacement of symlinks, etc
	Mode        string      //requested permissions from Prepare
	FileMode    os.FileMode //calculated permissions
	UserInfo    *user.User
	GroupInfo   *user.Group
	Content     []byte

	action        string //create, delete, modify
	modifyContent bool   // does content need to be changed

	*resource.Status
	renderer resource.Renderer
}

// Apply  changes to file resources
func (f *File) Apply(r resource.Renderer) (resource.TaskStatus, error) {

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
		return f.Status, errors.Wrapf(err, "%s on %s failed: %s", f.action, f.Destination)
	}
	return f.Status, err
}

// Check File settings
func (f *File) Check(r resource.Renderer) (resource.TaskStatus, error) {
	var err error
	f.renderer = r
	f.Status = &resource.Status{}
	f.Status.Output = append(f.Status.Output, fmt.Sprintf("%s (%s)", f.Destination, f.Type))

	var actual *File

	// Get information about the current file
	stat, err := os.Lstat(f.Destination) //link aware
	if os.IsNotExist(err) {              //file not found
		switch f.State {
		case "absent": // if "absent" is set and the file doesn't exist, return with no changes
			return f, nil
		case "present": //file doesn't exist, we need to create it
			actual = &File{Destination: "<file does not exist>",
				State:     "absent",
				UserInfo:  &user.User{},
				GroupInfo: &user.Group{},
				Status:    &resource.Status{}}
			err = f.diffFile(actual)
			if err != nil {
				f.Status.WarningLevel = resource.StatusFatal
				return f, errors.Wrapf(err, "check")
			}
			f.action = "create"
			return f, nil
		default:
			return f, fmt.Errorf("unknown state %s", f.State)
		}
	} else { //file exists
		actual = &File{
			Destination: f.Destination,
			State:       "present"}
		err = GetFileInfo(actual, stat)
		if err != nil {
			f.Status.WarningLevel = resource.StatusFatal
			return f, errors.Wrapf(err, "unable to get file info for %s", f.Destination)
		}
		actual.Content, err = Content(actual.Destination)
		if err != nil {
			return f, errors.Wrap(err, "unable to get content")
		}

		err = f.diffFile(actual)
		if err != nil {
			return f, errors.Wrap(err, "check")
		}
		switch f.State {
		case "absent": //file exists -> absent
			f.Status.WillChange = true
			f.Status.WarningLevel = resource.StatusWillChange
			f.action = "delete"
		case "present": //file exists -> modified file
			if f.Status.WillChange {
				f.action = "modify"
			}
		}
	}

	if f.Status.WarningLevel == resource.StatusFatal {
		return f, errors.New("failure during check")
	} else {
		return f, nil
	}
}

// Validate runs checks against a File resource
func (f *File) Validate() error {
	var err error
	if f.Destination == "" {
		return errors.New("file requires a destination parameter")
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
			return errors.Wrapf(err, "unable to get user information for username %s: %s", f.UserInfo.Username, err)
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
			return errors.Wrapf(err, "unable to get user information for username %s: %s", f.GroupInfo.Name)
		}
		f.GroupInfo = g
	}
	return nil
}

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
		return errors.Wrapf(err, "error determining type of %s : %s", f.Destination)
	}

	// follow symlinks
	if f.Type == "symlink" {
		f.Target, err = os.Readlink(f.Destination)
		if err != nil {
			return errors.Wrapf(err, "error determining target of symlink %s : %s", f.Destination)
		}
	}

	f.FileMode = stat.Mode() & os.ModePerm //strip out higher order bits from perms

	f.UserInfo, err = UserInfo(stat)
	if err != nil {
		return errors.Wrapf(err, "error determining owner of %s : %s", f.Destination)
	}

	f.GroupInfo, err = GroupInfo(stat)
	if err != nil {
		return errors.Wrapf(err, "error determining group of %s : %s", f.Destination)
	}
	return err
}

// Compute the difference between desired and actual state
func (f *File) diffFile(actual *File) error {
	status := f.Status

	if f.State != actual.State {
		status.AddDifference("state", actual.State, f.State, "")
	}

	if f.Type != actual.Type && f.Type != "hardlink" {
		status.AddDifference("type", actual.Type, f.Type, "")
	}

	switch f.Type {
	case "symlink":
		if f.Target != actual.Target {
			status.AddDifference("target", actual.Target, f.Target, "")
			return nil
		}

	case "hardlink":
		//check to see if target exists
		_, err := os.Stat(f.Target)
		if os.IsNotExist(err) {
			status.WarningLevel = resource.StatusFatal
			return errors.Wrap(err, "hardlink target does not exist")
		}
		if err != nil {
			status.WarningLevel = resource.StatusFatal
			return errors.Wrap(err, "error looking up link target")
		}

		switch f.action {
		case "modify":
			same, err := SameFile(f.Destination, f.Target)

			if err != nil {
				status.WarningLevel = resource.StatusFatal
				return errors.Wrapf(err, "failed to check link status of %s -> %s", f.Destination, f.Target)
			}

			if !same {
				actual.Type = "hardlink"
				status.AddDifference("hardlink", f.Destination, f.Target, "")
				return err
			}
		case "create":
			status.AddDifference("hardlink", f.Destination, f.Target, "")
			return err
		}
	}

	mode := desiredMode(f, actual)
	if mode != actual.FileMode {
		status.AddDifference("permissions", actual.FileMode.String(), mode.String(), "")
	}
	// determine if the file owner needs to be changed
	user, userChanges, err := desiredUser(f.UserInfo, actual.UserInfo)
	if err != nil {
		status.WarningLevel = resource.StatusFatal
		return err
	}

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
		status.WarningLevel = resource.StatusFatal
	}

	if groupChanges == true {
		if group.Name != actual.GroupInfo.Name {
			status.AddDifference("group", actual.GroupInfo.Name, group.Name, "")
		}

		if group.Gid != actual.GroupInfo.Gid {
			status.AddDifference("gid", actual.GroupInfo.Gid, group.Gid, "")
		}
	}

	// only check content on file types
	if f.Type == "file" {
		fHash := hash(f.Content)
		actualHash := hash(actual.Content)

		if fHash != actualHash {
			status.AddDifference("content", string(actual.Content), string(f.Content), "")
			f.modifyContent = true
		}
	}

	if len(status.Differences) > 0 {
		status.WillChange = true
		switch status.WarningLevel {
		case resource.StatusFatal:
		default:
			status.WarningLevel = resource.StatusWillChange
		}

	}

	f.Status = status

	return err
}

// hash is used to compare file content
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
			return errors.Wrapf(err, "unable to write file %s", f.Destination)
		}

	case "directory":
		err = os.MkdirAll(f.Destination, f.FileMode)
		if err != nil {
			return errors.Wrapf(err, "unable to create directory %s", f.Destination)
		}
	case "symlink":
		err := os.Symlink(f.Target, f.Destination)
		if err != nil {
			return errors.Wrapf(err, "unable to create symlink %s", f.Destination)
		}
	case "hardlink":
		err := os.Link(f.Target, f.Destination)
		if err != nil {
			return errors.Wrapf(err, "unable to create hardlink %s", f.Destination)
		}
	}

	tgtUser, _, err := desiredUser(f.UserInfo, &user.User{})
	if err != nil {
		return errors.Wrapf(err, "unable to get file owner information %s", f.Destination)
	}
	tgtGroup, _, err := desiredGroup(f.GroupInfo, &user.Group{})
	if err != nil {
		return errors.Wrapf(err, "unable to get file group information %s", f.Destination)
	}
	uid, _ := strconv.Atoi(tgtUser.Uid)
	gid, _ := strconv.Atoi(tgtGroup.Gid)
	err = os.Chown(f.Destination, uid, gid)
	if err != nil {
		return errors.Wrapf(err, "unable to change permissions on file %s: %s", f.Destination)
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
		return errors.Wrapf(err, "apply: unable to get information about %s: %s", f.Destination)
	}
	err = GetFileInfo(actual, stat)
	if err != nil {
		return errors.Wrapf(err, "apply: unable to get information about %s: %s", f.Destination)
	}

	// if the file type changes, delete and recreate
	if f.Type != actual.Type {
		err := f.Delete()
		if err != nil {
			return errors.Wrapf(err, "apply: unable to recreate %s: %s", f.Destination)
		}
		err = f.Create()
		if err != nil {
			return errors.Wrapf(err, "apply: unable to recreate %s: %s", f.Destination)
		}
		return nil

		if f.Target != actual.Target {

		}
	}

	if f.modifyContent && f.Type == "file" {
		err = ioutil.WriteFile(f.Destination, f.Content, f.FileMode)
		if err != nil {
			return errors.Wrapf(err, "unable to write file %s: %s", f.Destination)
		}
	}

	//only modify gid/uid of a file if it has been requested
	tgtUser, userChanges, err := desiredUser(f.UserInfo, actual.UserInfo)
	if err != nil {
		return errors.Wrapf(err, "unable to get file owner information %s: %s", f.Destination)
	}
	tgtGroup, groupChanges, err := desiredGroup(f.GroupInfo, actual.GroupInfo)
	if err != nil {
		return errors.Wrapf(err, "unable to get file group information %s: %s", f.Destination)
	}

	if userChanges || groupChanges {
		uid, _ := strconv.Atoi(tgtUser.Uid)
		gid, _ := strconv.Atoi(tgtGroup.Gid)
		err = os.Chown(f.Destination, uid, gid)
		if err != nil {
			return errors.Wrapf(err, "unable to change ownership on %s", f.Destination)
		}
	}

	mode := desiredMode(f, actual)
	if mode != actual.FileMode {
		err = os.Chmod(f.Destination, mode.Perm())
		if err != nil {
			return errors.Wrapf(err, "unable to change permissions on %s", f.Destination)
		}
	}

	// handle links
	switch f.Type {
	case "symlink":
		err := os.Symlink(f.Target, f.Destination)
		if err != nil {
			return errors.Wrapf(err, "unable to create symlink %s -> %s", f.Destination, f.Target)
		}
	case "hardlink":
		same, err := SameFile(f.Destination, f.Target)
		if err != nil {
			return errors.Wrapf(err, "unable to check hardlink status %s -> %s", f.Destination, f.Target)
		}
		if !same {
			err := os.Link(actual.Target, f.Destination)
			if err != nil {
				return errors.Wrapf(err, "unable to create hardlink %s -> %s", f.Destination, f.Target)
			}
		}
	}

	return err
}
