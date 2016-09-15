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

package group_test

import (
	"fmt"
	"math"
	"os/user"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/asteris-llc/converge/helpers/fakerenderer"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/group"
	"github.com/fgrid/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	currUser  *user.User
	currGid   string
	currGroup *user.Group
	currName  string
	userErr   error
	groupErr  error
	tempName  []string
	fakeName  string
	fakeGid   string
	gidErr    error
)

const (
	// Valid GID range varies based on system
	// At a minimum, 0-32676 is valid
	GID_MIN = 0
	GID_MAX = math.MaxInt16
)

func init() {
	currUser, userErr = user.Current()
	if userErr != nil {
		panic(userErr)
	}

	currGid = currUser.Gid
	currGroup, groupErr = user.LookupGroupId(currGid)
	if groupErr != nil {
		panic(groupErr)
	}

	fakeGid, gidErr = setFakeGid()
	if gidErr != nil {
		panic(gidErr)
	}

	currName = currGroup.Name
	tempName = strings.Split(uuid.NewV4().String(), "-")
	fakeName = strings.Join(tempName[0:], "")
}

func TestGroupInterface(t *testing.T) {
	t.Parallel()

	assert.Implements(t, (*resource.Task)(nil), new(group.Group))
}

func TestCheckFoundGidFoundNameStatePresent(t *testing.T) {
	t.Parallel()

	g := group.NewGroup(new(group.System))
	g.GID = currGid
	g.Name = currName
	g.State = group.StatePresent
	status, err := g.Check(fakerenderer.New())

	if runtime.GOOS == "linux" {
		assert.NoError(t, err)
		assert.Equal(t, resource.StatusNoChange, status.StatusCode())
		assert.False(t, status.HasChanges())
	} else {
		assert.EqualError(t, err, "group: not supported on this system")
	}
}

func TestCheckFoundGidFoundNameStateAbsent(t *testing.T) {
	t.Parallel()

	g := group.NewGroup(new(group.System))
	g.GID = currGid
	g.Name = currName
	g.State = group.StateAbsent
	status, err := g.Check(fakerenderer.New())

	if runtime.GOOS == "linux" {
		assert.NoError(t, err)
		assert.Equal(t, resource.StatusWillChange, status.StatusCode())
		assert.Equal(t, fmt.Sprintf("group %s with gid %s", g.Name, g.GID), status.Diffs()["group"].Original())
		assert.Equal(t, string(group.StateAbsent), status.Diffs()["group"].Current())
		assert.True(t, status.HasChanges())
	} else {
		assert.EqualError(t, err, "group: not supported on this system")
	}
}

func TestCheckFoundGidNotNameStatePresent(t *testing.T) {
	t.Parallel()

	g := group.NewGroup(new(group.System))
	g.GID = currGid
	g.Name = fakeName
	g.State = group.StatePresent
	status, err := g.Check(fakerenderer.New())

	if runtime.GOOS == "linux" {
		assert.NoError(t, err)
		assert.Equal(t, resource.StatusFatal, status.StatusCode())
		assert.Equal(t, fmt.Sprintf("group gid %s already exists", g.GID), status.Messages()[0])
		assert.False(t, status.HasChanges())
	} else {
		assert.EqualError(t, err, "group: not supported on this system")
	}
}

func TestCheckFoundGidNotNameStateAbsent(t *testing.T) {
	t.Parallel()

	g := group.NewGroup(new(group.System))
	g.GID = currGid
	g.Name = fakeName
	g.State = group.StateAbsent
	status, err := g.Check(fakerenderer.New())

	if runtime.GOOS == "linux" {
		assert.NoError(t, err)
		assert.Equal(t, resource.StatusFatal, status.StatusCode())
		assert.Equal(t, fmt.Sprintf("group %s does not exist", g.Name), status.Messages()[0])
		assert.False(t, status.HasChanges())
	} else {
		assert.EqualError(t, err, "group: not supported on this system")
	}
}

