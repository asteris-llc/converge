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

var (
	currUser      *os.User
	currUsername  string
	currUID       string
	currGroup     *os.Group
	currGroupName string
	currGID       string
	userErr       error
	groupErr      error
	tempUsername  []string
	fakeUsername  string
	fakeUID       string
	tempGroupName []string
	fakeGroupName string
	fakeGID       string
	gidErr        error
	uidErr        error
)

const (
	// minGID designates the smallest valid GID
	// At a minimum, 0-32676 is valid
	minGID = 0

	// maxGID designates the largest valid GID
	// At a minimum, 0-32676 is valid
	maxGID = math.MaxInt16

	// minUID designates the smallest valid UID
	// At a minimum, 0-32676 is valid
	minUID = 0

	// maxUID designates the largest valid UID
	// At a minimum, 0-32676 is valid
	maxUID = math.MaxInt16
)

func init() {
	currUser, userErr = os.Current()
	if userErr != nil {
		panic(userErr)
	}

	currUsername = currUser.Username
	currUID = currUser.Uid

	currGID = currUser.Gid
	currGroup, groupErr = os.LookupGroupId(currGID)
	if groupErr != nil {
		panic(groupErr)
	}

	fakeUID, uidErr = setFakeUid()
	if uidErr != nil {
		panic(uidErr)
	}
	fakeGID, gidErr = setFakeGid()
	if gidErr != nil {
		panic(gidErr)
	}

	currUsername = currUser.Username
	tempUsername = strings.Split(uuid.NewV4().String(), "-")
	fakeUsername = strings.Join(tempUsername[0:], "")
	tempGroupName = strings.Split(uuid.NewV4().String(), "-")
	fakeGroupName = strings.Join(tempUsername[0:], "")
}

// TestUserInterface tests that User is properly implemented
func TestUserInterface(t *testing.T) {
	t.Parallel()

	assert.Implements(t, (*resource.Task)(nil), new(user.User))
}

