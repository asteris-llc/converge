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

package user_test

import (
	"fmt"
	"math"
	os "os/user"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/asteris-llc/converge/helpers/fakerenderer"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/user"
	"github.com/fgrid/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGroupInterface(t *testing.T) {
	t.Parallel()

	assert.Implements(t, (*resource.Task)(nil), new(user.User))
}

var (
	currUser      *os.User
	currUsername  string
	currUID       string
	currGroup     *os.Group
	currGroupName string
	currGid       string
	userErr       error
	groupErr      error
	tempUsername  []string
	fakeUsername  string
	fakeUID       string
	tempGroupName []string
	fakeGroupName string
	fakeGid       string
	gidErr        error
	uidErr        error
)

const (
	// GIDMin designates the smallest valid GID
	// At a minimum, 0-32676 is valid
	GIDMin = 0

	// GIDMax designates the largest valid GID
	// At a minimum, 0-32676 is valid
	GIDMax = math.MaxInt16

	// UIDMin designates the smallest valid UID
	// At a minimum, 0-32676 is valid
	UIDMin = 0

	// UIDMax designates the largest valid UID
	// At a minimum, 0-32676 is valid
	UIDMax = math.MaxInt16
)

func init() {
	currUser, userErr = os.Current()
	if userErr != nil {
		panic(userErr)
	}

	currUsername = currUser.Username
	currUID = currUser.Uid

	currGid = currUser.Gid
	currGroup, groupErr = os.LookupGroupId(currGid)
	if groupErr != nil {
		panic(groupErr)
	}

	fakeUID, uidErr = setFakeUid()
	if uidErr != nil {
		panic(uidErr)
	}
	fakeGid, gidErr = setFakeGid()
	if gidErr != nil {
		panic(gidErr)
	}

	currUsername = currUser.Username
	tempUsername = strings.Split(uuid.NewV4().String(), "-")
	fakeUsername = strings.Join(tempUsername[0:], "")
	tempGroupName = strings.Split(uuid.NewV4().String(), "-")
	fakeGroupName = strings.Join(tempUsername[0:], "")
}

func TestCheckPresentNoUidUserExists(t *testing.T) {
	t.Parallel()

	u := user.NewUser(new(user.System))
	u.Username = currUsername
	u.State = user.StatePresent
	status, err := u.Check(fakerenderer.New())

	if runtime.GOOS == "linux" {
		assert.NoError(t, err)
		assert.Equal(t, resource.StatusFatal, status.StatusCode())
		assert.Equal(t, fmt.Sprintf("user %s already exists", u.Username), status.Messages()[0])
		assert.False(t, status.HasChanges())
	} else {
		assert.EqualError(t, err, "user: not supported on this system")
	}
}

func TestCheckPresentNoUidWithGidFails(t *testing.T) {
	t.Parallel()

	gid, err := setFakeGid()
	if err != nil {
		panic(err)
	}
	u := user.NewUser(new(user.System))
	u.Username = fakeUsername
	u.GID = gid
	u.State = user.StatePresent
	status, err := u.Check(fakerenderer.New())

	if runtime.GOOS == "linux" {
		assert.EqualError(t, err, fmt.Sprintf("will not add user %s", u.Username))
		assert.Equal(t, resource.StatusFatal, status.StatusCode())
		assert.Equal(t, fmt.Sprintf("group gid %s does not exist", u.GID), status.Messages()[0])
		assert.False(t, status.HasChanges())
	} else {
		assert.EqualError(t, err, "user: not supported on this system")
	}
}

func TestCheckPresentNoUidWithGidAddUser(t *testing.T) {
	t.Parallel()

	u := user.NewUser(new(user.System))
	u.Username = fakeUsername
	u.GID = currGid
	u.State = user.StatePresent
	status, err := u.Check(fakerenderer.New())

	if runtime.GOOS == "linux" {
		assert.NoError(t, err)
		assert.Equal(t, "user does not exist", status.Messages()[0])
		assert.Equal(t, resource.StatusWillChange, status.StatusCode())
		assert.Equal(t, string(user.StateAbsent), status.Diffs()["user"].Original())
		assert.Equal(t, fmt.Sprintf("user %s", u.Username), status.Diffs()["user"].Current())
		assert.True(t, status.HasChanges())
	} else {
		assert.EqualError(t, err, "user: not supported on this system")
	}
}