func TestCheckFoundNameNotGidStatePresent(t *testing.T) {
	t.Parallel()

	g := group.NewGroup(new(group.System))
	g.GID = fakeGid
	g.Name = currName
	g.State = group.StatePresent
	status, err := g.Check(fakerenderer.New())

	if runtime.GOOS == "linux" {
		assert.NoError(t, err)
		assert.Equal(t, resource.StatusFatal, status.StatusCode())
		assert.Equal(t, fmt.Sprintf("group %s already exists", g.Name), status.Messages()[0])
		assert.False(t, status.HasChanges())
	} else {
		assert.EqualError(t, err, "group: not supported on this system")
	}
}

func TestCheckFoundNameNotGidStateAbsent(t *testing.T) {
	t.Parallel()

	g := group.NewGroup(new(group.System))
	g.GID = fakeGid
	g.Name = currName
	g.State = group.StateAbsent
	status, err := g.Check(fakerenderer.New())

	if runtime.GOOS == "linux" {
		assert.NoError(t, err)
		assert.Equal(t, resource.StatusFatal, status.StatusCode())
		assert.Equal(t, fmt.Sprintf("group gid %s does not exist", g.GID), status.Messages()[0])
		assert.False(t, status.HasChanges())
	} else {
		assert.EqualError(t, err, "group: not supported on this system")
	}
}

func TestCheckNameAndGidMismatchStatePresent(t *testing.T) {
	t.Parallel()

	gid, err := setGid()
	if err != nil {
		panic(err)
	}
	g := group.NewGroup(new(group.System))
	g.GID = gid
	g.Name = currName
	g.State = group.StatePresent
	status, err := g.Check(fakerenderer.New())

	if runtime.GOOS == "linux" {
		assert.NoError(t, err)
		assert.Equal(t, resource.StatusFatal, status.StatusCode())
		assert.Equal(t, fmt.Sprintf("group %s and gid %s belong to different groups", g.Name, g.GID), status.Messages()[0])
		assert.False(t, status.HasChanges())
	} else {
		assert.EqualError(t, err, "group: not supported on this system")
	}
}

func TestCheckNameAndGidMismatchStateAbsent(t *testing.T) {
	t.Parallel()

	gid, err := setGid()
	if err != nil {
		panic(err)
	}
	g := group.NewGroup(new(group.System))
	g.GID = gid
	g.Name = currName
	g.State = group.StateAbsent
	status, err := g.Check(fakerenderer.New())

	if runtime.GOOS == "linux" {
		assert.NoError(t, err)
		assert.Equal(t, resource.StatusFatal, status.StatusCode())
		assert.Equal(t, fmt.Sprintf("group %s and gid %s belong to different groups", g.Name, g.GID), status.Messages()[0])
		assert.False(t, status.HasChanges())
	} else {
		assert.EqualError(t, err, "group: not supported on this system")
	}
}

func TestCheckNameAndGidNotFoundStatePresent(t *testing.T) {
	t.Parallel()

	g := group.NewGroup(new(group.System))
	g.GID = fakeGid
	g.Name = fakeName
	g.State = group.StatePresent
	status, err := g.Check(fakerenderer.New())

	if runtime.GOOS == "linux" {
		assert.NoError(t, err)
		assert.Equal(t, resource.StatusWillChange, status.StatusCode())
		assert.Equal(t, "group name and gid do not exist", status.Messages()[0])
		assert.Equal(t, string(group.StateAbsent), status.Diffs()["group"].Original())
		assert.Equal(t, fmt.Sprintf("group %s with gid %s", g.Name, g.GID), status.Diffs()["group"].Current())
		assert.True(t, status.HasChanges())
	} else {
		assert.EqualError(t, err, "group: not supported on this system")
	}
}

func TestCheckNameAndGidNotFoundStateAbsent(t *testing.T) {
	t.Parallel()

	g := group.NewGroup(new(group.System))
	g.GID = fakeGid
	g.Name = fakeName
	g.State = group.StateAbsent
	status, err := g.Check(fakerenderer.New())

	if runtime.GOOS == "linux" {
		assert.NoError(t, err)
		assert.Equal(t, resource.StatusNoChange, status.StatusCode())
		assert.Equal(t, "group name and gid do not exist", status.Messages()[0])
		assert.False(t, status.HasChanges())
	} else {
		assert.EqualError(t, err, "group: not supported on this system")
	}
}

