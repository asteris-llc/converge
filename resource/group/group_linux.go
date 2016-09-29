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

package group

import (
	"fmt"
	"os/exec"
	"os/user"
)

// System implements SystemUtils
type System struct{}

// AddGroup adds a group
func (s *System) AddGroup(groupName, groupID string) error {
	args := []string{groupName}
	if groupID != "" {
		args = append(args, "-g", groupID)
	}
	cmd := exec.Command("groupadd", args...)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("groupadd: %s", err)
	}
	return nil
}

// DelGroup deletes a group
func (s *System) DelGroup(groupName string) error {
	cmd := exec.Command("groupdel", groupName)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("groupdel: %s", err)
	}
	return nil
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
