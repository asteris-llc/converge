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
	"testing"

	"github.com/asteris-llc/converge/helpers/fakerenderer"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/user"
	"github.com/stretchr/testify/assert"
)

// TestPreparerInterface tests that the Preparer interface is properly implemeted
func TestPreparerInterface(t *testing.T) {
	t.Parallel()

	assert.Implements(t, (*resource.Resource)(nil), new(user.Preparer))
}

// TestPrepare tests the valid and invalid cases of Prepare
func TestPrepare(t *testing.T) {
	t.Parallel()

	fr := fakerenderer.FakeRenderer{}
	var invalidID = uint32(math.MaxUint32)
	var maxID = uint32(math.MaxUint32 - 1)
	var minID uint32
	var testID uint32 = 123

	t.Run("valid", func(t *testing.T) {
		t.Run("no uid", func(t *testing.T) {
			p := user.Preparer{GID: &testID, Username: "test", Name: "test", HomeDir: "tmp", State: user.StateAbsent}
			_, err := p.Prepare(&fr)

			assert.NoError(t, err)
		})

		t.Run("min allowable uid", func(t *testing.T) {
			p := user.Preparer{UID: &minID, Username: "test"}
			_, err := p.Prepare(&fr)

			assert.NoError(t, err)
		})

		t.Run("max allowable uid", func(t *testing.T) {
			p := user.Preparer{UID: &maxID, Username: "test"}
			_, err := p.Prepare(&fr)

			assert.NoError(t, err)
		})

		t.Run("no group", func(t *testing.T) {
			p := user.Preparer{UID: &testID, Username: "test", Name: "test", HomeDir: "tmp", State: user.StateAbsent}
			_, err := p.Prepare(&fr)

			assert.NoError(t, err)
		})

		t.Run("with groupname", func(t *testing.T) {
			p := user.Preparer{UID: &testID, Username: "test", GroupName: currGroupName, Name: "test", HomeDir: "tmp", State: user.StateAbsent}
			_, err := p.Prepare(&fr)

			assert.NoError(t, err)
		})

		t.Run("with gid", func(t *testing.T) {
			p := user.Preparer{UID: &testID, Username: "test", GID: &testID, Name: "test", HomeDir: "tmp", State: user.StateAbsent}
			_, err := p.Prepare(&fr)

			assert.NoError(t, err)
		})

		t.Run("min allowable gid", func(t *testing.T) {
			p := user.Preparer{GID: &minID, Username: "test"}
			_, err := p.Prepare(&fr)

			assert.NoError(t, err)
		})

		t.Run("max allowable gid", func(t *testing.T) {
			p := user.Preparer{GID: &maxID, Username: "test"}
			_, err := p.Prepare(&fr)

			assert.NoError(t, err)
		})

		t.Run("no name parameter", func(t *testing.T) {
			p := user.Preparer{UID: &testID, GID: &testID, Username: "test", HomeDir: "tmp", State: user.StateAbsent}
			_, err := p.Prepare(&fr)

			assert.NoError(t, err)
		})

		t.Run("no home_dir parameter", func(t *testing.T) {
			p := user.Preparer{UID: &testID, GID: &testID, Username: "test", Name: "test", State: user.StateAbsent}
			_, err := p.Prepare(&fr)

			assert.NoError(t, err)
		})

		t.Run("no state parameter", func(t *testing.T) {
			p := user.Preparer{UID: &testID, GID: &testID, Username: "test", Name: "test", HomeDir: "tmp"}
			_, err := p.Prepare(&fr)

			assert.NoError(t, err)
		})

		t.Run("home_dir and move_dir parameters", func(t *testing.T) {
			p := user.Preparer{Username: "test", MoveDir: true, HomeDir: "tmp"}
			_, err := p.Prepare(&fr)

			assert.NoError(t, err)
		})
	})

	t.Run("invalid", func(t *testing.T) {
		t.Run("uid out of range", func(t *testing.T) {
			p := user.Preparer{UID: &invalidID, Username: "test"}
			_, err := p.Prepare(&fr)

			assert.EqualError(t, err, fmt.Sprintf("user \"uid\" parameter out of range"))
		})

		t.Run("gid out of range", func(t *testing.T) {
			p := user.Preparer{GID: &invalidID, Username: "test"}
			_, err := p.Prepare(&fr)

			assert.EqualError(t, err, fmt.Sprintf("user \"gid\" parameter out of range"))
		})

		t.Run("no home_dir with move_dir", func(t *testing.T) {
			p := user.Preparer{Username: "test", MoveDir: true}
			_, err := p.Prepare(&fr)

			assert.EqualError(t, err, fmt.Sprintf("user \"home_dir\" parameter required with \"move_dir\" parameter"))
		})
	})
}
