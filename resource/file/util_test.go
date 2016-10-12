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
	"errors"
	"os"
	"os/user"
	"runtime"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestModeType checks to see if type bits are set correctly
func TestModeType(t *testing.T) {
	t.Run("modeType", func(t *testing.T) {
		var tests = []struct {
			mode     uint32
			filetype Type
			expected string
		}{
			{uint32(0750), TypeDirectory, "drwxr-x---"},
			{uint32(0557), TypeFile, "-r-xr-xrwx"},
			{uint32(0440), TypeSymlink, "Lr--r-----"},
		}

		for _, tt := range tests {
			m := ModeType(tt.mode, tt.filetype)
			assert.Equal(t, tt.expected, os.FileMode(m).String())
		}
	})
}

// TestDesiredGroup checks to see that the proper group is returned
func TestDesiredGroup(t *testing.T) {
	effectiveGroup, err := user.LookupGroupId(strconv.Itoa(os.Getegid()))
	require.NoError(t, err)

	type Case struct {
		f        *user.Group
		actual   *user.Group
		expected *user.Group
		changed  bool
		err      error
	}

	var groups []Case

	switch runtime.GOOS {
	case "darwin":
		groups = []Case{
			{&user.Group{Name: "wheel", Gid: "0"}, &user.Group{}, &user.Group{Name: "wheel", Gid: "0"}, true, nil},
			{&user.Group{Name: "wheel", Gid: "0"}, &user.Group{Name: "wheel", Gid: "0"}, &user.Group{Name: "wheel", Gid: "0"}, false, nil},
			{&user.Group{}, &user.Group{Name: "wheel", Gid: "0"}, &user.Group{Name: "wheel", Gid: "0"}, false, nil},
			{&user.Group{}, &user.Group{}, effectiveGroup, true, nil},
			{&user.Group{Name: "converge-bad-group"}, &user.Group{}, effectiveGroup, true, errors.New("unable to get information: group: unknown group converge-bad-group")},
		}
	case "linux":
		groups = []Case{
			{&user.Group{Name: "root", Gid: "0"}, &user.Group{}, &user.Group{Name: "root", Gid: "0"}, true, nil},
			{&user.Group{Name: "root", Gid: "0"}, &user.Group{Name: "root", Gid: "0"}, &user.Group{Name: "root", Gid: "0"}, false, nil},
			{&user.Group{}, &user.Group{Name: "root", Gid: "0"}, &user.Group{Name: "root", Gid: "0"}, false, nil},
			{&user.Group{}, &user.Group{}, effectiveGroup, true, nil},
			{&user.Group{Name: "converge-bad-group"}, &user.Group{}, effectiveGroup, true, errors.New("unable to get information: group: unknown group converge-bad-group")},
		}
	default:
		groups = []Case{}
	}

	for i, tt := range groups {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			group, changed, err := desiredGroup(tt.f, tt.actual)
			switch tt.err {
			case nil:
				assert.Equal(t, tt.err, err)
				assert.Equal(t, tt.expected.Name, group.Name, "group names")
				assert.Equal(t, tt.expected.Gid, group.Gid, "group gid")
				assert.Equal(t, tt.changed, changed, "change status")
			default:
				assert.Equal(t, tt.err.Error(), err.Error(), "errors")
				assert.Equal(t, tt.changed, changed, "change status")
			}
		})
	}
}

// TestDesiredUser checks to see what user is returned
func TestDesiredUser(t *testing.T) {
	effectiveUser, err := user.LookupId(strconv.Itoa(os.Geteuid()))
	require.NoError(t, err)

	type Case struct {
		f        *user.User
		actual   *user.User
		expected *user.User
		changed  bool
		err      error
	}

	var users []Case
	switch runtime.GOOS {
	case "darwin", "linux":
		users = []Case{
			{&user.User{Username: "root", Uid: "0"}, &user.User{}, &user.User{Username: "root", Uid: "0"}, true, nil},
			{&user.User{Username: "root", Uid: "0"}, &user.User{Username: "root", Uid: "0"}, &user.User{Username: "root", Uid: "0"}, false, nil},
			{&user.User{}, &user.User{Username: "root", Uid: "0"}, &user.User{Username: "root", Uid: "0"}, false, nil},
			{&user.User{}, &user.User{}, effectiveUser, true, nil},
			{&user.User{Username: "converge-bad-user"}, &user.User{}, effectiveUser, true, errors.New("unable to get user information: user: unknown user converge-bad-user")},
		}
	default:
		users = []Case{}
	}

	for i, tt := range users {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			user, changed, err := desiredUser(tt.f, tt.actual)
			switch err {
			case nil:
				assert.Equal(t, tt.err, err)
				assert.Equal(t, tt.expected.Username, user.Username, "username mismatch")
				assert.Equal(t, tt.expected.Uid, user.Uid, "uid mismatch")
				assert.Equal(t, tt.changed, changed, "change state did not match")
			default:
				assert.Equal(t, tt.err.Error(), err.Error(), "error mismatch")
				assert.Equal(t, tt.changed, changed, "change state did not match")
			}
		})
	}
}