// TestCheck tests the possible cases Check handles
func TestCheck(t *testing.T) {
	t.Parallel()

	t.Run("state=present", func(t *testing.T) {
		u := user.NewUser(new(user.System))
		u.State = user.StatePresent

		t.Run("uid not provided", func(t *testing.T) {
			t.Run("no add-user already exists", func(t *testing.T) {
				u.Username = currUsername
				status, err := u.Check(fakerenderer.New())

				if runtime.GOOS == "linux" {
					assert.NoError(t, err)
					assert.Equal(t, resource.StatusNoChange, status.StatusCode())
					assert.Equal(t, fmt.Sprintf("user %s already exists", u.Username), status.Messages()[0])
					assert.False(t, status.HasChanges())
				} else {
					assert.EqualError(t, err, "user: not supported on this system")
				}
			})

			t.Run("add user", func(t *testing.T) {
				u.Username = fakeUsername
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
			})

			t.Run("group provided", func(t *testing.T) {
				t.Run("no add-group name does not exist", func(t *testing.T) {
					u.Username = fakeUsername
					u.GroupName = fakeGroupName
					status, err := u.Check(fakerenderer.New())

					if runtime.GOOS == "linux" {
						assert.EqualError(t, err, fmt.Sprintf("cannot add user %s", u.Username))
						assert.Equal(t, resource.StatusCantChange, status.StatusCode())
						assert.Equal(t, fmt.Sprintf("group %s does not exist", u.GroupName), status.Messages()[0])
						assert.True(t, status.HasChanges())
					} else {
						assert.EqualError(t, err, "user: not supported on this system")
					}
				})

				t.Run("add user with group name", func(t *testing.T) {
					u.Username = fakeUsername
					u.GroupName = currGroupName
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
				})

				t.Run("no add-group gid does not exist", func(t *testing.T) {
					u.Username = fakeUsername
					u.GID = fakeGID
					status, err := u.Check(fakerenderer.New())

					if runtime.GOOS == "linux" {
						assert.EqualError(t, err, fmt.Sprintf("cannot add user %s", u.Username))
						assert.Equal(t, resource.StatusCantChange, status.StatusCode())
						assert.Equal(t, fmt.Sprintf("group gid %s does not exist", u.GID), status.Messages()[0])
						assert.True(t, status.HasChanges())
					} else {
						assert.EqualError(t, err, "user: not supported on this system")
					}
				})

				t.Run("add user with group gid", func(t *testing.T) {
					u.Username = fakeUsername
					u.GID = currGID
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
				})
			})
		})

		t.Run("uid provided", func(t *testing.T) {
			t.Run("add user with uid", func(t *testing.T) {
				u.Username = fakeUsername
				u.UID = fakeUID
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
			})

			t.Run("group provided", func(t *testing.T) {
				t.Run("no add-group name does not exist", func(t *testing.T) {
					u.Username = fakeUsername
					u.UID = fakeUID
					u.GroupName = fakeGroupName
					status, err := u.Check(fakerenderer.New())

					if runtime.GOOS == "linux" {
						assert.EqualError(t, err, fmt.Sprintf("cannot add user %s with uid %s", u.Username, u.UID))
						assert.Equal(t, resource.StatusCantChange, status.StatusCode())
						assert.Equal(t, fmt.Sprintf("group %s does not exist", u.GroupName), status.Messages()[0])
						assert.True(t, status.HasChanges())
					} else {
						assert.EqualError(t, err, "user: not supported on this system")
					}
				})

				t.Run("add user with group name", func(t *testing.T) {
					u.Username = fakeUsername
					u.UID = fakeUID
					u.GroupName = currGroupName
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
				})

				t.Run("no add-group gid does not exist", func(t *testing.T) {
					u.Username = fakeUsername
					u.UID = fakeUID
					u.GID = fakeGID
					status, err := u.Check(fakerenderer.New())

					if runtime.GOOS == "linux" {
						assert.EqualError(t, err, fmt.Sprintf("cannot add user %s with uid %s", u.Username, u.UID))
						assert.Equal(t, resource.StatusCantChange, status.StatusCode())
						assert.Equal(t, fmt.Sprintf("group gid %s does not exist", u.GID), status.Messages()[0])
						assert.True(t, status.HasChanges())
					} else {
						assert.EqualError(t, err, "user: not supported on this system")
					}
				})

				t.Run("add user with group gid", func(t *testing.T) {
					gid, err := setGid()
					if err != nil {
						panic(err)
					}
					u.Username = fakeUsername
					u.UID = fakeUID
					u.GID = gid
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
				})
			})

			t.Run("no add-user uid already exists", func(t *testing.T) {
				u.Username = fakeUsername
				u.UID = currUID
				status, err := u.Check(fakerenderer.New())

				if runtime.GOOS == "linux" {
					assert.EqualError(t, err, fmt.Sprintf("cannot add user %s with uid %s", u.Username, u.UID))
					assert.Equal(t, resource.StatusCantChange, status.StatusCode())
					assert.Equal(t, fmt.Sprintf("user uid %s already exists", u.UID), status.Messages()[0])
					assert.True(t, status.HasChanges())
				} else {
					assert.EqualError(t, err, "user: not supported on this system")
				}
			})

			t.Run("no add-user name already exists", func(t *testing.T) {
				u.Username = currUsername
				u.UID = fakeUID
				status, err := u.Check(fakerenderer.New())

				if runtime.GOOS == "linux" {
					assert.EqualError(t, err, fmt.Sprintf("cannot add user %s with uid %s", u.Username, u.UID))
					assert.Equal(t, resource.StatusCantChange, status.StatusCode())
					assert.Equal(t, fmt.Sprintf("user %s already exists", u.Username), status.Messages()[0])
					assert.True(t, status.HasChanges())
				} else {
					assert.EqualError(t, err, "user: not supported on this system")
				}
			})

			t.Run("no add-user name and uid belong to different users", func(t *testing.T) {
				uid, err := setUid()
				if err != nil {
					panic(err)
				}
				u.Username = currUsername
				u.UID = uid
				status, err := u.Check(fakerenderer.New())

				if runtime.GOOS == "linux" {
					assert.EqualError(t, err, fmt.Sprintf("cannot add user %s with uid %s", u.Username, u.UID))
					assert.Equal(t, resource.StatusCantChange, status.StatusCode())
					assert.Equal(t, fmt.Sprintf("user %s and uid %s belong to different users", u.Username, u.UID), status.Messages()[0])
					assert.True(t, status.HasChanges())
				} else {
					assert.EqualError(t, err, "user: not supported on this system")
				}
			})
		})
	})

	t.Run("state=absent", func(t *testing.T) {
		u := user.NewUser(new(user.System))
		u.State = user.StateAbsent

		t.Run("uid not provided", func(t *testing.T) {
			t.Run("no delete-user does not exist", func(t *testing.T) {
				u.Username = fakeUsername
				status, err := u.Check(fakerenderer.New())

				if runtime.GOOS == "linux" {
					assert.NoError(t, err)
					assert.Equal(t, resource.StatusNoChange, status.StatusCode())
					assert.Equal(t, fmt.Sprintf("user %s does not exist", u.Username), status.Messages()[0])
					assert.False(t, status.HasChanges())
				} else {
					assert.EqualError(t, err, "user: not supported on this system")
				}
			})

			t.Run("delete user", func(t *testing.T) {
				u.Username = currUsername
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
			})
		})

		t.Run("uid provided", func(t *testing.T) {
			t.Run("no delete-user name and uid do not exist", func(t *testing.T) {
				u.Username = fakeUsername
				u.UID = fakeUID
				status, err := u.Check(fakerenderer.New())

				if runtime.GOOS == "linux" {
					assert.NoError(t, err)
					assert.Equal(t, resource.StatusNoChange, status.StatusCode())
					assert.Equal(t, fmt.Sprint("user name and uid do not exist"), status.Messages()[0])
					assert.False(t, status.HasChanges())
				} else {
					assert.EqualError(t, err, "user: not supported on this system")
				}
			})

			t.Run("no delete-user name does not exist", func(t *testing.T) {
				u.Username = fakeUsername
				u.UID = currUID
				status, err := u.Check(fakerenderer.New())

				if runtime.GOOS == "linux" {
					assert.EqualError(t, err, fmt.Sprintf("cannot delete user %s with uid %s", u.Username, u.UID))
					assert.Equal(t, resource.StatusCantChange, status.StatusCode())
					assert.Equal(t, fmt.Sprintf("user %s does not exist", u.Username), status.Messages()[0])
					assert.True(t, status.HasChanges())
				} else {
					assert.EqualError(t, err, "user: not supported on this system")
				}
			})

			t.Run("no delete-user uid does not exist", func(t *testing.T) {
				u.Username = currUsername
				u.UID = fakeUID
				status, err := u.Check(fakerenderer.New())

				if runtime.GOOS == "linux" {
					assert.EqualError(t, err, fmt.Sprintf("cannot delete user %s with uid %s", u.Username, u.UID))
					assert.Equal(t, resource.StatusCantChange, status.StatusCode())
					assert.Equal(t, fmt.Sprintf("user uid %s does not exist", u.UID), status.Messages()[0])
					assert.True(t, status.HasChanges())
				} else {
					assert.EqualError(t, err, "user: not supported on this system")
				}
			})

			t.Run("no delete-user name and uid belong to different users", func(t *testing.T) {
				uid, err := setUid()
				if err != nil {
					panic(err)
				}
				u.Username = currUsername
				u.UID = uid
				status, err := u.Check(fakerenderer.New())

				if runtime.GOOS == "linux" {
					assert.EqualError(t, err, fmt.Sprintf("cannot delete user %s with uid %s", u.Username, u.UID))
					assert.Equal(t, resource.StatusCantChange, status.StatusCode())
					assert.Equal(t, fmt.Sprintf("user %s and uid %s belong to different users", u.Username, u.UID), status.Messages()[0])
					assert.True(t, status.HasChanges())
				} else {
					assert.EqualError(t, err, "user: not supported on this system")
				}
			})

			t.Run("delete user with uid", func(t *testing.T) {
				u.Username = currUsername
				u.UID = currUID
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
			})
		})
	})

	t.Run("state unknown", func(t *testing.T) {
		u := user.NewUser(new(user.System))
		u.Username = currUsername
		u.UID = currUID
		u.State = "test"
		status, err := u.Check(fakerenderer.New())

		if runtime.GOOS == "linux" {
			assert.EqualError(t, err, fmt.Sprintf("user: unrecognized state %s", u.State))
			assert.Equal(t, resource.StatusFatal, status.StatusCode())
		} else {
			assert.EqualError(t, err, "user: not supported on this system")
		}
	})

}