func TestCheckPresentNoUidNoGidAddUser(t *testing.T) {
	t.Parallel()

	u := user.NewUser(new(user.System))
	u.Username = fakeUsername
	u.State = user.StatePresent
	status, err := u.Check(fakerenderer.New())

	if runtime.GOOS == "linux" {
		assert.NoError(t, err)
		assert.Equal(t, "user does not exist", status.Messages()[0])
		assert.Equal(t, resource.StatusWillChange, status.StatusCode())
		assert.Equal(t, string(user.StateAbsent), status.Diffs()["user"].Original())
		assert.Equal(t, fmt.Sprintf("user %s", u.Username), status.Diffs()["user"].Current())
		assert.True(t, status.HasChanges())
	} else {
		assert.EqualError(t, err, "user: not supported on this system")
	}
}

func TestCheckPresentWithUidWithGidFails(t *testing.T) {
	t.Parallel()

	gid, err := setFakeGid()
	if err != nil {
		panic(err)
	}
	u := user.NewUser(new(user.System))
	u.Username = fakeUsername
	u.UID = fakeUID
	u.GID = gid
	u.State = user.StatePresent
	status, err := u.Check(fakerenderer.New())

	if runtime.GOOS == "linux" {
		assert.EqualError(t, err, fmt.Sprintf("will not add user %s with uid %s", u.Username, u.UID))
		assert.Equal(t, resource.StatusFatal, status.StatusCode())
		assert.Equal(t, fmt.Sprintf("group gid %s does not exist", u.GID), status.Messages()[0])
		assert.False(t, status.HasChanges())
	} else {
		assert.EqualError(t, err, "user: not supported on this system")
	}
}

func TestCheckPresentWithUidWithGidAddUser(t *testing.T) {
	t.Parallel()

	gid, err := setGid()
	if err != nil {
		panic(err)
	}
	u := user.NewUser(new(user.System))
	u.Username = fakeUsername
	u.UID = fakeUID
	u.GID = gid
	u.State = user.StatePresent
	status, err := u.Check(fakerenderer.New())

	if runtime.GOOS == "linux" {
		assert.NoError(t, err)
		assert.Equal(t, resource.StatusWillChange, status.StatusCode())
		assert.Equal(t, fmt.Sprintf("user name and uid do not exist"), status.Messages()[0])
		assert.Equal(t, string(user.StateAbsent), status.Diffs()["user"].Original())
		assert.Equal(t, fmt.Sprintf("user %s with uid %s", u.Username, u.UID), status.Diffs()["user"].Current())
		assert.True(t, status.HasChanges())
	} else {
		assert.EqualError(t, err, "user: not supported on this system")
	}
}

func TestCheckPresentWithUidNoGidAddUser(t *testing.T) {
	t.Parallel()

	u := user.NewUser(new(user.System))
	u.Username = fakeUsername
	u.UID = fakeUID
	u.State = user.StatePresent
	status, err := u.Check(fakerenderer.New())

	if runtime.GOOS == "linux" {
		assert.NoError(t, err)
		assert.Equal(t, resource.StatusWillChange, status.StatusCode())
		assert.Equal(t, fmt.Sprintf("user name and uid do not exist"), status.Messages()[0])
		assert.Equal(t, string(user.StateAbsent), status.Diffs()["user"].Original())
		assert.Equal(t, fmt.Sprintf("user %s with uid %s", u.Username, u.UID), status.Diffs()["user"].Current())
		assert.True(t, status.HasChanges())
	} else {
		assert.EqualError(t, err, "user: not supported on this system")
	}
}

func TestCheckPresentUidExists(t *testing.T) {
	t.Parallel()

	u := user.NewUser(new(user.System))
	u.Username = fakeUsername
	u.UID = currUID
	u.State = user.StatePresent
	status, err := u.Check(fakerenderer.New())

	if runtime.GOOS == "linux" {
		assert.NoError(t, err)
		assert.Equal(t, resource.StatusFatal, status.StatusCode())
		assert.Equal(t, fmt.Sprintf("user uid %s already exists", u.UID), status.Messages()[0])
		assert.False(t, status.HasChanges())
	} else {
		assert.EqualError(t, err, "user: not supported on this system")
	}
}

