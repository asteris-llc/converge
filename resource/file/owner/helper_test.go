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

package owner_test

import (
	"os/user"
	"path/filepath"
	"strconv"

	"github.com/stretchr/testify/mock"
)

var any = mock.Anything

// MockOS mocks OSProxy
type MockOS struct {
	mock.Mock
}

// Walk mocks Walk
func (m *MockOS) Walk(root string, walkFunc filepath.WalkFunc) error {
	args := m.Called(root, walkFunc)
	return args.Error(0)
}

// Chown mocks Chown
func (m *MockOS) Chown(path string, uid, gid int) error {
	args := m.Called(path, uid, gid)
	return args.Error(0)
}

// GetUID mocks GetUID
func (m *MockOS) GetUID(path string) (int, error) {
	args := m.Called(path)
	return args.Int(0), args.Error(1)
}

// GetGID mocks GetGID
func (m *MockOS) GetGID(path string) (int, error) {
	args := m.Called(path)
	return args.Int(0), args.Error(1)
}

// LookupGroupId mocks LookupGroupId
func (m *MockOS) LookupGroupId(path string) (*user.Group, error) {
	args := m.Called(path)
	return args.Get(0).(*user.Group), args.Error(1)
}

// LookupGroup mocks LookupGroup
func (m *MockOS) LookupGroup(path string) (*user.Group, error) {
	args := m.Called(path)
	return args.Get(0).(*user.Group), args.Error(1)
}

// LookupId mocks LookupId
func (m *MockOS) LookupId(path string) (*user.User, error) {
	args := m.Called(path)
	return args.Get(0).(*user.User), args.Error(1)
}

// Lookup mocks Lookup
func (m *MockOS) Lookup(path string) (*user.User, error) {
	args := m.Called(path)
	return args.Get(0).(*user.User), args.Error(1)
}

func fakeUser(uid, gid, name string) *user.User {
	return &user.User{
		Uid:      uid,
		Gid:      gid,
		Username: name,
		Name:     name,
		HomeDir:  "~" + name,
	}
}

func fakeGroup(gid, name string) *user.Group {
	return &user.Group{Gid: gid, Name: name}
}

type ownershipRecord struct {
	Path  string
	User  *user.User
	Group *user.Group
}

var (
	rootUser = &user.User{
		Uid:      "0",
		Gid:      "0",
		Username: "root",
		Name:     "root",
		HomeDir:  "/root",
	}
	rootGroup = &user.Group{
		Gid:  "0",
		Name: "root",
	}
)

func makeOwned(path, username, uid, groupname, gid string) ownershipRecord {
	return ownershipRecord{
		Path: path,
		User: &user.User{
			Uid:      uid,
			Gid:      gid,
			Username: username,
			Name:     username,
			HomeDir:  "~" + username,
		},
		Group: &user.Group{
			Gid:  gid,
			Name: groupname,
		},
	}
}

func newMockOS(ownedFiles []ownershipRecord,
	users []*user.User,
	groups []*user.Group,
	defaultUser *user.User,
	defaultGroup *user.Group,
) *MockOS {
	m := &MockOS{}
	if defaultUser == nil {
		defaultUser = rootUser
	}
	if defaultGroup == nil {
		defaultGroup = rootGroup
	}
	m.On("Walk", any, any).Return(nil)
	m.On("Chown", any, any, any).Return(nil)
	for _, rec := range ownedFiles {
		m.On("GetUID", rec.Path).Return(toInt(rec.User.Uid), nil)
		m.On("GetGID", rec.Path).Return(toInt(rec.User.Gid), nil)
	}
	m.On("GetUID", any).Return(0, nil)
	m.On("GetGID", any).Return(0, nil)
	for _, user := range users {
		m.On("LookupId", user.Uid).Return(user, nil)
		m.On("Lookup", user.Username).Return(user, nil)
	}
	for _, group := range groups {
		m.On("LookupGroupId", group.Gid).Return(group, nil)
		m.On("LookupGroup", group.Name).Return(group, nil)
	}
	m.On("Lookup", any).Return(defaultUser, nil)
	m.On("LookupId", any).Return(defaultUser, nil)
	m.On("LookupGroup", any).Return(defaultGroup, nil)
	m.On("LookupGroupId", any).Return(defaultGroup, nil)
	return m
}

func toInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}
