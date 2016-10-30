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
	"golang.org/x/net/context"
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
	minGID = 0
	maxGID = math.MaxInt16
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

// TestGroupInterface tests that Group is properly implemented
func TestGroupInterface(t *testing.T) {
	t.Parallel()

	assert.Implements(t, (*resource.Task)(nil), new(group.Group))
}

// TestCheck tests all possible cases Check handles
func TestCheck(t *testing.T) {
	t.Parallel()

	t.Run("state=present", func(t *testing.T) {
		g := group.NewGroup(new(group.System))
		g.State = group.StatePresent
		gid, err := setGid()
		if err != nil {
			panic(err)
		}
		tempGroup, err := user.LookupGroupId(gid)
		if err != nil {
			panic(err)
		}

		t.Run("gid not provided", func(t *testing.T) {
			t.Run("NewName not provided", func(t *testing.T) {
				t.Run("no add-group already exists", func(t *testing.T) {
					g.Name = currName
					status, err := g.Check(context.Background(), fakerenderer.New())

					if runtime.GOOS == "linux" {
						assert.NoError(t, err)
						assert.Equal(t, resource.StatusNoChange, status.StatusCode())
						assert.Equal(t, fmt.Sprintf("group add: group %s already exists", g.Name), status.Messages()[0])
						assert.False(t, status.HasChanges())
					} else {
						assert.EqualError(t, err, "group: not supported on this system")
					}
				})

				t.Run("add group", func(t *testing.T) {
					g.Name = fakeName
					status, err := g.Check(context.Background(), fakerenderer.New())

					if runtime.GOOS == "linux" {
						assert.NoError(t, err)
						assert.Equal(t, resource.StatusWillChange, status.StatusCode())
						assert.Equal(t, "add group", status.Messages()[0])
						assert.Equal(t, string(group.StateAbsent), status.Diffs()["group"].Original())
						assert.Equal(t, fmt.Sprintf("group %s", g.Name), status.Diffs()["group"].Current())
						assert.True(t, status.HasChanges())
					} else {
						assert.EqualError(t, err, "group: not supported on this system")
					}
				})
			})

			t.Run("NewName provided", func(t *testing.T) {
				t.Run("no modify-group does not exist", func(t *testing.T) {
					g.Name = fakeName
					g.NewName = fakeName
					status, err := g.Check(context.Background(), fakerenderer.New())

					if runtime.GOOS == "linux" {
						assert.EqualError(t, err, "cannot modify group")
						assert.Equal(t, resource.StatusCantChange, status.StatusCode())
						assert.Equal(t, fmt.Sprintf("group modify: group %s does not exist", g.Name), status.Messages()[0])
						assert.True(t, status.HasChanges())
					} else {
						assert.EqualError(t, err, "group: not supported on this system")
					}
				})

				t.Run("no modify-new group already exists", func(t *testing.T) {
					g.Name = currName
					g.NewName = tempGroup.Name
					status, err := g.Check(context.Background(), fakerenderer.New())

					if runtime.GOOS == "linux" {
						assert.EqualError(t, err, "cannot modify group")
						assert.Equal(t, resource.StatusCantChange, status.StatusCode())
						assert.Equal(t, fmt.Sprintf("group modify: group %s already exists", g.NewName), status.Messages()[0])
						assert.True(t, status.HasChanges())
					} else {
						assert.EqualError(t, err, "group: not supported on this system")
					}
				})

				t.Run("modify group", func(t *testing.T) {
					g.Name = currName
					g.NewName = fakeName
					status, err := g.Check(context.Background(), fakerenderer.New())

					if runtime.GOOS == "linux" {
						assert.NoError(t, err)
						assert.Equal(t, resource.StatusWillChange, status.StatusCode())
						assert.Equal(t, "modify group name", status.Messages()[0])
						assert.Equal(t, fmt.Sprintf("group %s", g.Name), status.Diffs()["group"].Original())
						assert.Equal(t, fmt.Sprintf("group %s", g.NewName), status.Diffs()["group"].Current())
						assert.True(t, status.HasChanges())
					} else {
						assert.EqualError(t, err, "group: not supported on this system")
					}
				})
			})
		})

		t.Run("gid provided", func(t *testing.T) {
			t.Run("NewName not provided", func(t *testing.T) {
				g.NewName = ""

				t.Run("add group with gid", func(t *testing.T) {
					g.GID = fakeGid
					g.Name = fakeName
					status, err := g.Check(context.Background(), fakerenderer.New())

					if runtime.GOOS == "linux" {
						assert.NoError(t, err)
						assert.Equal(t, resource.StatusWillChange, status.StatusCode())
						assert.Equal(t, "add group with gid", status.Messages()[0])
						assert.Equal(t, string(group.StateAbsent), status.Diffs()["group"].Original())
						assert.Equal(t, fmt.Sprintf("group %s with gid %s", g.Name, g.GID), status.Diffs()["group"].Current())
						assert.True(t, status.HasChanges())
					} else {
						assert.EqualError(t, err, "group: not supported on this system")
					}
				})

				t.Run("no add-group gid already exists", func(t *testing.T) {
					g.GID = currGid
					g.Name = fakeName
					status, err := g.Check(context.Background(), fakerenderer.New())

					if runtime.GOOS == "linux" {
						assert.EqualError(t, err, "cannot add group")
						assert.Equal(t, resource.StatusCantChange, status.StatusCode())
						assert.Equal(t, fmt.Sprintf("group add: gid %s already exists", g.GID), status.Messages()[0])
						assert.True(t, status.HasChanges())
					} else {
						assert.EqualError(t, err, "group: not supported on this system")
					}
				})

				t.Run("modify group gid", func(t *testing.T) {
					g.GID = fakeGid
					g.Name = currName
					status, err := g.Check(context.Background(), fakerenderer.New())

					if runtime.GOOS == "linux" {
						assert.NoError(t, err)
						assert.Equal(t, resource.StatusWillChange, status.StatusCode())
						assert.Equal(t, "modify group gid", status.Messages()[0])
						assert.Equal(t, fmt.Sprintf("group %s with gid %s", g.Name, currGid), status.Diffs()["group"].Original())
						assert.Equal(t, fmt.Sprintf("group %s with gid %s", g.Name, g.GID), status.Diffs()["group"].Current())
						assert.True(t, status.HasChanges())
					} else {
						assert.EqualError(t, err, "group: not supported on this system")
					}
				})

				t.Run("no add or modify-group name and gid belong to different groups", func(t *testing.T) {
					g.GID = gid
					g.Name = currName
					status, err := g.Check(context.Background(), fakerenderer.New())

					if runtime.GOOS == "linux" {
						assert.EqualError(t, err, "cannot add or modify group")
						assert.Equal(t, resource.StatusCantChange, status.StatusCode())
						assert.Equal(t, fmt.Sprintf("group add/modify: group %s and gid %s belong to different groups", g.Name, g.GID), status.Messages()[0])
						assert.True(t, status.HasChanges())
					} else {
						assert.EqualError(t, err, "group: not supported on this system")
					}
				})

				t.Run("no add or modify-group with gid already exists", func(t *testing.T) {
					g.GID = currGid
					g.Name = currName
					g.State = group.StatePresent
					status, err := g.Check(context.Background(), fakerenderer.New())

					if runtime.GOOS == "linux" {
						assert.EqualError(t, err, "cannot add or modify group")
						assert.Equal(t, resource.StatusCantChange, status.StatusCode())
						assert.Equal(t, fmt.Sprintf("group add/modify: group %s with gid %s already exists", g.Name, g.GID), status.Messages()[0])
						assert.True(t, status.HasChanges())
					} else {
						assert.EqualError(t, err, "group: not supported on this system")
					}
				})
			})

			t.Run("NewName provided", func(t *testing.T) {
				t.Run("modify group name and gid", func(t *testing.T) {
					g.GID = fakeGid
					g.Name = currName
					g.NewName = fakeName
					status, err := g.Check(context.Background(), fakerenderer.New())

					if runtime.GOOS == "linux" {
						assert.NoError(t, err)
						assert.Equal(t, resource.StatusWillChange, status.StatusCode())
						assert.Equal(t, "modify group name and gid", status.Messages()[0])
						assert.Equal(t, fmt.Sprintf("group %s with gid %s", g.Name, currGid), status.Diffs()["group"].Original())
						assert.Equal(t, fmt.Sprintf("group %s with gid %s", g.NewName, g.GID), status.Diffs()["group"].Current())
						assert.True(t, status.HasChanges())
					} else {
						assert.EqualError(t, err, "group: not supported on this system")
					}
				})

				t.Run("no modify-group already exists", func(t *testing.T) {
					g.GID = fakeGid
					g.Name = currName
					g.NewName = tempGroup.Name
					status, err := g.Check(context.Background(), fakerenderer.New())

					if runtime.GOOS == "linux" {
						assert.EqualError(t, err, "cannot modify group")
						assert.Equal(t, resource.StatusCantChange, status.StatusCode())
						assert.Equal(t, fmt.Sprintf("group modify: group %s already exists", g.NewName), status.Messages()[0])
						assert.True(t, status.HasChanges())
					} else {
						assert.EqualError(t, err, "group: not supported on this system")
					}
				})

				t.Run("no modify-gid already exists", func(t *testing.T) {
					g.GID = gid
					g.Name = currName
					g.NewName = fakeName
					status, err := g.Check(context.Background(), fakerenderer.New())

					if runtime.GOOS == "linux" {
						assert.EqualError(t, err, "cannot modify group")
						assert.Equal(t, resource.StatusCantChange, status.StatusCode())
						assert.Equal(t, fmt.Sprintf("group modify: gid %s already exists", g.GID), status.Messages()[0])
						assert.True(t, status.HasChanges())
					} else {
						assert.EqualError(t, err, "group: not supported on this system")
					}
				})
			})
		})
	})

	t.Run("state=absent", func(t *testing.T) {
		g := group.NewGroup(new(group.System))
		g.State = group.StateAbsent

		t.Run("gid not provided", func(t *testing.T) {
			t.Run("no delete-group does not exist", func(t *testing.T) {
				g.Name = fakeName
				status, err := g.Check(context.Background(), fakerenderer.New())

				if runtime.GOOS == "linux" {
					assert.NoError(t, err)
					assert.Equal(t, resource.StatusNoChange, status.StatusCode())
					assert.Equal(t, fmt.Sprintf("group delete: group %s does not exist", g.Name), status.Messages()[0])
					assert.False(t, status.HasChanges())
				} else {
					assert.EqualError(t, err, "group: not supported on this system")
				}
			})

			t.Run("delete group", func(t *testing.T) {
				g.Name = currName
				status, err := g.Check(context.Background(), fakerenderer.New())

				if runtime.GOOS == "linux" {
					assert.NoError(t, err)
					assert.Equal(t, resource.StatusWillChange, status.StatusCode())
					assert.Equal(t, fmt.Sprintf("group %s", g.Name), status.Diffs()["group"].Original())
					assert.Equal(t, string(group.StateAbsent), status.Diffs()["group"].Current())
					assert.True(t, status.HasChanges())
				} else {
					assert.EqualError(t, err, "group: not supported on this system")
				}
			})
		})

		t.Run("gid provided", func(t *testing.T) {
			t.Run("no delete-group name and gid do not exist", func(t *testing.T) {
				g.GID = fakeGid
				g.Name = fakeName
				status, err := g.Check(context.Background(), fakerenderer.New())

				if runtime.GOOS == "linux" {
					assert.NoError(t, err)
					assert.Equal(t, resource.StatusNoChange, status.StatusCode())
					assert.Equal(t, fmt.Sprintf("group delete: group %s and gid %s do not exist", g.Name, g.GID), status.Messages()[0])
					assert.False(t, status.HasChanges())
				} else {
					assert.EqualError(t, err, "group: not supported on this system")
				}
			})

			t.Run("no delete-group name does not exist", func(t *testing.T) {
				g.GID = currGid
				g.Name = fakeName
				status, err := g.Check(context.Background(), fakerenderer.New())

				if runtime.GOOS == "linux" {
					assert.EqualError(t, err, "cannot delete group")
					assert.Equal(t, resource.StatusCantChange, status.StatusCode())
					assert.Equal(t, fmt.Sprintf("group delete: group %s does not exist", g.Name), status.Messages()[0])
					assert.True(t, status.HasChanges())
				} else {
					assert.EqualError(t, err, "group: not supported on this system")
				}
			})

			t.Run("no delete-group gid does not exist", func(t *testing.T) {
				g.GID = fakeGid
				g.Name = currName
				status, err := g.Check(context.Background(), fakerenderer.New())

				if runtime.GOOS == "linux" {
					assert.EqualError(t, err, "cannot delete group")
					assert.Equal(t, resource.StatusCantChange, status.StatusCode())
					assert.Equal(t, fmt.Sprintf("group delete: gid %s does not exist", g.GID), status.Messages()[0])
					assert.True(t, status.HasChanges())
				} else {
					assert.EqualError(t, err, "group: not supported on this system")
				}
			})

			t.Run("no delete-group name and gid belong to different groups", func(t *testing.T) {
				gid, err := setGid()
				if err != nil {
					panic(err)
				}
				g.GID = gid
				g.Name = currName
				status, err := g.Check(context.Background(), fakerenderer.New())

				if runtime.GOOS == "linux" {
					assert.EqualError(t, err, "cannot delete group")
					assert.Equal(t, resource.StatusCantChange, status.StatusCode())
					assert.Equal(t, fmt.Sprintf("group delete: group %s and gid %s belong to different groups", g.Name, g.GID), status.Messages()[0])
					assert.True(t, status.HasChanges())
				} else {
					assert.EqualError(t, err, "group: not supported on this system")
				}
			})

			t.Run("delete group with gid", func(t *testing.T) {
				g.GID = currGid
				g.Name = currName
				status, err := g.Check(context.Background(), fakerenderer.New())

				if runtime.GOOS == "linux" {
					assert.NoError(t, err)
					assert.Equal(t, resource.StatusWillChange, status.StatusCode())
					assert.Equal(t, fmt.Sprintf("group %s with gid %s", g.Name, g.GID), status.Diffs()["group"].Original())
					assert.Equal(t, string(group.StateAbsent), status.Diffs()["group"].Current())
					assert.True(t, status.HasChanges())
				} else {
					assert.EqualError(t, err, "group: not supported on this system")
				}
			})
		})
	})

	t.Run("state unknown", func(t *testing.T) {
		g := group.NewGroup(new(group.System))
		g.GID = fakeGid
		g.Name = fakeName
		g.State = "test"
		_, err := g.Check(context.Background(), fakerenderer.New())

		if runtime.GOOS == "linux" {
			assert.EqualError(t, err, fmt.Sprintf("group: unrecognized state %s", g.State))
		} else {
			assert.EqualError(t, err, "group: not supported on this system")
		}
	})
}

