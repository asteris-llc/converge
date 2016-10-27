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
	"os/user"

	"github.com/asteris-llc/converge/resource"
	"github.com/pkg/errors"
)

// State type for User
type State string

const (
	// StatePresent indicates the user should be present
	StatePresent State = "present"

	// StateAbsent indicates the user should be absent
	StateAbsent State = "absent"
)

// User manages user users
type User struct {
	Username    string
	NewUsername string
	UID         string
	GroupName   string
	GID         string
	Name        string
	CreateHome  bool
	SkelDir     string
	HomeDir     string
	MoveDir     bool
	State       State
	system      SystemUtils

	*resource.Status
}

// AddUserOptions are the options specified in the configuration to be used
// when adding a user
type AddUserOptions struct {
	UID        string
	Group      string
	Comment    string
	CreateHome bool
	SkelDir    string
	Directory  string
}

// ModUserOptions are the options specified in the configuration to be used
// when modifying a user
type ModUserOptions struct {
	Username  string
	UID       string
	Group     string
	Comment   string
	Directory string
	MoveDir   bool
}

// SystemUtils provides system utilities for user
type SystemUtils interface {
	AddUser(userName string, options *AddUserOptions) error
	DelUser(userName string) error
	ModUser(userName string, options *ModUserOptions) error
	Lookup(userName string) (*user.User, error)
	LookupID(userID string) (*user.User, error)
	LookupGroup(groupName string) (*user.Group, error)
	LookupGroupID(groupID string) (*user.Group, error)
}

// ErrUnsupported is used when a system is not supported
var ErrUnsupported = fmt.Errorf("user: not supported on this system")

// NewUser constructs and returns a new User
func NewUser(system SystemUtils) *User {
	return &User{
		system: system,
	}
}

// Check if a user user exists
func (u *User) Check(resource.Renderer) (resource.TaskStatus, error) {
	var (
		userByID *user.User
		uidErr   error
	)

	// lookup the user by name and lookup the user by uid
	// the lookups return ErrUnsupported if the system is not supported
	// Lookup returns user.UnknownUserError if the user is not found
	// LookupID returns user.UnknownUserIdError if the uid is not found
	userByName, nameErr := u.system.Lookup(u.Username)
	if u.UID != "" {
		userByID, uidErr = u.system.LookupID(u.UID)
	}

	u.Status = resource.NewStatus()

	if nameErr == ErrUnsupported {
		u.Status.RaiseLevel(resource.StatusFatal)
		return u, ErrUnsupported
	}

	switch u.State {
	case StatePresent:
		_, nameNotFound := nameErr.(user.UnknownUserError)

		switch {
		case nameNotFound:
			_, err := u.DiffAdd(u.Status)
			if err != nil {
				return u, errors.Wrapf(err, "cannot add user %s", u.Username)
			}
			if resource.AnyChanges(u.Status.Differences) {
				u.Status.AddMessage("add user")
			}
		case userByName != nil:
			_, err := u.DiffMod(u.Status, userByName)
			if err != nil {
				return u, errors.Wrapf(err, "cannot modify user %s", u.Username)
			}
			if resource.AnyChanges(u.Status.Differences) {
				u.Status.AddMessage("modify user")
			}
		}
	case StateAbsent:
		switch {
		case u.UID == "":
			_, nameNotFound := nameErr.(user.UnknownUserError)

			switch {
			case nameNotFound:
				u.Status.AddMessage(fmt.Sprintf("user %s does not exist", u.Username))
			case userByName != nil:
				u.Status.RaiseLevel(resource.StatusWillChange)
				u.Status.AddDifference("user", fmt.Sprintf("user %s", u.Username), fmt.Sprintf("<%s>", string(StateAbsent)), "")
			}
		case u.UID != "":
			_, nameNotFound := nameErr.(user.UnknownUserError)
			_, uidNotFound := uidErr.(user.UnknownUserIdError)

			switch {
			case nameNotFound && uidNotFound:
				u.Status.AddMessage(fmt.Sprintf("user %s and uid %s do not exist", u.Username, u.UID))
			case nameNotFound:
				u.Status.RaiseLevel(resource.StatusCantChange)
				return u, fmt.Errorf("cannot delete user %s with uid %s: user does not exist", u.Username, u.UID)
			case uidNotFound:
				u.Status.RaiseLevel(resource.StatusCantChange)
				return u, fmt.Errorf("cannot delete user %s with uid %s: uid does not exist", u.Username, u.UID)
			case userByName != nil && userByID != nil && userByName.Name != userByID.Name || userByName.Uid != userByID.Uid:
				u.Status.RaiseLevel(resource.StatusCantChange)
				return u, fmt.Errorf("cannot delete user %s with uid %s: user and uid belong to different users", u.Username, u.UID)
			case userByName != nil && userByID != nil && *userByName == *userByID:
				u.Status.RaiseLevel(resource.StatusWillChange)
				u.Status.AddDifference("user", fmt.Sprintf("user %s with uid %s", u.Username, u.UID), fmt.Sprintf("<%s>", string(StateAbsent)), "")
			}
		}
	default:
		u.Status.RaiseLevel(resource.StatusFatal)
		return u, fmt.Errorf("user: unrecognized state %v", u.State)
	}

	return u.Status, nil
}

