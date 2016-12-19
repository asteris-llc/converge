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
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/stretchr/testify/mock"
)

// any aliases mock.Anything for terseness
var any = mock.Anything

// MockOS mocks OSProxy
type MockOS struct {
	mock.Mock
	walkArgs []*walkArgs
}

// Walk mocks Walk
func (m *MockOS) Walk(root string, walkFunc filepath.WalkFunc) error {
	args := m.Called(root, walkFunc)
	for _, args := range m.walkArgs {
		walkFunc(args.Path, args.Info, args.Err)
	}
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
	Path      string
	User      *user.User
	Group     *user.Group
	FileSize  int64
	FileMode  os.FileMode
	FileIsDir bool
}

func (o ownershipRecord) Name() string {
	return o.Path
}

func (o ownershipRecord) Size() int64 {
	return o.FileSize
}

func (o ownershipRecord) Mode() os.FileMode {
	return o.FileMode
}

func (o ownershipRecord) ModTime() time.Time {
	return time.Now()
}

func (o ownershipRecord) IsDir() bool {
	return o.FileIsDir
}

func (o ownershipRecord) Sys() interface{} {
	uid, _ := strconv.Atoi(o.User.Uid)
	gid, _ := strconv.Atoi(o.User.Gid)
	return &syscall.Stat_t{
		Uid:  uint32(uid),
		Gid:  uint32(gid),
		Size: o.FileSize,
	}
}

func (o *ownershipRecord) ToWalkArgs() *walkArgs {
	return &walkArgs{
		Path: o.Path,
		Err:  nil,
		Info: o,
	}
}

type walkArgs struct {
	Path string
	Info os.FileInfo
	Err  error
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

func makeOwnedFull(path, username, uid, groupname, gid string, isDir bool, size int64, mode os.FileMode) ownershipRecord {
	o := makeOwned(path, username, uid, groupname, gid)
	o.FileSize = size
	o.FileMode = mode
	o.FileIsDir = isDir
	return o
}

func newMockOS(ownedFiles []ownershipRecord,
	users []*user.User,
	groups []*user.Group,
	defaultUser *user.User,
	defaultGroup *user.Group,
) *MockOS {
	m := &MockOS{}
	for _, o := range ownedFiles {
		m.walkArgs = append(m.walkArgs, o.ToWalkArgs())
	}
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

func failingMockOS(failOn map[string]error) *MockOS {
	m := &MockOS{}
	m.On("Walk", any, any).Return(failOn["Walk"])
	m.On("Chown", any, any, any).Return(failOn["Chown"])
	m.On("GetUID", any).Return(0, failOn["GetUID"])
	m.On("GetGID", any).Return(0, failOn["GetGID"])
	if err, ok := failOn["Lookup"]; ok {
		m.On("Lookup", any).Return(nil, err)
	} else {
		m.On("Lookup", any).Return(rootUser, nil)
	}
	if err, ok := failOn["LookupGroup"]; ok {
		m.On("LookupGroup", any).Return(nil, err)
	} else {
		m.On("LookupGroup", any).Return(rootGroup, nil)
	}
	if err, ok := failOn["LookupId"]; ok {
		m.On("LookupId", any).Return(nil, err)
	} else {
		m.On("LookupId", any).Return(rootUser, nil)
	}
	if err, ok := failOn["LookupGroupId"]; ok {
		m.On("LookupGroupId", any).Return(nil, err)
	} else {
		m.On("LookupGroupId", any).Return(rootGroup, nil)
	}
	return m
}

func toInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}