// TestApply tests all possible cases Apply handles
func TestApply(t *testing.T) {
	t.Parallel()

	t.Run("state=present", func(t *testing.T) {
		t.Run("gid not provided", func(t *testing.T) {
			t.Run("NewName not provided", func(t *testing.T) {
				t.Run("add group", func(t *testing.T) {
					grp := &user.Group{
						Name: fakeName,
						Gid:  fakeGid,
					}
					m := &MockSystem{}
					g := group.NewGroup(m)
					g.Name = grp.Name
					g.State = group.StatePresent

					m.On("LookupGroup", g.Name).Return(new(user.Group), user.UnknownGroupError(""))
					m.On("AddGroup", g.Name, g.GID).Return(nil)
					status, err := g.Apply(context.Background())

					m.AssertCalled(t, "AddGroup", g.Name, g.GID)
					assert.NoError(t, err)
					assert.Equal(t, fmt.Sprintf("added group %s", g.Name), status.Messages()[0])
				})

				t.Run("no add-error adding group", func(t *testing.T) {
					grp := &user.Group{
						Name: fakeName,
						Gid:  fakeGid,
					}
					m := &MockSystem{}
					g := group.NewGroup(m)
					g.Name = grp.Name
					g.State = group.StatePresent

					m.On("LookupGroup", g.Name).Return(new(user.Group), user.UnknownGroupError(""))
					m.On("AddGroup", g.Name, g.GID).Return(fmt.Errorf(""))
					status, err := g.Apply(context.Background())

					m.AssertCalled(t, "AddGroup", g.Name, g.GID)
					assert.EqualError(t, err, "group add: ")
					assert.Equal(t, resource.StatusFatal, status.StatusCode())
					assert.Equal(t, fmt.Sprintf("error adding group %s", g.Name), status.Messages()[0])
				})

				t.Run("no add-will not attempt add", func(t *testing.T) {
					grp := &user.Group{
						Name: fakeName,
						Gid:  fakeGid,
					}
					m := &MockSystem{}
					g := group.NewGroup(m)
					g.Name = grp.Name
					g.State = group.StatePresent

					m.On("LookupGroup", g.Name).Return(grp, nil)
					m.On("AddGroup", g.Name, g.GID).Return(nil)
					status, err := g.Apply(context.Background())

					m.AssertNotCalled(t, "AddGroup", g.Name, g.GID)
					assert.EqualError(t, err, fmt.Sprintf("will not attempt add: group %s", g.Name))
					assert.Equal(t, resource.StatusCantChange, status.StatusCode())
				})
			})

			t.Run("NewName provided", func(t *testing.T) {
				t.Run("modify group name", func(t *testing.T) {
					grp := &user.Group{
						Name: currName,
						Gid:  currGid,
					}
					m := &MockSystem{}
					g := group.NewGroup(m)
					g.Name = grp.Name
					g.NewName = fakeName
					g.State = group.StatePresent
					options := group.ModGroupOptions{NewName: g.NewName}

					m.On("LookupGroup", g.Name).Return(grp, nil)
					m.On("LookupGroup", g.NewName).Return(new(user.Group), user.UnknownGroupError(""))
					m.On("ModGroup", g.Name, &options).Return(nil)
					status, err := g.Apply(context.Background())

					m.AssertCalled(t, "ModGroup", g.Name, &options)
					assert.NoError(t, err)
					assert.Equal(t, fmt.Sprintf("modified group %s with new name %s", g.Name, g.NewName), status.Messages()[0])
				})

				t.Run("no modify-error modifying group", func(t *testing.T) {
					grp := &user.Group{
						Name: currName,
						Gid:  currGid,
					}
					m := &MockSystem{}
					g := group.NewGroup(m)
					g.Name = grp.Name
					g.NewName = fakeName
					g.State = group.StatePresent
					options := group.ModGroupOptions{NewName: g.NewName}

					m.On("LookupGroup", g.Name).Return(grp, nil)
					m.On("LookupGroup", g.NewName).Return(new(user.Group), user.UnknownGroupError(""))
					m.On("ModGroup", g.Name, &options).Return(fmt.Errorf(""))
					status, err := g.Apply(context.Background())

					m.AssertCalled(t, "ModGroup", g.Name, &options)
					assert.EqualError(t, err, "group modify: ")
					assert.Equal(t, resource.StatusFatal, status.StatusCode())
					assert.Equal(t, fmt.Sprintf("error modifying group %s", g.Name), status.Messages()[0])
				})

				t.Run("no modify-will not attempt modify", func(t *testing.T) {
					grp := &user.Group{
						Name: currName,
						Gid:  currGid,
					}
					m := &MockSystem{}
					g := group.NewGroup(m)
					g.Name = grp.Name
					g.NewName = fakeName
					g.State = group.StatePresent
					options := group.ModGroupOptions{NewName: g.NewName}

					m.On("LookupGroup", g.Name).Return(grp, nil)
					m.On("LookupGroup", g.NewName).Return(grp, nil)
					m.On("ModGroup", g.Name, &options).Return(nil)
					status, err := g.Apply(context.Background())

					m.AssertNotCalled(t, "ModGroup", g.Name, &options)
					assert.EqualError(t, err, fmt.Sprintf("will not attempt modify: group %s", g.Name))
					assert.Equal(t, resource.StatusCantChange, status.StatusCode())
				})
			})
		})

		t.Run("gid provided", func(t *testing.T) {
			t.Run("NewName not provided", func(t *testing.T) {
				t.Run("add group with gid", func(t *testing.T) {
					grp := &user.Group{
						Name: fakeName,
						Gid:  fakeGid,
					}
					m := &MockSystem{}
					g := group.NewGroup(m)
					g.GID = grp.Gid
					g.Name = grp.Name
					g.State = group.StatePresent

					m.On("LookupGroup", g.Name).Return(new(user.Group), user.UnknownGroupError(""))
					m.On("LookupGroupID", g.GID).Return(new(user.Group), user.UnknownGroupIdError(""))
					m.On("AddGroup", g.Name, g.GID).Return(nil)
					status, err := g.Apply(context.Background())

					m.AssertCalled(t, "AddGroup", g.Name, g.GID)
					assert.NoError(t, err)
					assert.Equal(t, fmt.Sprintf("added group %s with gid %s", g.Name, g.GID), status.Messages()[0])
				})

				t.Run("no add-error adding group", func(t *testing.T) {
					grp := &user.Group{
						Name: fakeName,
						Gid:  fakeGid,
					}
					m := &MockSystem{}
					g := group.NewGroup(m)
					g.GID = grp.Gid
					g.Name = grp.Name
					g.State = group.StatePresent

					m.On("LookupGroup", g.Name).Return(new(user.Group), user.UnknownGroupError(""))
					m.On("LookupGroupID", g.GID).Return(new(user.Group), user.UnknownGroupIdError(""))
					m.On("AddGroup", g.Name, g.GID).Return(fmt.Errorf(""))
					status, err := g.Apply(context.Background())

					m.AssertCalled(t, "AddGroup", g.Name, g.GID)
					assert.EqualError(t, err, "group add: ")
					assert.Equal(t, resource.StatusFatal, status.StatusCode())
					assert.Equal(t, fmt.Sprintf("error adding group %s with gid %s", g.Name, g.GID), status.Messages()[0])
				})

				t.Run("no add-will not attempt add/modify", func(t *testing.T) {
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
					status, err := g.Apply(context.Background())

					m.AssertNotCalled(t, "AddGroup", g.Name, g.GID)
					assert.EqualError(t, err, fmt.Sprintf("will not attempt add/modify: group %s with gid %s", g.Name, g.GID))
					assert.Equal(t, resource.StatusCantChange, status.StatusCode())
				})

				t.Run("modify group gid", func(t *testing.T) {
					grp := &user.Group{
						Name: currName,
						Gid:  fakeGid,
					}
					m := &MockSystem{}
					g := group.NewGroup(m)
					g.GID = grp.Gid
					g.Name = grp.Name
					g.State = group.StatePresent
					options := group.ModGroupOptions{GID: g.GID}

					m.On("LookupGroup", g.Name).Return(grp, nil)
					m.On("LookupGroupID", g.GID).Return(new(user.Group), user.UnknownGroupIdError(""))
					m.On("ModGroup", g.Name, &options).Return(nil)
					status, err := g.Apply(context.Background())

					m.AssertCalled(t, "ModGroup", g.Name, &options)
					assert.NoError(t, err)
					assert.Equal(t, fmt.Sprintf("modified group %s with new gid %s", g.Name, g.GID), status.Messages()[0])
				})

				t.Run("no modify-error modifying group gid", func(t *testing.T) {
					grp := &user.Group{
						Name: currName,
						Gid:  fakeGid,
					}
					m := &MockSystem{}
					g := group.NewGroup(m)
					g.GID = grp.Gid
					g.Name = grp.Name
					g.State = group.StatePresent
					options := group.ModGroupOptions{GID: g.GID}

					m.On("LookupGroup", g.Name).Return(grp, nil)
					m.On("LookupGroupID", g.GID).Return(new(user.Group), user.UnknownGroupIdError(""))
					m.On("ModGroup", g.Name, &options).Return(fmt.Errorf(""))
					status, err := g.Apply(context.Background())

					m.AssertCalled(t, "ModGroup", g.Name, &options)
					assert.EqualError(t, err, "group modify: ")
					assert.Equal(t, resource.StatusFatal, status.StatusCode())
					assert.Equal(t, fmt.Sprintf("error modifying group %s with new gid %s", g.Name, g.GID), status.Messages()[0])
				})

				t.Run("no modify-will not attempt add/modify", func(t *testing.T) {
					grp := &user.Group{
						Name: currName,
						Gid:  fakeGid,
					}
					m := &MockSystem{}
					g := group.NewGroup(m)
					g.GID = grp.Gid
					g.Name = grp.Name
					g.State = group.StatePresent

					m.On("LookupGroup", g.Name).Return(grp, nil)
					m.On("LookupGroupID", g.GID).Return(grp, nil)
					m.On("ModGroup", g.Name, g.GID).Return(nil)
					status, err := g.Apply(context.Background())

					m.AssertNotCalled(t, "ModGroup", g.Name, g.GID)
					assert.EqualError(t, err, fmt.Sprintf("will not attempt add/modify: group %s with gid %s", g.Name, g.GID))
					assert.Equal(t, resource.StatusCantChange, status.StatusCode())
				})
			})

			t.Run("NewName provided", func(t *testing.T) {
				t.Run("modify group name and gid", func(t *testing.T) {
					grp := &user.Group{
						Name: currName,
						Gid:  currGid,
					}
					m := &MockSystem{}
					g := group.NewGroup(m)
					g.Name = grp.Name
					g.NewName = fakeName
					g.GID = fakeGid
					g.State = group.StatePresent
					options := group.ModGroupOptions{NewName: g.NewName, GID: g.GID}

					m.On("LookupGroup", g.Name).Return(grp, nil)
					m.On("LookupGroup", g.NewName).Return(new(user.Group), user.UnknownGroupError(""))
					m.On("LookupGroupID", g.GID).Return(new(user.Group), user.UnknownGroupIdError(""))
					m.On("ModGroup", g.Name, &options).Return(nil)
					status, err := g.Apply(context.Background())

					m.AssertCalled(t, "ModGroup", g.Name, &options)
					assert.NoError(t, err)
					assert.Equal(t, fmt.Sprintf("modified group %s with new name %s and new gid %s", g.Name, g.NewName, g.GID), status.Messages()[0])
				})

				t.Run("no modify-error modifying group", func(t *testing.T) {
					grp := &user.Group{
						Name: currName,
						Gid:  currGid,
					}
					m := &MockSystem{}
					g := group.NewGroup(m)
					g.Name = grp.Name
					g.NewName = fakeName
					g.GID = fakeGid
					g.State = group.StatePresent
					options := group.ModGroupOptions{NewName: g.NewName, GID: g.GID}

					m.On("LookupGroup", g.Name).Return(grp, nil)
					m.On("LookupGroup", g.NewName).Return(new(user.Group), user.UnknownGroupError(""))
					m.On("LookupGroupID", g.GID).Return(new(user.Group), user.UnknownGroupIdError(""))
					m.On("ModGroup", g.Name, &options).Return(fmt.Errorf(""))
					status, err := g.Apply(context.Background())

					m.AssertCalled(t, "ModGroup", g.Name, &options)
					assert.EqualError(t, err, "group modify: ")
					assert.Equal(t, resource.StatusFatal, status.StatusCode())
					assert.Equal(t, fmt.Sprintf("error modifying group %s with new name %s and new gid %s", g.Name, g.NewName, g.GID), status.Messages()[0])
				})

				t.Run("no modify-will not attempt modify", func(t *testing.T) {
					grp := &user.Group{
						Name: currName,
						Gid:  currGid,
					}
					m := &MockSystem{}
					g := group.NewGroup(m)
					g.Name = grp.Name
					g.NewName = fakeName
					g.GID = fakeGid
					g.State = group.StatePresent
					options := group.ModGroupOptions{NewName: g.NewName, GID: g.GID}

					m.On("LookupGroup", g.Name).Return(grp, nil)
					m.On("LookupGroup", g.NewName).Return(grp, nil)
					m.On("LookupGroupID", g.GID).Return(grp, nil)
					m.On("ModGroup", g.Name, &options).Return(nil)
					status, err := g.Apply(context.Background())

					m.AssertNotCalled(t, "ModGroup", g.Name, &options)
					assert.EqualError(t, err, fmt.Sprintf("will not attempt modify: group %s with new name %s and new gid %s", g.Name, g.NewName, g.GID))
					assert.Equal(t, resource.StatusCantChange, status.StatusCode())
				})
			})
		})
	})

	t.Run("state=absent", func(t *testing.T) {
		t.Run("gid not provided", func(t *testing.T) {
			t.Run("delete group", func(t *testing.T) {
				grp := &user.Group{
					Name: fakeName,
					Gid:  fakeGid,
				}
				m := &MockSystem{}
				g := group.NewGroup(m)
				g.Name = grp.Name
				g.State = group.StateAbsent

				m.On("LookupGroup", g.Name).Return(grp, nil)
				m.On("DelGroup", g.Name).Return(nil)
				status, err := g.Apply(context.Background())

				m.AssertCalled(t, "DelGroup", g.Name)
				assert.NoError(t, err)
				assert.Equal(t, fmt.Sprintf("deleted group %s", g.Name), status.Messages()[0])
			})

			t.Run("no delete-error deleting group", func(t *testing.T) {
				grp := &user.Group{
					Name: fakeName,
					Gid:  fakeGid,
				}
				m := &MockSystem{}
				g := group.NewGroup(m)
				g.Name = grp.Name
				g.State = group.StateAbsent

				m.On("LookupGroup", g.Name).Return(grp, nil)
				m.On("DelGroup", g.Name).Return(fmt.Errorf(""))
				status, err := g.Apply(context.Background())

				m.AssertCalled(t, "DelGroup", g.Name)
				assert.EqualError(t, err, "group delete: ")
				assert.Equal(t, resource.StatusFatal, status.StatusCode())
				assert.Equal(t, fmt.Sprintf("error deleting group %s", g.Name), status.Messages()[0])
			})

			t.Run("no delete-will not attempt delete", func(t *testing.T) {
				grp := &user.Group{
					Name: fakeName,
					Gid:  fakeGid,
				}
				m := &MockSystem{}
				g := group.NewGroup(m)
				g.Name = grp.Name
				g.State = group.StateAbsent

				m.On("LookupGroup", g.Name).Return(new(user.Group), user.UnknownGroupError(""))
				m.On("DelGroup", g.Name).Return(nil)
				status, err := g.Apply(context.Background())

				m.AssertNotCalled(t, "DelGroup", g.Name)
				assert.EqualError(t, err, fmt.Sprintf("will not attempt delete: group %s", g.Name))
				assert.Equal(t, resource.StatusCantChange, status.StatusCode())
			})
		})

		t.Run("gid provided", func(t *testing.T) {
			t.Run("delete group with gid", func(t *testing.T) {
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
				status, err := g.Apply(context.Background())

				m.AssertCalled(t, "DelGroup", g.Name)
				assert.NoError(t, err)
				assert.Equal(t, fmt.Sprintf("deleted group %s with gid %s", g.Name, g.GID), status.Messages()[0])
			})

			t.Run("no delete-error deleting group", func(t *testing.T) {
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
				status, err := g.Apply(context.Background())

				m.AssertCalled(t, "DelGroup", g.Name)
				assert.EqualError(t, err, "group delete: ")
				assert.Equal(t, resource.StatusFatal, status.StatusCode())
				assert.Equal(t, fmt.Sprintf("error deleting group %s with gid %s", g.Name, g.GID), status.Messages()[0])
			})

			t.Run("no delete-will not attempt delete", func(t *testing.T) {
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
				status, err := g.Apply(context.Background())

				m.AssertNotCalled(t, "DelGroup", g.Name)
				assert.EqualError(t, err, fmt.Sprintf("will not attempt delete: group %s with gid %s", g.Name, g.GID))
				assert.Equal(t, resource.StatusCantChange, status.StatusCode())
			})
		})
	})

	t.Run("state unknown", func(t *testing.T) {
		m := &MockSystem{}
		g := group.NewGroup(m)
		grp := &user.Group{
			Name: fakeName,
			Gid:  fakeGid,
		}
		g.GID = grp.Gid
		g.Name = grp.Name
		g.State = "test"

		m.On("LookupGroup", g.Name).Return(grp, nil)
		m.On("LookupGroupID", g.GID).Return(grp, nil)
		m.On("AddGroup", g.Name, g.GID)
		m.On("DelGroup", g.Name)
		_, err := g.Apply(context.Background())

		m.AssertNotCalled(t, "AddGroup", g.Name, g.GID)
		m.AssertNotCalled(t, "DelGroup", g.Name)
		assert.EqualError(t, err, fmt.Sprintf("group: unrecognized state %s", g.State))
	})
}