func TestCheckPresentUsernameExists(t *testing.T) {
	t.Parallel()

	u := user.NewUser(new(user.System))
	u.Username = currUsername
	u.State = user.StatePresent
	status, err := u.Check(fakerenderer.New())

	if runtime.GOOS == "linux" {
		assert.NoError(t, err)
		assert.Equal(t, resource.StatusFatal, status.StatusCode())
		assert.Equal(t, fmt.Sprintf("user %s already exists", u.Username), status.Messages()[0])
		assert.False(t, status.HasChanges())
	} else {
		assert.EqualError(t, err, "user: not supported on this system")
	}
}

func TestCheckPresentUserNameUidMismatch(t *testing.T) {
	t.Parallel()

	uid, err := setUid()
	if err != nil {
		panic(err)
	}
	u := user.NewUser(new(user.System))
	u.Username = currUsername
	u.UID = uid
	u.State = user.StatePresent
	status, err := u.Check(fakerenderer.New())

	if runtime.GOOS == "linux" {
		assert.NoError(t, err)
		assert.Equal(t, resource.StatusFatal, status.StatusCode())
		assert.Equal(t, fmt.Sprintf("user %s and uid %s belong to different users", u.Username, u.UID), status.Messages()[0])
		assert.False(t, status.HasChanges())
	} else {
		assert.EqualError(t, err, "user: not supported on this system")
	}
}

func TestCheckAbsentNoUidUserNotFound(t *testing.T) {
	t.Parallel()

	u := user.NewUser(new(user.System))
	u.Username = fakeUsername
	u.State = user.StateAbsent
	status, err := u.Check(fakerenderer.New())

	if runtime.GOOS == "linux" {
		assert.NoError(t, err)
		assert.Equal(t, resource.StatusFatal, status.StatusCode())
		assert.Equal(t, fmt.Sprintf("user %s does not exist", u.Username), status.Messages()[0])
		assert.False(t, status.HasChanges())
	} else {
		assert.EqualError(t, err, "user: not supported on this system")
	}
}

func TestCheckAbsentNoUidDeleteUser(t *testing.T) {
	t.Parallel()

	u := user.NewUser(new(user.System))
	u.Username = currUsername
	u.State = user.StateAbsent
	status, err := u.Check(fakerenderer.New())

	if runtime.GOOS == "linux" {
		assert.NoError(t, err)
		assert.Equal(t, resource.StatusWillChange, status.StatusCode())
		assert.Equal(t, fmt.Sprintf("user %s", u.Username), status.Diffs()["user"].Original())
		assert.Equal(t, string(user.StateAbsent), status.Diffs()["user"].Current())
		assert.True(t, status.HasChanges())
	} else {
		assert.EqualError(t, err, "user: not supported on this system")
	}
}

func TestCheckAbsentWithUidUserNotFound(t *testing.T) {
	t.Parallel()

	u := user.NewUser(new(user.System))
	u.Username = fakeUsername
	u.UID = fakeUID
	u.State = user.StateAbsent
	status, err := u.Check(fakerenderer.New())

	if runtime.GOOS == "linux" {
		assert.NoError(t, err)
		assert.Equal(t, resource.StatusNoChange, status.StatusCode())
		assert.Equal(t, fmt.Sprint("user name and uid do not exist"), status.Messages()[0])
		assert.False(t, status.HasChanges())
	} else {
		assert.EqualError(t, err, "user: not supported on this system")
	}
}

func TestCheckAbsentWithUidUsernameNotFound(t *testing.T) {
	t.Parallel()

	u := user.NewUser(new(user.System))
	u.Username = fakeUsername
	u.UID = currUID
	u.State = user.StateAbsent
	status, err := u.Check(fakerenderer.New())

	if runtime.GOOS == "linux" {
		assert.NoError(t, err)
		assert.Equal(t, resource.StatusFatal, status.StatusCode())
		assert.Equal(t, fmt.Sprintf("user %s does not exist", u.Username), status.Messages()[0])
		assert.False(t, status.HasChanges())
	} else {
		assert.EqualError(t, err, "user: not supported on this system")
	}
}