func TestCheckStateUnknown(t *testing.T) {
	t.Parallel()

	g := group.NewGroup(new(group.System))
	g.GID = fakeGid
	g.Name = fakeName
	g.State = "test"
	_, err := g.Check(fakerenderer.New())

	if runtime.GOOS == "linux" {
		assert.EqualError(t, err, fmt.Sprintf("group: unrecognized state %s", g.State))
	} else {
		assert.EqualError(t, err, "group: not supported on this system")
	}
}

func TestApplyAddGroup(t *testing.T) {
	t.Parallel()

	grp := &user.Group{
		Name: fakeName,
		Gid:  fakeGid,
	}
	m := &MockSystem{}
	g := group.NewGroup(m)
	g.GID = grp.Gid
	g.Name = grp.Name
	g.State = group.StatePresent

	m.On("LookupGroup", g.Name).Return(grp, user.UnknownGroupError(""))
	m.On("LookupGroupID", g.GID).Return(grp, user.UnknownGroupIdError(""))
	m.On("AddGroup", g.Name, g.GID).Return(nil)
	status, err := g.Apply(fakerenderer.New())

	m.AssertCalled(t, "AddGroup", g.Name, g.GID)
	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("added group %s with gid %s", g.Name, g.GID), status.Messages()[0])
}

func TestApplyDeleteGroup(t *testing.T) {
	t.Parallel()

	grp := &user.Group{
		Name: fakeName,
		Gid:  fakeGid,
	}
	m := &MockSystem{}
	g := group.NewGroup(m)
	g.GID = grp.Gid
	g.Name = grp.Name
	g.State = group.StateAbsent

	m.On("LookupGroup", g.Name).Return(grp, nil)
	m.On("LookupGroupID", g.GID).Return(grp, nil)
	m.On("DelGroup", g.Name).Return(nil)
	status, err := g.Apply(fakerenderer.New())

	m.AssertCalled(t, "DelGroup", g.Name)
	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("deleted group %s with gid %s", g.Name, g.GID), status.Messages()[0])
}

func TestApplyAddGroupErrorAdding(t *testing.T) {
	t.Parallel()

	grp := &user.Group{
		Name: fakeName,
		Gid:  fakeGid,
	}
	m := &MockSystem{}
	g := group.NewGroup(m)
	g.GID = grp.Gid
	g.Name = grp.Name
	g.State = group.StatePresent

	m.On("LookupGroup", g.Name).Return(grp, user.UnknownGroupError(""))
	m.On("LookupGroupID", g.GID).Return(grp, user.UnknownGroupIdError(""))
	m.On("AddGroup", g.Name, g.GID).Return(fmt.Errorf(""))
	status, err := g.Apply(fakerenderer.New())

	m.AssertCalled(t, "AddGroup", g.Name, g.GID)
	assert.EqualError(t, err, fmt.Sprintf(""))
	assert.Equal(t, resource.StatusFatal, status.StatusCode())
	assert.Equal(t, fmt.Sprintf("error adding group %s with gid %s", g.Name, g.GID), status.Messages()[0])
}

func TestApplyAddGroupNotAdded(t *testing.T) {
	t.Parallel()

	grp := &user.Group{
		Name: fakeName,
		Gid:  fakeGid,
	}
	m := &MockSystem{}
	g := group.NewGroup(m)
	g.GID = grp.Gid
	g.Name = grp.Name
	g.State = group.StatePresent

	m.On("LookupGroup", g.Name).Return(grp, nil)
	m.On("LookupGroupID", g.GID).Return(grp, nil)
	m.On("AddGroup", g.Name, g.GID).Return(nil)
	status, err := g.Apply(fakerenderer.New())

	m.AssertNotCalled(t, "AddGroup", g.Name, g.GID)
	assert.EqualError(t, err, fmt.Sprintf("will not attempt add: group %s with gid %s", g.Name, g.GID))
	assert.Equal(t, resource.StatusFatal, status.StatusCode())
}