// Apply changes for user
func (u *User) Apply() (resource.TaskStatus, error) {
	var (
		userByID *user.User
		uidErr   error
	)

	// lookup the user by name and lookup the user by uid
	// the lookups return ErrUnsupported if the system is not supported
	// Lookup returns user.UnknownUserError if the user is not found
	// LookupID returns user.UnknownUserIdError if the uid is not found
	userByName, nameErr := u.system.Lookup(u.Username)
	if u.UID != "" {
		userByID, uidErr = u.system.LookupID(u.UID)
	}

	u.Status = resource.NewStatus()

	if nameErr == ErrUnsupported {
		u.Status.RaiseLevel(resource.StatusFatal)
		return u, ErrUnsupported
	}

	switch u.State {
	case StatePresent:
		_, nameNotFound := nameErr.(user.UnknownUserError)

		switch {
		case nameNotFound:
			options, err := u.DiffAdd(u.Status)
			if err != nil {
				return u, errors.Wrapf(err, "will not attempt to add user %s", u.Username)
			}
			if resource.AnyChanges(u.Status.Differences) {
				err = u.system.AddUser(u.Username, options)
				if err != nil {
					u.Status.RaiseLevel(resource.StatusFatal)
					u.Status.AddMessage(fmt.Sprintf("error adding user %s", u.Username))
					return u, errors.Wrap(err, "user add")
				}
				u.Status.AddMessage(fmt.Sprintf("added user %s", u.Username))
			}
		case userByName != nil:
			options, err := u.DiffMod(u.Status, userByName)
			if err != nil {
				return u, errors.Wrapf(err, "will not attempt to modify user %s", u.Username)
			}
			if resource.AnyChanges(u.Status.Differences) {
				err = u.system.ModUser(u.Username, options)
				if err != nil {
					u.Status.RaiseLevel(resource.StatusFatal)
					u.Status.AddMessage(fmt.Sprintf("error modifying user %s", u.Username))
					return u, errors.Wrap(err, "user modify")
				}
				u.Status.AddMessage(fmt.Sprintf("modified user %s", u.Username))
			}
		}
	case StateAbsent:
		switch {
		case u.UID == "":
			_, nameNotFound := nameErr.(user.UnknownUserError)

			switch {
			case !nameNotFound && userByName != nil:
				err := u.system.DelUser(u.Username)
				if err != nil {
					u.Status.RaiseLevel(resource.StatusFatal)
					u.Status.AddMessage(fmt.Sprintf("error deleting user %s", u.Username))
					return u, errors.Wrap(err, "user delete")
				}
				u.Status.AddMessage(fmt.Sprintf("deleted user %s", u.Username))
			default:
				u.Status.RaiseLevel(resource.StatusCantChange)
				return u, fmt.Errorf("will not attempt to delete user %s", u.Username)
			}
		case u.UID != "":
			_, nameNotFound := nameErr.(user.UnknownUserError)
			_, uidNotFound := uidErr.(user.UnknownUserIdError)

			switch {
			case !nameNotFound && !uidNotFound && userByName != nil && userByID != nil && *userByName == *userByID:
				err := u.system.DelUser(u.Username)
				if err != nil {
					u.Status.RaiseLevel(resource.StatusFatal)
					u.Status.AddMessage(fmt.Sprintf("error deleting user %s with uid %s", u.Username, u.UID))
					return u, errors.Wrap(err, "user delete")
				}
				u.Status.AddMessage(fmt.Sprintf("deleted user %s with uid %s", u.Username, u.UID))
			default:
				u.Status.RaiseLevel(resource.StatusCantChange)
				return u, fmt.Errorf("will not attempt to delete user %s with uid %s", u.Username, u.UID)
			}
		}
	default:
		u.Status.RaiseLevel(resource.StatusFatal)
		return u, fmt.Errorf("user: unrecognized state %s", u.State)
	}

	return u, nil
}