func TestCheckAbsentWithUid_UidNotFound(t *testing.T) {
	t.Parallel()

	u := user.NewUser(new(user.System))
	u.Username = currUsername
	u.UID = fakeUID
	u.State = user.StateAbsent
	status, err := u.Check(fakerenderer.New())

	if runtime.GOOS == "linux" {
		assert.NoError(t, err)
		assert.Equal(t, resource.StatusFatal, status.StatusCode())
		assert.Equal(t, fmt.Sprintf("user uid %s does not exist", u.UID), status.Messages()[0])
		assert.False(t, status.HasChanges())
	} else {
		assert.EqualError(t, err, "user: not supported on this system")
	}
}

func TestCheckAbsentWithUidUsernameUidMismatch(t *testing.T) {
	t.Parallel()

	uid, err := setUid()
	if err != nil {
		panic(err)
	}
	u := user.NewUser(new(user.System))
	u.Username = currUsername
	u.UID = uid
	u.State = user.StateAbsent
	status, err := u.Check(fakerenderer.New())

	if runtime.GOOS == "linux" {
		assert.NoError(t, err)
		assert.Equal(t, resource.StatusFatal, status.StatusCode())
		assert.Equal(t, fmt.Sprintf("user %s and uid %s belong to different users", u.Username, u.UID), status.Messages()[0])
		assert.False(t, status.HasChanges())
	} else {
		assert.EqualError(t, err, "user: not supported on this system")
	}
}

func TestCheckAbsentWithUidDeleteUser(t *testing.T) {
	t.Parallel()

	u := user.NewUser(new(user.System))
	u.Username = currUsername
	u.UID = currUID
	u.State = user.StateAbsent
	status, err := u.Check(fakerenderer.New())

	if runtime.GOOS == "linux" {
		assert.NoError(t, err)
		assert.Equal(t, resource.StatusWillChange, status.StatusCode())
		assert.Equal(t, fmt.Sprintf("user %s with uid %s", u.Username, u.UID), status.Diffs()["user"].Original())
		assert.Equal(t, string(user.StateAbsent), status.Diffs()["user"].Current())
		assert.True(t, status.HasChanges())
	} else {
		assert.EqualError(t, err, "user: not supported on this system")
	}
}

func TestCheckStateUnknown(t *testing.T) {
	t.Parallel()

	u := user.NewUser(new(user.System))
	u.Username = currUsername
	u.UID = currUID
	u.State = "test"
	_, err := u.Check(fakerenderer.New())

	if runtime.GOOS == "linux" {
		assert.EqualError(t, err, fmt.Sprintf("user: unrecognized state %s", u.State))
	} else {
		assert.EqualError(t, err, "user: not supported on this system")
	}
}

func TestApplyPresentNoUidAddUser(t *testing.T) {
	t.Parallel()

	usr := &os.User{
		Username: fakeUsername,
	}
	m := &MockSystem{}
	u := user.NewUser(m)
	u.Username = usr.Username
	u.State = user.StatePresent
	var options = map[string]string{}

	m.On("Lookup", u.Username).Return(usr, os.UnknownUserError(""))
	m.On("AddUser", u.Username, options).Return(nil)
	status, err := u.Apply()

	m.AssertCalled(t, "AddUser", u.Username, options)
	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("added user %s", u.Username), status.Messages()[0])
}

func TestApplyPresentNoUidAddUserError(t *testing.T) {
	t.Parallel()

	usr := &os.User{
		Username: fakeUsername,
	}
	m := &MockSystem{}
	u := user.NewUser(m)
	u.Username = usr.Username
	u.State = user.StatePresent
	var options = map[string]string{}

	m.On("Lookup", u.Username).Return(usr, os.UnknownUserError(""))
	m.On("AddUser", u.Username, options).Return(fmt.Errorf(""))
	status, err := u.Apply()

	m.AssertCalled(t, "AddUser", u.Username, options)
	assert.EqualError(t, err, fmt.Sprintf(""))
	assert.Equal(t, resource.StatusFatal, status.StatusCode())
	assert.Equal(t, fmt.Sprintf("error adding user %s", u.Username), status.Messages()[0])
}

