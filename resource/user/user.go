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
	Username  string
	UID       string
	GroupName string
	GID       string
	Name      string
	HomeDir   string
	State     State
	system    SystemUtils
}

// SystemUtils provides system utilities for user
type SystemUtils interface {
	AddUser(userName string, options map[string]string) error
	DelUser(userName string) error
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

	status := &resource.Status{}

	if nameErr == ErrUnsupported {
		status.Level = resource.StatusFatal
		return status, ErrUnsupported
	}

	switch u.State {
	case StatePresent:
		switch {
		case u.UID == "":
			_, nameNotFound := nameErr.(user.UnknownUserError)

			switch {
			case userByName != nil:
				status.Level = resource.StatusFatal
				status.Output = append(status.Output, fmt.Sprintf("user %s already exists", u.Username))
			case nameNotFound:
				switch {
				case u.GroupName != "":
					_, err := u.system.LookupGroup(u.GroupName)
					if err != nil {
						status.Level = resource.StatusFatal
						status.Output = append(status.Output, fmt.Sprintf("group %s does not exist", u.GroupName))
						return status, fmt.Errorf("will not add user %s", u.Username)
					}
				case u.GID != "":
					_, err := u.system.LookupGroupID(u.GID)
					if err != nil {
						status.Level = resource.StatusFatal
						status.Output = append(status.Output, fmt.Sprintf("group gid %s does not exist", u.GID))
						return status, fmt.Errorf("will not add user %s", u.Username)
					}
				}
				status.Level = resource.StatusWillChange
				status.Output = append(status.Output, "user does not exist")
				status.AddDifference("user", string(StateAbsent), fmt.Sprintf("user %s", u.Username), "")
			}
		case u.UID != "":
			_, nameNotFound := nameErr.(user.UnknownUserError)
			_, uidNotFound := uidErr.(user.UnknownUserIdError)

			switch {
			case nameNotFound && uidNotFound:
				switch {
				case u.GroupName != "":
					_, err := u.system.LookupGroup(u.GroupName)
					if err != nil {
						status.Level = resource.StatusFatal
						status.Output = append(status.Output, fmt.Sprintf("group %s does not exist", u.GroupName))
						return status, fmt.Errorf("will not add user %s with uid %s", u.Username, u.UID)
					}
				case u.GID != "":
					_, err := u.system.LookupGroupID(u.GID)
					if err != nil {
						status.Level = resource.StatusFatal
						status.Output = append(status.Output, fmt.Sprintf("group gid %s does not exist", u.GID))
						return status, fmt.Errorf("will not add user %s with uid %s", u.Username, u.UID)
					}
				}
				status.Level = resource.StatusWillChange
				status.Output = append(status.Output, "user name and uid do not exist")
				status.AddDifference("user", string(StateAbsent), fmt.Sprintf("user %s with uid %s", u.Username, u.UID), "")
			case nameNotFound:
				status.Level = resource.StatusFatal
				status.Output = append(status.Output, fmt.Sprintf("user uid %s already exists", u.UID))
			case uidNotFound:
				status.Level = resource.StatusFatal
				status.Output = append(status.Output, fmt.Sprintf("user %s already exists", u.Username))
			case userByName != nil && userByID != nil && userByName.Name != userByID.Name || userByName.Uid != userByID.Uid:
				status.Level = resource.StatusFatal
				status.Output = append(status.Output, fmt.Sprintf("user %s and uid %s belong to different users", u.Username, u.UID))
			case userByName != nil && userByID != nil && *userByName == *userByID:
				status.Output = append(status.Output, fmt.Sprintf("user %s with uid %s already exists", u.Username, u.UID))
			}
		}
	case StateAbsent:
		switch {
		case u.UID == "":
			_, nameNotFound := nameErr.(user.UnknownUserError)

			switch {
			case nameNotFound:
				status.Level = resource.StatusFatal
				status.Output = append(status.Output, fmt.Sprintf("user %s does not exist", u.Username))
			case userByName != nil:
				status.Level = resource.StatusWillChange
				status.AddDifference("user", fmt.Sprintf("user %s", u.Username), string(StateAbsent), "")
			}
		case u.UID != "":
			_, nameNotFound := nameErr.(user.UnknownUserError)
			_, uidNotFound := uidErr.(user.UnknownUserIdError)

			switch {
			case nameNotFound && uidNotFound:
				status.Output = append(status.Output, "user name and uid do not exist")
			case nameNotFound:
				status.Level = resource.StatusFatal
				status.Output = append(status.Output, fmt.Sprintf("user %s does not exist", u.Username))
			case uidNotFound:
				status.Level = resource.StatusFatal
				status.Output = append(status.Output, fmt.Sprintf("user uid %s does not exist", u.UID))
			case userByName != nil && userByID != nil && userByName.Name != userByID.Name || userByName.Uid != userByID.Uid:
				status.Level = resource.StatusFatal
				status.Output = append(status.Output, fmt.Sprintf("user %s and uid %s belong to different users", u.Username, u.UID))
			case userByName != nil && userByID != nil && *userByName == *userByID:
				status.Level = resource.StatusWillChange
				status.AddDifference("user", fmt.Sprintf("user %s with uid %s", u.Username, u.UID), string(StateAbsent), "")
			}
		}
	default:
		status.Level = resource.StatusFatal
		return status, fmt.Errorf("user: unrecognized state %v", u.State)
	}

	return status, nil
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

	status := &resource.Status{}

	if nameErr == ErrUnsupported {
		status.Level = resource.StatusFatal
		return status, ErrUnsupported
	}

	switch u.State {
	case StatePresent:
		switch {
		case u.UID == "":
			_, nameNotFound := nameErr.(user.UnknownUserError)

			switch {
			case nameNotFound:
				userAddOptions := SetUserAddOptions(u)
				err := u.system.AddUser(u.Username, userAddOptions)
				if err != nil {
					status.Level = resource.StatusFatal
					status.Output = append(status.Output, fmt.Sprintf("error adding user %s", u.Username))
					return status, err
				}
				status.Output = append(status.Output, fmt.Sprintf("added user %s", u.Username))
			default:
				status.Level = resource.StatusFatal
				return status, fmt.Errorf("will not attempt add: user %s", u.Username)
			}
		case u.UID != "":
			_, nameNotFound := nameErr.(user.UnknownUserError)
			_, uidNotFound := uidErr.(user.UnknownUserIdError)

			switch {
			case nameNotFound && uidNotFound:
				userAddOptions := SetUserAddOptions(u)
				err := u.system.AddUser(u.Username, userAddOptions)
				if err != nil {
					status.Level = resource.StatusFatal
					status.Output = append(status.Output, fmt.Sprintf("error adding user %s with uid %s", u.Username, u.UID))
					return status, err
				}
				status.Output = append(status.Output, fmt.Sprintf("added user %s with uid %s", u.Username, u.UID))
			default:
				status.Level = resource.StatusFatal
				return status, fmt.Errorf("will not attempt add: user %s with uid %s", u.Username, u.UID)
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
					status.Level = resource.StatusFatal
					status.Output = append(status.Output, fmt.Sprintf("error deleting user %s", u.Username))
					return status, err
				}
				status.Output = append(status.Output, fmt.Sprintf("deleted user %s", u.Username))
			default:
				status.Level = resource.StatusFatal
				return status, fmt.Errorf("will not attempt delete: user %s", u.Username)
			}
		case u.UID != "":
			_, nameNotFound := nameErr.(user.UnknownUserError)
			_, uidNotFound := uidErr.(user.UnknownUserIdError)

			switch {
			case !nameNotFound && !uidNotFound && userByName != nil && userByID != nil && *userByName == *userByID:
				err := u.system.DelUser(u.Username)
				if err != nil {
					status.Level = resource.StatusFatal
					status.Output = append(status.Output, fmt.Sprintf("error deleting user %s with uid %s", u.Username, u.UID))
					return status, err
				}
				status.Output = append(status.Output, fmt.Sprintf("deleted user %s with uid %s", u.Username, u.UID))
			default:
				status.Level = resource.StatusFatal
				return status, fmt.Errorf("will not attempt delete: user %s with uid %s", u.Username, u.UID)
			}
		}
	default:
		status.Level = resource.StatusFatal
		return status, fmt.Errorf("user: unrecognized state %s", u.State)
	}

	return status, nil
}

// SetUserAddOptions populates a map with options specified
// in the configuration to use in the userAdd command
func SetUserAddOptions(u *User) map[string]string {
	var userAddOptions = map[string]string{}

	if u.UID != "" {
		userAddOptions["uid"] = u.UID
	}

	switch {
	case u.GroupName != "":
		userAddOptions["group"] = u.GroupName
	case u.GID != "":
		userAddOptions["group"] = u.GID
	}

	if u.Name != "" {
		userAddOptions["comment"] = u.Name
	}

	if u.HomeDir != "" {
		userAddOptions["directory"] = u.HomeDir
	}

	return userAddOptions
}
