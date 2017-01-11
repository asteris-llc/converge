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

// +build linux

package user

import (
	"bytes"
	"github.com/pkg/errors"
	"os/exec"
	"os/user"
	"strings"
	"time"
)

// System implements SystemUtils
type System struct{}

// AddUser adds a user
func (s *System) AddUser(userName string, options *AddUserOptions) error {
	args := []string{userName}
	if options.UID != "" {
		args = append(args, "-u", options.UID)
	}
	if options.Group != "" {
		args = append(args, "-g", options.Group)
	}
	if options.Comment != "" {
		args = append(args, "-c", options.Comment)
	}
	if options.CreateHome {
		args = append(args, "-m")
		if options.SkelDir != "" {
			args = append(args, "-k", options.SkelDir)
		}
	}
	if options.Directory != "" {
		args = append(args, "-d", options.Directory)
	}
	if options.Expiry != "" {
		args = append(args, "-e", options.Expiry)
	}

	cmd := exec.Command("useradd", args...)
	err := cmd.Run()
	if err != nil {
		return errors.Wrap(err, "useradd")
	}
	return nil
}

// DelUser deletes a user
func (s *System) DelUser(userName string) error {
	cmd := exec.Command("userdel", userName)
	err := cmd.Run()
	if err != nil {
		return errors.Wrap(err, "userdel")
	}
	return nil
}

// ModUser modifies a user
func (s *System) ModUser(userName string, options *ModUserOptions) error {
	args := []string{userName}
	if options.Username != "" {
		args = append(args, "-l", options.Username)
	}
	if options.UID != "" {
		args = append(args, "-u", options.UID)
	}
	if options.Group != "" {
		args = append(args, "-g", options.Group)
	}
	if options.Comment != "" {
		args = append(args, "-c", options.Comment)
	}
	if options.Directory != "" {
		args = append(args, "-d", options.Directory)
		if options.MoveDir {
			args = append(args, "-m")
		}
	}
	if options.Expiry != "" {
		args = append(args, "-e", options.Expiry)
	}

	cmd := exec.Command("usermod", args...)
	err := cmd.Run()
	if err != nil {
		return errors.Wrap(err, "usermod")
	}
	return nil
}

// LookupUserExpiry looks up a user's expiry
func (s *System) LookupUserExpiry(userName string) (time.Time, error) {
	var out bytes.Buffer

	args := []string{"-l", userName}
	cmd := exec.Command("chage", args...)
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return time.Time{}, errors.Wrap(err, "chage")
	}

	expiry, err := parseForExpiry(out.String())
	if err != nil {
		return time.Time{}, err
	}

	return expiry, nil
}

// Lookup looks up a user by name
// If the user cannot be found an error is returned
func (s *System) Lookup(userName string) (*user.User, error) {
	return user.Lookup(userName)
}

// LookupID looks up a user by uid
// If the user cannot be found an error is returned
func (s *System) LookupID(userID string) (*user.User, error) {
	return user.LookupId(userID)
}

// LookupGroup looks up a group by name
// If the group cannot be found an error is returned
func (s *System) LookupGroup(groupName string) (*user.Group, error) {
	return user.LookupGroup(groupName)
}

// LookupGroupID looks up a group by gid
// If the group cannot be found an error is returned
func (s *System) LookupGroupID(groupID string) (*user.Group, error) {
	return user.LookupGroupId(groupID)
}

// parseForExpiry takes a string and extracts the account expiration date and
// converts it to a time.Time. This function is specifically written to handle
// the output from the `chage -l <username>` command.
func parseForExpiry(data string) (time.Time, error) {
	split := strings.Split(data, "\n")

	for _, line := range split {
		if strings.Contains(line, "Account expires") {
			newsplit := strings.Split(line, ":")
			rawExpiry := strings.Trim(newsplit[1], " ")
			zone := time.FixedZone(time.Now().In(time.Local).Zone())

			if rawExpiry == "never" {
				// set current user time to max time
				return time.ParseInLocation(ShortForm, MaxTime, zone)
			}
			return time.ParseInLocation("Jan 2, 2006", strings.Trim(newsplit[1], " "), zone)
		}
	}

	return time.Time{}, errors.New("could not parse expiry data for current user")
}