func noOptionsSet(u *User) bool {
	switch {
	case u.UID != "", u.GroupName != "", u.GID != "", u.Name != "", u.HomeDir != "":
		return false
	}
	return true
}

// DiffAdd checks for differences between the current and desired state for the
// user to be added indicated by the User fields. The options to be used for the
// add command are set.
func (u *User) DiffAdd(status *resource.Status) (*AddUserOptions, error) {
	options := new(AddUserOptions)

	// if a group exists with the same name as the user being added, a groupname
	// must also be indicated so the user may be added to that group
	grp, _ := user.LookupGroup(u.Username)
	if grp != nil && grp.Name == u.Username && u.GroupName == "" {
		u.Status.RaiseLevel(resource.StatusCantChange)
		u.Status.AddMessage("if you want to add this user to that group, use the groupname field")
		return nil, fmt.Errorf("group %s exists", u.Username)
	}
	u.Status.AddDifference("username", fmt.Sprintf("<%s>", string(StateAbsent)), u.Username, "")

	if u.UID != "" {
		usr, err := user.LookupId(u.UID)
		_, uidNotFound := err.(user.UnknownUserIdError)

		if uidNotFound {
			options.UID = u.UID
			status.AddDifference("uid", fmt.Sprintf("<%s>", string(StateAbsent)), u.UID, "")
		} else if usr != nil {
			u.Status.RaiseLevel(resource.StatusCantChange)
			return nil, fmt.Errorf("uid %s already exists", u.UID)
		}
	}

	switch {
	case u.GroupName != "":
		grp, err := user.LookupGroup(u.GroupName)
		if err != nil {
			u.Status.RaiseLevel(resource.StatusCantChange)
			return nil, fmt.Errorf("group %s does not exist", u.GroupName)
		} else if grp != nil {
			options.Group = u.GroupName
			status.AddDifference("group", fmt.Sprintf("<%s>", string(StateAbsent)), u.GroupName, "")
		}
	case u.GID != "":
		grp, err := user.LookupGroupId(u.GID)
		if err != nil {
			u.Status.RaiseLevel(resource.StatusCantChange)
			return nil, fmt.Errorf("group gid %s does not exist", u.GID)
		} else if grp != nil {
			options.Group = u.GID
			status.AddDifference("gid", fmt.Sprintf("<%s>", string(StateAbsent)), u.GID, "")
		}
	}

	if u.Name != "" {
		options.Comment = u.Name
		status.AddDifference("comment", fmt.Sprintf("<%s>", string(StateAbsent)), u.Name, "")
	}

	if u.CreateHome {
		dirDiff := u.HomeDir
		if u.HomeDir == "" {
			dirDiff = "<default home>"
		}
		options.CreateHome = true
		status.AddDifference("create_home", fmt.Sprintf("<%s>", string(StateAbsent)), dirDiff, "")
		if u.SkelDir != "" {
			options.SkelDir = u.SkelDir
			status.AddDifference("skel_dir contents", u.SkelDir, dirDiff, "")
		}
	}

	if u.HomeDir != "" {
		options.Directory = u.HomeDir
		status.AddDifference("home_dir name", "<default home>", u.HomeDir, "")
	}

	if resource.AnyChanges(u.Status.Differences) {
		u.Status.RaiseLevel(resource.StatusWillChange)
	}

	return options, nil
}

