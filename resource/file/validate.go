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
	"fmt"
	"os/user"

	"github.com/pkg/errors"
)

// Validate runs checks against a File resource
func (f *File) Validate() error {
	var err error
	if f.Destination == "" {
		return errors.New("file requires a destination parameter")
	}

	if f.Type == TypeSymlink && f.Mode != nil {
		return fmt.Errorf("permissions on symlinks not supported")
	}

	err = f.validateState()
	if err != nil {
		return err
	}

	err = f.validateType()
	if err != nil {
		return err
	}

	f.validateMode()

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
	switch f.State {
	case StateUndefined:
		f.State = DefaultState
		return nil
	default:
		for _, s := range ValidStates {
			if f.State == s {
				return nil
			}
		}
		return fmt.Errorf("state should be one of %v, got %q", ValidStates, f.State)
	}
}

// Validate the type or set default value
func (f *File) validateType() error {
	switch f.Type {
	case "": //use default if not set
		f.Type = DefaultType
		return nil
	default:
		for _, t := range AllTypes {
			if f.Type == t {
				return nil
			}
		}
		return fmt.Errorf("type should be one of %v, got %q", AllTypes, f.Type)
	}
}

// Validate the mode of the file
func (f *File) validateMode() {
	if f.Mode == nil {
		return
	}
	m := new(uint32)
	*m = ModeType(*f.Mode, f.Type)
	f.Mode = m
}

// A target needs to be set if you are creating a link
func (f *File) validateTarget() error {
	switch f.Target {
	case "":
		if f.Type == TypeSymlink || f.Type == TypeLink {
			return fmt.Errorf("must define a target if you are using a %q", f.Type)
		}
		return nil
	default:
		// is target set for a file or directory type?
		if f.Type == TypeSymlink || f.Type == TypeLink {
			return nil
		}
		return fmt.Errorf("cannot define target on a type of %q: target: %q", f.Type, f.Target)
	}
}

// if a user is provided, ensure it exists on the system
func (f *File) validateUser() error {
	if f.UserInfo == nil {
		f.UserInfo = &user.User{}
		return nil
	}

	if f.UserInfo.Username != "" {
		u, err := user.Lookup(f.UserInfo.Username)
		if err != nil {
			return errors.Wrap(err, "validation error")
		}
		f.UserInfo = u
		return err
	}
	return nil
}

//if a group is provided, make sure it exists on the system
func (f *File) validateGroup() error {
	if f.GroupInfo == nil {
		f.GroupInfo = &user.Group{}
		return nil
	}

	if f.GroupInfo.Name != "" {
		g, err := user.LookupGroup(f.GroupInfo.Name)
		if err != nil {
			return errors.Wrap(err, "validation error")
		}
		f.GroupInfo = g
	}
	return nil
}