// TestApply tests all possible cases Apply handles
func TestApply(t *testing.T) {
	t.Parallel()

	t.Run("state=present", func(t *testing.T) {
		t.Run("uid not provided", func(t *testing.T) {
			t.Run("add user", func(t *testing.T) {
				usr := &os.User{
					Username: fakeUsername,
				}
				m := &MockSystem{}
				o := &MockOptions{}
				u := user.NewUser(m)
				u.Username = usr.Username
				u.State = user.StatePresent
				options := user.AddUserOptions{}

				m.On("Lookup", u.Username).Return(usr, os.UnknownUserError(""))
				o.On("SetAddUserOptions", u).Return(options, nil)
				m.On("AddUser", u.Username, &options).Return(nil)
				status, err := u.Apply()

				m.AssertCalled(t, "AddUser", u.Username, &options)
				assert.NoError(t, err)
				assert.Equal(t, fmt.Sprintf("added user %s", u.Username), status.Messages()[0])
			})

			t.Run("no add-group name does not exist", func(t *testing.T) {
				usr := &os.User{
					Username: fakeUsername,
				}
				grp := &os.Group{
					Name: fakeGroupName,
				}
				m := &MockSystem{}
				o := &MockOptions{}
				u := user.NewUser(m)
				u.Username = usr.Username
				u.GroupName = grp.Name
				u.State = user.StatePresent
				options := user.AddUserOptions{}
				optErr := fmt.Sprintf("group %s does not exist", u.GroupName)

				m.On("Lookup", u.Username).Return(usr, os.UnknownUserError(""))
				m.On("LookupGroup", u.GroupName).Return(grp, os.UnknownGroupError(""))
				o.On("SetAddUserOptions", u).Return(nil, optErr)
				m.On("AddUser", u.Username, &options).Return(nil)
				status, err := u.Apply()

				m.AssertNotCalled(t, "AddUser", u.Username, &options)
				assert.EqualError(t, err, fmt.Sprintf("will not attempt add: user %s", u.Username))
				assert.Equal(t, optErr, status.Messages()[0])
			})

			t.Run("no add-group gid does not exist", func(t *testing.T) {
				usr := &os.User{
					Username: fakeUsername,
				}
				grp := &os.Group{
					Gid: fakeGID,
				}
				m := &MockSystem{}
				o := &MockOptions{}
				u := user.NewUser(m)
				u.Username = usr.Username
				u.GID = grp.Gid
				u.State = user.StatePresent
				options := user.AddUserOptions{}
				optErr := fmt.Sprintf("group gid %s does not exist", u.GID)

				m.On("Lookup", u.Username).Return(usr, os.UnknownUserError(""))
				m.On("LookupGroupID", u.GID).Return(grp, os.UnknownGroupError(""))
				o.On("SetAddUserOptions", u).Return(nil, optErr)
				m.On("AddUser", u.Username, &options).Return(nil)
				status, err := u.Apply()

				m.AssertNotCalled(t, "AddUser", u.Username, &options)
				assert.EqualError(t, err, fmt.Sprintf("will not attempt add: user %s", u.Username))
				assert.Equal(t, optErr, status.Messages()[0])
			})

			t.Run("no add-error adding user", func(t *testing.T) {
				usr := &os.User{
					Username: fakeUsername,
				}
				m := &MockSystem{}
				o := &MockOptions{}
				u := user.NewUser(m)
				u.Username = usr.Username
				u.State = user.StatePresent
				options := user.AddUserOptions{}

				m.On("Lookup", u.Username).Return(usr, os.UnknownUserError(""))
				o.On("SetAddUserOptions", u).Return(options, nil)
				m.On("AddUser", u.Username, &options).Return(fmt.Errorf(""))
				status, err := u.Apply()

				m.AssertCalled(t, "AddUser", u.Username, &options)
				assert.EqualError(t, err, fmt.Sprintf("user add: "))
				assert.Equal(t, resource.StatusFatal, status.StatusCode())
				assert.Equal(t, fmt.Sprintf("error adding user %s", u.Username), status.Messages()[0])
			})

			t.Run("no add-will not attempt add", func(t *testing.T) {
				usr := &os.User{
					Username: fakeUsername,
				}
				m := &MockSystem{}
				o := &MockOptions{}
				u := user.NewUser(m)
				u.Username = usr.Username
				u.State = user.StatePresent
				options := user.AddUserOptions{}

				m.On("Lookup", u.Username).Return(usr, nil)
				o.On("SetAddUserOptions", u).Return(options, nil)
				m.On("AddUser", u.Username, &options).Return(nil)
				status, err := u.Apply()

				m.AssertNotCalled(t, "AddUser", u.Username, &options)
				assert.EqualError(t, err, fmt.Sprintf("will not attempt add: user %s", u.Username))
				assert.Equal(t, resource.StatusCantChange, status.StatusCode())
			})
		})

		t.Run("uid provided", func(t *testing.T) {
			t.Run("add user with uid", func(t *testing.T) {
				usr := &os.User{
					Username: fakeUsername,
					Uid:      fakeUID,
				}
				m := &MockSystem{}
				o := &MockOptions{}
				u := user.NewUser(m)
				u.Username = usr.Username
				u.UID = usr.Uid
				u.State = user.StatePresent
				options := user.AddUserOptions{UID: u.UID}

				m.On("Lookup", u.Username).Return(usr, os.UnknownUserError(""))
				m.On("LookupID", u.UID).Return(usr, os.UnknownUserIdError(1))
				o.On("SetAddUserOptions", u).Return(options, nil)
				m.On("AddUser", u.Username, &options).Return(nil)
				status, err := u.Apply()

				m.AssertCalled(t, "AddUser", u.Username, &options)
				assert.NoError(t, err)
				assert.Equal(t, fmt.Sprintf("added user %s with uid %s", u.Username, u.UID), status.Messages()[0])
			})

			t.Run("no add-group name does not exist", func(t *testing.T) {
				usr := &os.User{
					Username: fakeUsername,
					Uid:      fakeUID,
				}
				grp := &os.Group{
					Name: fakeGroupName,
				}
				m := &MockSystem{}
				o := &MockOptions{}
				u := user.NewUser(m)
				u.Username = usr.Username
				u.GroupName = grp.Name
				u.State = user.StatePresent
				options := user.AddUserOptions{}
				optErr := fmt.Sprintf("group %s does not exist", u.GroupName)

				m.On("Lookup", u.Username).Return(usr, os.UnknownUserError(""))
				m.On("LookupID", u.UID).Return(usr, os.UnknownUserIdError(1))
				m.On("LookupGroup", u.GroupName).Return(grp, os.UnknownGroupError(""))
				o.On("SetAddUserOptions", u).Return(nil, optErr)
				m.On("AddUser", u.Username, &options).Return(nil)
				status, err := u.Apply()

				m.AssertNotCalled(t, "AddUser", u.Username, &options)
				assert.EqualError(t, err, fmt.Sprintf("will not attempt add: user %s", u.Username))
				assert.Equal(t, optErr, status.Messages()[0])
			})

			t.Run("no add-group gid does not exist", func(t *testing.T) {
				usr := &os.User{
					Username: fakeUsername,
					Uid:      fakeUID,
				}
				grp := &os.Group{
					Gid: fakeGID,
				}
				m := &MockSystem{}
				o := &MockOptions{}
				u := user.NewUser(m)
				u.Username = usr.Username
				u.GID = grp.Gid
				u.State = user.StatePresent
				options := user.AddUserOptions{}
				optErr := fmt.Sprintf("group gid %s does not exist", u.GID)

				m.On("Lookup", u.Username).Return(usr, os.UnknownUserError(""))
				m.On("LookupID", u.UID).Return(usr, os.UnknownUserIdError(1))
				m.On("LookupGroupID", u.GID).Return(grp, os.UnknownGroupError(""))
				o.On("SetAddUserOptions", u).Return(nil, optErr)
				m.On("AddUser", u.Username, &options).Return(nil)
				status, err := u.Apply()

				m.AssertNotCalled(t, "AddUser", u.Username, &options)
				assert.EqualError(t, err, fmt.Sprintf("will not attempt add: user %s", u.Username))
				assert.Equal(t, optErr, status.Messages()[0])
			})

			t.Run("no add-error adding user", func(t *testing.T) {
				usr := &os.User{
					Username: fakeUsername,
					Uid:      fakeUID,
				}
				m := &MockSystem{}
				o := &MockOptions{}
				u := user.NewUser(m)
				u.Username = usr.Username
				u.UID = usr.Uid
				u.State = user.StatePresent
				options := user.AddUserOptions{UID: u.UID}

				m.On("Lookup", u.Username).Return(usr, os.UnknownUserError(""))
				m.On("LookupID", u.UID).Return(usr, os.UnknownUserIdError(1))
				o.On("SetAddUserOptions", u).Return(options, nil)
				m.On("AddUser", u.Username, &options).Return(fmt.Errorf(""))
				status, err := u.Apply()

				m.AssertCalled(t, "AddUser", u.Username, &options)
				assert.EqualError(t, err, fmt.Sprintf("user add: "))
				assert.Equal(t, resource.StatusFatal, status.StatusCode())
				assert.Equal(t, fmt.Sprintf("error adding user %s with uid %s", u.Username, u.UID), status.Messages()[0])
			})

			t.Run("no add-will not attempt add", func(t *testing.T) {
				usr := &os.User{
					Username: fakeUsername,
					Uid:      fakeUID,
				}
				m := &MockSystem{}
				o := &MockOptions{}
				u := user.NewUser(m)
				u.Username = usr.Username
				u.UID = usr.Uid
				u.State = user.StatePresent
				options := user.AddUserOptions{UID: u.UID}

				m.On("Lookup", u.Username).Return(usr, nil)
				m.On("LookupID", u.UID).Return(usr, nil)
				o.On("SetAddUserOptions", u).Return(options, nil)
				m.On("AddUser", u.Username, &options).Return(nil)
				status, err := u.Apply()

				m.AssertNotCalled(t, "AddUser", u.Username, &options)
				assert.EqualError(t, err, fmt.Sprintf("will not attempt add: user %s with uid %s", u.Username, u.UID))
				assert.Equal(t, resource.StatusCantChange, status.StatusCode())
			})
		})
	})

	t.Run("state=absent", func(t *testing.T) {
		t.Run("uid not provided", func(t *testing.T) {
			t.Run("delete user", func(t *testing.T) {
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
			})

			t.Run("no delete-error deleting user", func(t *testing.T) {
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
				assert.EqualError(t, err, fmt.Sprintf("user delete: "))
				assert.Equal(t, resource.StatusFatal, status.StatusCode())
				assert.Equal(t, fmt.Sprintf("error deleting user %s", u.Username), status.Messages()[0])
			})

			t.Run("no delete-will not attempt delete", func(t *testing.T) {
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
				assert.Equal(t, resource.StatusCantChange, status.StatusCode())
			})
		})

		t.Run("uid provided", func(t *testing.T) {
			t.Run("delete user with uid", func(t *testing.T) {
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
			})

			t.Run("no delete-error deleting user", func(t *testing.T) {
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
				assert.EqualError(t, err, fmt.Sprintf("user delete: "))
				assert.Equal(t, resource.StatusFatal, status.StatusCode())
				assert.Equal(t, fmt.Sprintf("error deleting user %s with uid %s", u.Username, u.UID), status.Messages()[0])
			})

			t.Run("no delete-will not attempt delete", func(t *testing.T) {
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
				assert.Equal(t, resource.StatusCantChange, status.StatusCode())
			})
		})
	})

	t.Run("state unknown", func(t *testing.T) {
		usr := &os.User{
			Username: fakeUsername,
			Uid:      fakeUID,
		}
		m := &MockSystem{}
		o := &MockOptions{}
		u := user.NewUser(m)
		u.Username = usr.Username
		u.UID = usr.Uid
		u.State = "test"
		options := user.AddUserOptions{UID: u.UID}

		m.On("Lookup", u.Username).Return(usr, nil)
		m.On("LookupID", u.UID).Return(usr, nil)
		o.On("SetAddUserOptions", u).Return(options, nil)
		m.On("AddUser", u.Username, &options).Return(nil)
		m.On("DelUser", u.Username).Return(nil)
		status, err := u.Apply()

		o.AssertNotCalled(t, "SetAddUserOptions", u)
		m.AssertNotCalled(t, "AddUser", u.Username, &options)
		m.AssertNotCalled(t, "DelUser", u.Username)
		assert.EqualError(t, err, fmt.Sprintf("user: unrecognized state %s", u.State))
		assert.Equal(t, resource.StatusFatal, status.StatusCode())
	})
}