// DiffMod checks for differences between the user associated with u.Username
// and the desired modifications of that user indicated by the other User
// fields. The options to be used for the modify command are set.
func (u *User) DiffMod(status *resource.Status, currUser *user.User) (*ModUserOptions, error) {
	options := new(ModUserOptions)

	// Check for differences between currUser and the desired modifications
	if u.NewUsername != "" {
		usr, _ := user.Lookup(u.NewUsername)
		if usr != nil {
			u.Status.RaiseLevel(resource.StatusCantChange)
			return nil, fmt.Errorf("user %s already exists", u.NewUsername)
		}
		options.Username = u.NewUsername
		status.AddDifference("username", u.Username, u.NewUsername, "")
	}

	if u.UID != "" {
		usr, err := user.LookupId(u.UID)
		_, uidNotFound := err.(user.UnknownUserIdError)

		if uidNotFound {
			options.UID = u.UID
			status.AddDifference("uid", currUser.Uid, u.UID, "")
		} else if usr != nil && currUser.Uid != u.UID {
			u.Status.RaiseLevel(resource.StatusCantChange)
			return nil, fmt.Errorf("uid %s already exists", u.UID)
		}
	}

	switch {
	case u.GroupName != "":
		grp, err := user.LookupGroup(u.GroupName)
		if err != nil {
			u.Status.RaiseLevel(resource.StatusCantChange)
			return nil, fmt.Errorf("group %s does not exist", u.GroupName)
		} else if grp != nil && currUser.Gid != grp.Gid {
			currGroup, err := user.LookupGroupId(currUser.Gid)
			if err != nil {
				u.Status.RaiseLevel(resource.StatusCantChange)
				return nil, fmt.Errorf("group gid %s does not exist", currUser.Gid)
			}
			options.Group = u.GroupName
			status.AddDifference("group", currGroup.Name, u.GroupName, "")
		}
	case u.GID != "":
		grp, err := user.LookupGroupId(u.GID)
		if err != nil {
			u.Status.RaiseLevel(resource.StatusCantChange)
			return nil, fmt.Errorf("group gid %s does not exist", u.GID)
		} else if grp != nil && currUser.Gid != u.GID {
			options.Group = u.GID
			status.AddDifference("gid", currUser.Gid, u.GID, "")
		}
	}

	if u.Name != "" {
		if currUser.Name != u.Name {
			options.Comment = u.Name
			status.AddDifference("comment", currUser.Name, u.Name, "")
		}
	}

	if u.HomeDir != "" {
		if currUser.HomeDir != u.HomeDir {
			options.Directory = u.HomeDir
			status.AddDifference("home_dir name", currUser.HomeDir, u.HomeDir, "")
			if u.MoveDir {
				options.MoveDir = true
				status.AddDifference("home_dir contents", currUser.HomeDir, u.HomeDir, "")
			}
		}
	}

	if resource.AnyChanges(u.Status.Differences) {
		u.Status.RaiseLevel(resource.StatusWillChange)
	}

	return options, nil
}
