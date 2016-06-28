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

package file_test

import (
	"fmt"
	"io/ioutil"
	"os/user"
	"testing"

	"github.com/asteris-llc/converge/helpers"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/builtin/file"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test that Owner impelements the required interfaces.
func TestTaskInterfacesForOwner(t *testing.T) {
	t.Parallel()

	//Should Implement Stringer Interface
	assert.Implements(t, (*fmt.Stringer)(nil), new(file.Owner))
	//Should Implement Monitor Interface
	assert.Implements(t, (*resource.Monitor)(nil), new(file.Owner))
	//Should Implement Resource Interface
	assert.Implements(t, (*resource.Resource)(nil), new(file.Owner))
	//Should Implement Task Interface
	assert.Implements(t, (*resource.Task)(nil), new(file.Owner))
}

// Test Preparation fails if there is no owner with that name.
func TestOwnerCheck(t *testing.T) {
	helpers.InTempDir(t, func() {
		err := ioutil.WriteFile("x", []byte{}, 0755)
		require.NoError(t, err)
		u, err := user.Current()
		require.NoError(t, err)
		o := &file.Owner{
			RawDestination: "x",
			RawOwner:       u.Username,
		}
		require.NoError(t, o.Prepare(nil))
		status, willChange, err := o.Check()
		assert.NoError(t, err)
		assert.False(t, willChange)
		assert.Equal(t, u.Username, status)
	})
}

// Test that user exist on system
func TestOwnerValidate(t *testing.T) {
	o := &file.Owner{
		RawDestination: "x",
		RawOwner:       "userwillnotexist",
	}
	err := o.Prepare(nil)
	require.Error(t, err)
	assert.EqualError(t, err, "user: unknown user userwillnotexist")

}

// TestApply Checks if the owner will change.
func TestOwnerApply(t *testing.T) {
	helpers.InTempDir(t, func() {
		err := ioutil.WriteFile("x", []byte{}, 0755)
		o := &file.Owner{
			RawDestination: "x",
			RawOwner:       "nobody",
		}
		require.NoError(t, err)
		require.NoError(t, o.Prepare(nil))
		o.Apply()
		status, willChange, err := o.Check()
		u, err := user.Current()
		require.NoError(t, err)
		//Should fail if not root
		if u.Uid != "0" {
			assert.NoError(t, err)
			assert.True(t, willChange)
			assert.Equal(t, u.Username, status)
		} else {
			assert.NoError(t, err)
			assert.False(t, willChange)
			assert.Equal(t, "nobody", status)
		}
	})
}