func TestApplyPresentNoUidNotAdded(t *testing.T) {
	t.Parallel()

	usr := &os.User{
		Username: fakeUsername,
	}
	m := &MockSystem{}
	u := user.NewUser(m)
	u.Username = usr.Username
	u.State = user.StatePresent
	var options = map[string]string{}

	m.On("Lookup", u.Username).Return(usr, nil)
	m.On("AddUser", u.Username, options).Return(nil)
	status, err := u.Apply()

	m.AssertNotCalled(t, "AddUser", u.Username, options)
	assert.EqualError(t, err, fmt.Sprintf("will not attempt add: user %s", u.Username))
	assert.Equal(t, resource.StatusFatal, status.StatusCode())
}

func TestApplyPresentWithUidAddUser(t *testing.T) {
	t.Parallel()

	usr := &os.User{
		Username: fakeUsername,
		Uid:      fakeUID,
	}
	m := &MockSystem{}
	u := user.NewUser(m)
	u.Username = usr.Username
	u.UID = usr.Uid
	u.State = user.StatePresent
	var options = map[string]string{"uid": u.UID}

	m.On("Lookup", u.Username).Return(usr, os.UnknownUserError(""))
	m.On("LookupID", u.UID).Return(usr, os.UnknownUserIdError(1))
	m.On("AddUser", u.Username, options).Return(nil)
	status, err := u.Apply()

	m.AssertCalled(t, "AddUser", u.Username, options)
	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("added user %s with uid %s", u.Username, u.UID), status.Messages()[0])
}

func TestApplyPresentWithUidAddUserError(t *testing.T) {
	t.Parallel()

	usr := &os.User{
		Username: fakeUsername,
		Uid:      fakeUID,
	}
	m := &MockSystem{}
	u := user.NewUser(m)
	u.Username = usr.Username
	u.UID = usr.Uid
	u.State = user.StatePresent
	var options = map[string]string{"uid": u.UID}

	m.On("Lookup", u.Username).Return(usr, os.UnknownUserError(""))
	m.On("LookupID", u.UID).Return(usr, os.UnknownUserIdError(1))
	m.On("AddUser", u.Username, options).Return(fmt.Errorf(""))
	status, err := u.Apply()

	m.AssertCalled(t, "AddUser", u.Username, options)
	assert.EqualError(t, err, fmt.Sprintf(""))
	assert.Equal(t, resource.StatusFatal, status.StatusCode())
	assert.Equal(t, fmt.Sprintf("error adding user %s with uid %s", u.Username, u.UID), status.Messages()[0])
}

func TestApplyPresentWithUidNotAdded(t *testing.T) {
	t.Parallel()

	usr := &os.User{
		Username: fakeUsername,
		Uid:      fakeUID,
	}
	m := &MockSystem{}
	u := user.NewUser(m)
	u.Username = usr.Username
	u.UID = usr.Uid
	u.State = user.StatePresent
	var options = map[string]string{"uid": u.UID}

	m.On("Lookup", u.Username).Return(usr, nil)
	m.On("LookupID", u.UID).Return(usr, nil)
	m.On("AddUser", u.Username, options).Return(nil)
	status, err := u.Apply()

	m.AssertNotCalled(t, "AddUser", u.Username, options)
	assert.EqualError(t, err, fmt.Sprintf("will not attempt add: user %s with uid %s", u.Username, u.UID))
	assert.Equal(t, resource.StatusFatal, status.StatusCode())
}

func TestApplyAbsentNoUidDeleteUser(t *testing.T) {
	t.Parallel()

	usr := &os.User{
		Username: fakeUsername,
	}
	m := &MockSystem{}
	u := user.NewUser(m)
	u.Username = usr.Username
	u.State = user.StateAbsent

	m.On("Lookup", u.Username).Return(usr, nil)
	m.On("DelUser", u.Username).Return(nil)
	status, err := u.Apply()

	m.AssertCalled(t, "DelUser", u.Username)
	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("deleted user %s", u.Username), status.Messages()[0])
}

func TestApplyAbsentNoUidDeleteUserError(t *testing.T) {
	t.Parallel()

	usr := &os.User{
		Username: fakeUsername,
	}
	m := &MockSystem{}
	u := user.NewUser(m)
	u.Username = usr.Username
	u.State = user.StateAbsent

	m.On("Lookup", u.Username).Return(usr, nil)
	m.On("DelUser", u.Username).Return(fmt.Errorf(""))
	status, err := u.Apply()

	m.AssertCalled(t, "DelUser", u.Username)
	assert.EqualError(t, err, fmt.Sprintf(""))
	assert.Equal(t, resource.StatusFatal, status.StatusCode())
	assert.Equal(t, fmt.Sprintf("error deleting user %s", u.Username), status.Messages()[0])
}