// TestSetModGroupOptions tests options provided for modifying a group
// are properly set
func TestSetModGroupOptions(t *testing.T) {
	t.Parallel()

	t.Run("all options", func(t *testing.T) {
		g := group.NewGroup(new(group.System))
		g.Name = currName
		g.NewName = fakeName
		g.GID = fakeGid

		options := group.SetModGroupOptions(g)

		assert.Equal(t, g.NewName, options.NewName)
		assert.Equal(t, g.GID, options.GID)
	})

	t.Run("no options", func(t *testing.T) {
		g := group.NewGroup(new(group.System))
		g.Name = currName

		options := group.SetModGroupOptions(g)

		assert.Equal(t, "", options.NewName)
		assert.Equal(t, "", options.GID)
	})
}

// setGid is used to set a gid that exists but is not a match for
// the current user group name (currName).
func setGid() (string, error) {
	for i := 0; i <= maxGID; i++ {
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
	for i := minGID; i <= maxGID; i++ {
		gid := strconv.Itoa(i)
		_, err := user.LookupGroupId(gid)
		if err != nil {
			return gid, nil
		}
	}
	return "", fmt.Errorf("setFakeGid: could not set gid")
}

// MockSystem for Group
type MockSystem struct {
	mock.Mock
}

// AddGroup for MockSystem
func (m *MockSystem) AddGroup(name, gid string) error {
	args := m.Called(name, gid)
	return args.Error(0)
}

// DelGroup for MockSystem
func (m *MockSystem) DelGroup(name string) error {
	args := m.Called(name)
	return args.Error(0)
}

// ModGroup for MockSystem
func (m *MockSystem) ModGroup(name string, options *group.ModGroupOptions) error {
	args := m.Called(name, options)
	return args.Error(0)
}

// LookupGroup for MockSystem
func (m *MockSystem) LookupGroup(name string) (*user.Group, error) {
	args := m.Called(name)
	return args.Get(0).(*user.Group), args.Error(1)
}

// LookupGroupID for MockSystem
func (m *MockSystem) LookupGroupID(gid string) (*user.Group, error) {
	args := m.Called(gid)
	return args.Get(0).(*user.Group), args.Error(1)
}
