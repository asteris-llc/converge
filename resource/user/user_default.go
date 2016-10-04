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

// +build !linux

package user

import (
	"os/user"
)

// System implements SystemUtils
type System struct{}

// AddUser implementation for systems which are not supported
func (s *System) AddUser(userName string, options *AddUserOptions) error {
	return ErrUnsupported
}

// DelUser implementation for systems which are not supported
func (s *System) DelUser(userName string) error {
	return ErrUnsupported
}

// ModUser implementation for systems which are not supported
func (s *System) ModUser(userName string, options *ModUserOptions) error {
	return ErrUnsupported
}

// Lookup implementation for systems which are not supported
func (s *System) Lookup(userName string) (*user.User, error) {
	return nil, ErrUnsupported
}

// LookupID implementation for systems which are not supported
func (s *System) LookupID(userID string) (*user.User, error) {
	return nil, ErrUnsupported
}

// LookupGroup implementation for systems which are not supported
func (s *System) LookupGroup(groupName string) (*user.Group, error) {
	return nil, ErrUnsupported
}

// LookupGroupID implementation for systems which are not supported
func (s *System) LookupGroupID(groupID string) (*user.Group, error) {
	return nil, ErrUnsupported
}
