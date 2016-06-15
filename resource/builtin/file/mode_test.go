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
	"testing"

	"github.com/asteris-llc/converge/helpers"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/builtin/file"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShellTaskInterfaces(t *testing.T) {
	t.Parallel()

	assert.Implements(t, (*resource.Resource)(nil), new(file.Mode))
	assert.Implements(t, (*fmt.Stringer)(nil), new(file.Mode))
	assert.Implements(t, (*resource.Monitor)(nil), new(file.Mode))
	assert.Implements(t, (*resource.Task)(nil), new(file.Mode))
}

func TestModeCheck(t *testing.T) {
	helpers.InTempDir(t, func() {
		err := ioutil.WriteFile("x", []byte{}, 0755)
		require.NoError(t, err)

		m := &file.Mode{
			RawDestination: "x",
			RawMode:        "0755",
		}
		require.NoError(t, m.Prepare(nil))

		status, willChange, err := m.Check()
		assert.NoError(t, err)
		assert.False(t, willChange)
		assert.Equal(t, "755", status)
	})
}

func TestModeCheckWillChange(t *testing.T) {
	helpers.InTempDir(t, func() {
		err := ioutil.WriteFile("x", []byte{}, 0755)
		require.NoError(t, err)

		m := &file.Mode{
			RawDestination: "x",
			RawMode:        "0644",
		}
		require.NoError(t, m.Prepare(nil))

		status, willChange, err := m.Check()
		assert.NoError(t, err)
		assert.True(t, willChange)
		assert.Equal(t, "755", status)
	})
}

func TestModeApply(t *testing.T) {
	helpers.InTempDir(t, func() {
		err := ioutil.WriteFile("x", []byte{}, 0755)

		m := &file.Mode{
			RawDestination: "x",
			RawMode:        "0644",
		}
		require.NoError(t, m.Prepare(nil))

		assert.NoError(t, m.Apply())

		status, willChange, err := m.Check()
		assert.NoError(t, err)
		assert.False(t, willChange)
		assert.Equal(t, "644", status)
	})
}