// TestSetAddUserOptions tests options provided for adding a user
// are properly set
func TestSetAddUserOptions(t *testing.T) {
	t.Parallel()
	gid, err := setGid()
	if err != nil {
		panic(err)
	}
	grp, err := os.LookupGroupId(gid)
	if err != nil {
		panic(err)
	}

	t.Run("all options", func(t *testing.T) {
		u := user.NewUser(new(user.System))
		u.Username = fakeUsername
		u.UID = fakeUID
		u.GID = gid
		u.Name = "test"
		u.HomeDir = "testDir"

		options, err := user.SetAddUserOptions(u)

		assert.NoError(t, err)
		assert.Equal(t, u.UID, options.UID)
		assert.Equal(t, u.GID, options.Group)
		assert.Equal(t, u.Name, options.Comment)
		assert.Equal(t, u.HomeDir, options.Directory)
	})

	t.Run("group options", func(t *testing.T) {
		t.Run("group name and gid", func(t *testing.T) {
			u := user.NewUser(new(user.System))
			u.Username = fakeUsername
			u.GroupName = grp.Name
			u.GID = gid

			options, err := user.SetAddUserOptions(u)

			assert.NoError(t, err)
			assert.Equal(t, u.GroupName, options.Group)
		})

		t.Run("with group name", func(t *testing.T) {
			u := user.NewUser(new(user.System))
			u.Username = fakeUsername
			u.GroupName = grp.Name

			options, err := user.SetAddUserOptions(u)

			assert.NoError(t, err)
			assert.Equal(t, u.GroupName, options.Group)
		})

		t.Run("group name not found", func(t *testing.T) {
			u := user.NewUser(new(user.System))
			u.Username = fakeUsername
			u.GroupName = fakeGroupName

			options, err := user.SetAddUserOptions(u)

			assert.EqualError(t, err, fmt.Sprintf("group %s does not exist", u.GroupName))
			assert.Nil(t, options)
		})

		t.Run("with group gid", func(t *testing.T) {
			u := user.NewUser(new(user.System))
			u.Username = fakeUsername
			u.GID = gid

			options, err := user.SetAddUserOptions(u)

			assert.NoError(t, err)
			assert.Equal(t, u.GID, options.Group)
		})

		t.Run("group gid not found", func(t *testing.T) {
			u := user.NewUser(new(user.System))
			u.Username = fakeUsername
			u.GID = fakeGID

			options, err := user.SetAddUserOptions(u)

			assert.EqualError(t, err, fmt.Sprintf("group gid %s does not exist", u.GID))
			assert.Nil(t, options)
		})
	})

	t.Run("no options", func(t *testing.T) {
		u := user.NewUser(new(user.System))
		u.Username = fakeUsername

		options, err := user.SetAddUserOptions(u)

		assert.NoError(t, err)
		assert.Equal(t, "", options.UID)
		assert.Equal(t, "", options.Group)
		assert.Equal(t, "", options.Comment)
		assert.Equal(t, "", options.Directory)
	})
}