func TestApplyDeleteGroupErrorDeleting(t *testing.T) {
	t.Parallel()

	grp := &user.Group{
		Name: fakeName,
		Gid:  fakeGid,
	}
	m := &MockSystem{}
	g := group.NewGroup(m)
	g.GID = grp.Gid
	g.Name = grp.Name
	g.State = group.StateAbsent

	m.On("LookupGroup", g.Name).Return(grp, nil)
	m.On("LookupGroupID", g.GID).Return(grp, nil)
	m.On("DelGroup", g.Name).Return(fmt.Errorf(""))
	status, err := g.Apply(fakerenderer.New())

	m.AssertCalled(t, "DelGroup", g.Name)
	assert.EqualError(t, err, fmt.Sprintf(""))
	assert.Equal(t, resource.StatusFatal, status.StatusCode())
	assert.Equal(t, fmt.Sprintf("error deleting group %s with gid %s", g.Name, g.GID), status.Messages()[0])
}

func TestApplyDeleteGroupNotDeleted(t *testing.T) {
	t.Parallel()

	grp1 := &user.Group{
		Name: fakeName,
		Gid:  fakeGid,
	}
	grp2 := &user.Group{
		Name: currName,
		Gid:  currGid,
	}
	m := &MockSystem{}
	g := group.NewGroup(m)
	g.GID = grp2.Gid
	g.Name = grp1.Name
	g.State = group.StateAbsent

	m.On("LookupGroup", g.Name).Return(grp1, nil)
	m.On("LookupGroupID", g.GID).Return(grp2, nil)
	m.On("DelGroup", g.Name).Return(nil)
	status, err := g.Apply(fakerenderer.New())

	m.AssertNotCalled(t, "DelGroup", g.Name)
	assert.EqualError(t, err, fmt.Sprintf("will not attempt delete: group %s with gid %s", g.Name, g.GID))
	assert.Equal(t, resource.StatusFatal, status.StatusCode())
}

func TestApplyStateUnknown(t *testing.T) {
	t.Parallel()

	grp := &user.Group{
		Name: fakeName,
		Gid:  fakeGid,
	}
	m := &MockSystem{}
	g := group.NewGroup(m)
	g.GID = grp.Gid
	g.Name = grp.Name
	g.State = "test"

	m.On("LookupGroup", g.Name).Return(grp, nil)
	m.On("LookupGroupID", g.GID).Return(grp, nil)
	m.On("AddGroup", g.Name, g.GID)
	m.On("DelGroup", g.Name)
	_, err := g.Apply(fakerenderer.New())

	m.AssertNotCalled(t, "AddGroup", g.Name, g.GID)
	m.AssertNotCalled(t, "DelGroup", g.Name)
	assert.EqualError(t, err, fmt.Sprintf("group: unrecognized state %s", g.State))
}

// setGid is used for TestCheckNameAndGidMismatch. We need a gid that exists,
// but is not a match for the current user group name (currName).
func setGid() (string, error) {
	for i := 0; i <= GID_MAX; i++ {
		gid := strconv.Itoa(i)
		group, err := user.LookupGroupId(gid)
		if err == nil && group.Name != currName {
			return gid, nil
		}
	}
	return "", fmt.Errorf("setGid: could not set gid")
}

// setFakeGid is used to set a gid that does not exist.
func setFakeGid() (string, error) {
	for i := GID_MIN; i <= GID_MAX; i++ {
		gid := strconv.Itoa(i)
		_, err := user.LookupGroupId(gid)
		if err != nil {
			return gid, nil
		}
	}
	return "", fmt.Errorf("setFakeGid: could not set gid")
}

type MockSystem struct {
	mock.Mock
}

func (m *MockSystem) AddGroup(name, gid string) error {
	args := m.Called(name, gid)
	return args.Error(0)
}

func (m *MockSystem) DelGroup(name string) error {
	args := m.Called(name)
	return args.Error(0)
}

func (m *MockSystem) LookupGroup(name string) (*user.Group, error) {
	args := m.Called(name)
	return args.Get(0).(*user.Group), args.Error(1)
}

func (m *MockSystem) LookupGroupID(gid string) (*user.Group, error) {
	args := m.Called(gid)
	return args.Get(0).(*user.Group), args.Error(1)
}