func TestApplyAbsentNoUidNotDeleted(t *testing.T) {
	t.Parallel()

	usr := &os.User{
		Username: fakeUsername,
	}
	m := &MockSystem{}
	u := user.NewUser(m)
	u.Username = usr.Username
	u.State = user.StateAbsent

	m.On("Lookup", u.Username).Return(usr, os.UnknownUserError(""))
	m.On("DelUser", u.Username).Return(nil)
	status, err := u.Apply()

	m.AssertNotCalled(t, "DelUser", u.Username)
	assert.EqualError(t, err, fmt.Sprintf("will not attempt delete: user %s", u.Username))
	assert.Equal(t, resource.StatusFatal, status.StatusCode())
}

func TestApplyAbsentWithUidDeleteUser(t *testing.T) {
	t.Parallel()

	usr := &os.User{
		Username: fakeUsername,
		Uid:      fakeUID,
	}
	m := &MockSystem{}
	u := user.NewUser(m)
	u.Username = usr.Username
	u.UID = usr.Uid
	u.State = user.StateAbsent

	m.On("Lookup", u.Username).Return(usr, nil)
	m.On("LookupID", u.UID).Return(usr, nil)
	m.On("DelUser", u.Username).Return(nil)
	status, err := u.Apply()

	m.AssertCalled(t, "DelUser", u.Username)
	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("deleted user %s with uid %s", u.Username, u.UID), status.Messages()[0])
}

func TestApplyAbsentWithUidDeleteUserError(t *testing.T) {
	t.Parallel()

	usr := &os.User{
		Username: fakeUsername,
		Uid:      fakeUID,
	}
	m := &MockSystem{}
	u := user.NewUser(m)
	u.Username = usr.Username
	u.UID = usr.Uid
	u.State = user.StateAbsent

	m.On("Lookup", u.Username).Return(usr, nil)
	m.On("LookupID", u.UID).Return(usr, nil)
	m.On("DelUser", u.Username).Return(fmt.Errorf(""))
	status, err := u.Apply()

	m.AssertCalled(t, "DelUser", u.Username)
	assert.EqualError(t, err, fmt.Sprintf(""))
	assert.Equal(t, resource.StatusFatal, status.StatusCode())
	assert.Equal(t, fmt.Sprintf("error deleting user %s with uid %s", u.Username, u.UID), status.Messages()[0])
}

func TestApplyAbsentWithUidNotDeleted(t *testing.T) {
	t.Parallel()

	usr := &os.User{
		Username: fakeUsername,
		Uid:      fakeUID,
	}
	m := &MockSystem{}
	u := user.NewUser(m)
	u.Username = usr.Username
	u.UID = usr.Uid
	u.State = user.StateAbsent

	m.On("Lookup", u.Username).Return(usr, os.UnknownUserError(""))
	m.On("LookupID", u.UID).Return(usr, nil)
	m.On("DelUser", u.Username).Return(nil)
	status, err := u.Apply()

	m.AssertNotCalled(t, "DelUser", u.Username)
	assert.EqualError(t, err, fmt.Sprintf("will not attempt delete: user %s with uid %s", u.Username, u.UID))
	assert.Equal(t, resource.StatusFatal, status.StatusCode())
}

func TestApplyStateUnknown(t *testing.T) {
	t.Parallel()

	usr := &os.User{
		Username: fakeUsername,
		Uid:      fakeUID,
	}
	m := &MockSystem{}
	u := user.NewUser(m)
	u.Username = usr.Username
	u.UID = usr.Uid
	u.State = "test"
	var options = map[string]string{"uid": u.UID}

	m.On("Lookup", u.Username).Return(usr, nil)
	m.On("LookupID", u.UID).Return(usr, nil)
	m.On("AddUser", u.Username, mock.Anything).Return(nil)
	m.On("DelUser", u.Username).Return(nil)
	_, err := u.Apply()

	m.AssertNotCalled(t, "AddUser", u.Username, options)
	m.AssertNotCalled(t, "DelUser", u.Username)
	assert.EqualError(t, err, fmt.Sprintf("user: unrecognized state %s", u.State))
}