// setUid is used to find a uid that exists, but is not
// a match for the current user name (currUsername).
func setUid() (string, error) {
	for i := 0; i <= maxUID; i++ {
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
	for i := minUID; i <= maxUID; i++ {
		uid := strconv.Itoa(i)
		_, err := os.LookupId(uid)
		if err != nil {
			return uid, nil
		}
	}
	return "", fmt.Errorf("setFakeUid: could not set uid")
}

// setGid is used to find a gid that exists.
func setGid() (string, error) {
	for i := 0; i <= maxGID; i++ {
		gid := strconv.Itoa(i)
		_, err := os.LookupGroupId(gid)
		if err == nil {
			return gid, nil
		}
	}
	return "", fmt.Errorf("setGid: could not set gid")
}

// setFakeGid is used to set a gid that does not exist.
func setFakeGid() (string, error) {
	for i := minGID; i <= maxGID; i++ {
		gid := strconv.Itoa(i)
		_, err := os.LookupGroupId(gid)
		if err != nil {
			return gid, nil
		}
	}
	return "", fmt.Errorf("setFakeGid: could not set gid")
}

// MockOptions is a mock implementation for setting user options
type MockOptions struct {
	mock.Mock
}

// SetAddUserOptions sets the options for adding a user
func (m *MockOptions) SetAddUserOptions(u *user.User) (*user.AddUserOptions, error) {
	args := m.Called(u)
	return args.Get(0).(*user.AddUserOptions), args.Error(1)
}

// MockSystem is a mock implementation of user.System
type MockSystem struct {
	mock.Mock
}

// AddUser adds a user
func (m *MockSystem) AddUser(name string, options *user.AddUserOptions) error {
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
