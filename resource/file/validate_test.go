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
	"fmt"
	"os/user"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateHelpers(t *testing.T) {
	t.Parallel()

	t.Run("stateValidators", func(t *testing.T) {
		var typeTests = []struct {
			F   File
			Err error
		}{
			{File{Destination: "/aster/is", State: "present"}, nil},
			{File{Destination: "/aster/is", State: "absent"}, nil},
			{File{Destination: "/aster/is", State: ""}, nil},
			{File{Destination: "/aster/is", State: "bad"}, fmt.Errorf("state should be one of %v, got %q", ValidStates, "bad")},
		}
		for _, tt := range typeTests {
			err := tt.F.Validate()
			assert.Equal(t, tt.Err, err)
		}
	})

	t.Run("typeValidators", func(t *testing.T) {
		var typeTests = []struct {
			F   File
			Err error
		}{
			{File{Destination: "/aster/is", Type: "directory"}, nil},
			{File{Destination: "/aster/is", Type: "file"}, nil},
			{File{Destination: "/aster/is", Type: "hardlink", Target: "/converge"}, nil},
			{File{Destination: "/aster/is", Type: "symlink", Target: "/converge"}, nil},
			{File{Destination: "/aster/is", Type: "bad"}, fmt.Errorf("type should be one of %v, got %q", AllTypes, "bad")},
		}
		for _, tt := range typeTests {
			err := tt.F.Validate()
			assert.Equal(t, tt.Err, err)
		}
	})

	// validate link targets
	t.Run("targetValidators", func(t *testing.T) {
		var perms = new(uint32)
		*perms = 511
		var targetTests = []struct {
			F   File
			Err error
		}{
			{File{Destination: "/aster/is", Type: "directory", Target: "/bad"}, fmt.Errorf("cannot define target on a type of %q: target: %q", "directory", "/bad")},
			{File{Destination: "/aster/is", Type: "file", Target: "/bad"}, fmt.Errorf("cannot define target on a type of %q: target: %q", "file", "/bad")},
			{File{Destination: "/aster/is", Type: "hardlink", Target: "/converge"}, nil},
			{File{Destination: "/aster/is", Type: "symlink", Target: "/converge"}, nil},
			{File{Destination: "/aster/is", Type: "hardlink"}, fmt.Errorf("must define a target if you are using a %q", "hardlink")},
			{File{Destination: "/aster/is", Type: "symlink"}, fmt.Errorf("must define a target if you are using a %q", "symlink")},
			{File{Destination: "/aster/is", Mode: perms, Target: "/test", Type: "symlink"}, fmt.Errorf("permissions on symlinks not supported")},
		}
		for _, tt := range targetTests {
			err := tt.F.Validate()
			assert.EqualValues(t, err, tt.Err)
		}
	})

	t.Run("groupValidators", func(t *testing.T) {
		validGroup := &user.Group{Name: "root", Gid: "0"}
		switch runtime.GOOS {
		case "darwin":
			validGroup.Name = "wheel"
		}
		var groupTests = []struct {
			F   File
			Err error
		}{
			{File{Destination: "/aster/is", GroupInfo: validGroup}, nil},
			{File{Destination: "/aster/is", GroupInfo: nil}, nil},
			{File{Destination: "/aster/is", GroupInfo: &user.Group{Name: "badGroup", Gid: "0"}}, fmt.Errorf("validation error: group: unknown group badGroup")},
		}
		for _, tt := range groupTests {
			err := tt.F.Validate()
			switch tt.Err {
			case nil:
				assert.Nil(t, err)
			default:
				assert.EqualValues(t, tt.Err.Error(), err.Error())
			}
		}
	})

	t.Run("userValidators", func(t *testing.T) {
		validUser := &user.User{Username: "root", Uid: "0"}

		var userTests = []struct {
			F   File
			Err error
		}{
			{File{Destination: "/aster/is", UserInfo: validUser}, nil},
			{File{Destination: "/aster/is", UserInfo: nil}, nil},
			{File{Destination: "/aster/is", UserInfo: &user.User{Username: "badUser", Uid: "4774"}}, fmt.Errorf("validation error: user: unknown user %s", "badUser")},
		}
		for _, tt := range userTests {
			err := tt.F.Validate()
			switch tt.Err {
			case nil:
				assert.Nil(t, err)
			default:
				assert.EqualValues(t, tt.Err.Error(), err.Error())
			}
		}
	})

}