func TestSetUserOptionsAll(t *testing.T) {
	t.Parallel()

	u := user.NewUser(new(user.System))
	u.Username = fakeUsername
	u.UID = fakeUID
	u.GID = fakeGid
	u.Name = "test"
	u.HomeDir = "testDir"

	options := user.SetUserAddOptions(u)

	assert.Equal(t, u.UID, options["uid"])
	assert.Equal(t, u.GID, options["group"])
	assert.Equal(t, u.Name, options["comment"])
	assert.Equal(t, u.HomeDir, options["directory"])
}

func TestSetUserOptionsGroup(t *testing.T) {
	t.Parallel()

	u := user.NewUser(new(user.System))
	u.Username = fakeUsername
	u.GroupName = fakeGroupName
	u.GID = fakeGid

	options := user.SetUserAddOptions(u)

	assert.Equal(t, u.GroupName, options["group"])
}

func TestSetUserOptionsNone(t *testing.T) {
	t.Parallel()

	u := user.NewUser(new(user.System))
	u.Username = fakeUsername

	options := user.SetUserAddOptions(u)

	_, uidOk := options["uid"]
	_, groupOk := options["group"]
	_, commentOk := options["comment"]
	_, directoryOk := options["directory"]

	assert.False(t, uidOk)
	assert.False(t, groupOk)
	assert.False(t, commentOk)
	assert.False(t, directoryOk)
}

// setUid is used to find a uid that exists, but is not
// a match for the current user name (currUsername).
func setUid() (string, error) {
	for i := 0; i <= UIDMax; i++ {
		uid := strconv.Itoa(i)
		user, err := os.LookupId(uid)
		if err == nil && user.Username != currUsername {
			return uid, nil
		}
	}
	return "", fmt.Errorf("setUid: could not set uid")
}

// setFakeUid is used to set a uid that does not exist.
func setFakeUid() (string, error) {
	for i := UIDMin; i <= UIDMax; i++ {
		uid := strconv.Itoa(i)
		_, err := os.LookupId(uid)
		if err != nil {
			return uid, nil
		}
	}
	return "", fmt.Errorf("setFakeUid: could not set uid")
}

// setGid is used to find a gid that exists, but is not
// a match for the current user group name (currGroupName).
func setGid() (string, error) {
	for i := 0; i <= GIDMax; i++ {
		gid := strconv.Itoa(i)
		group, err := os.LookupGroupId(gid)
		if err == nil && group.Name != currGroupName {
			return gid, nil
		}
	}
	return "", fmt.Errorf("setGid: could not set gid")
}

// setFakeGid is used to set a gid that does not exist.
func setFakeGid() (string, error) {
	for i := GIDMin; i <= GIDMax; i++ {
		gid := strconv.Itoa(i)
		_, err := os.LookupGroupId(gid)
		if err != nil {
			return gid, nil
		}
	}
	return "", fmt.Errorf("setFakeGid: could not set gid")
}

// MockSystem is a mock implementation of user.System
type MockSystem struct {
	mock.Mock
}

// AddUser adds a user
func (m *MockSystem) AddUser(name string, options map[string]string) error {
	args := m.Called(name, options)
	return args.Error(0)
}

// DelUser deletes a user
func (m *MockSystem) DelUser(name string) error {
	args := m.Called(name)
	return args.Error(0)
}

// Lookup looks up a user by name
func (m *MockSystem) Lookup(name string) (*os.User, error) {
	args := m.Called(name)
	return args.Get(0).(*os.User), args.Error(1)
}

// LookupID looks up a user by ID
func (m *MockSystem) LookupID(uid string) (*os.User, error) {
	args := m.Called(uid)
	return args.Get(0).(*os.User), args.Error(1)
}

// LookupGroup looks up a group by name
func (m *MockSystem) LookupGroup(name string) (*os.Group, error) {
	args := m.Called(name)
	return args.Get(0).(*os.Group), args.Error(1)
}

// LookupGroupID looks up a group by ID
func (m *MockSystem) LookupGroupID(gid string) (*os.Group, error) {
	args := m.Called(gid)
	return args.Get(0).(*os.Group), args.Error(1)
}
