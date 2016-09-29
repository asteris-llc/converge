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
	"strconv"
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

	t.Run("valid", func(t *testing.T) {
		t.Run("no uid", func(t *testing.T) {
			p := user.Preparer{GID: 123, Username: "test", Name: "test", HomeDir: "tmp", State: user.StateAbsent}
			_, err := p.Prepare(&fr)

			assert.NoError(t, err)
		})

		t.Run("max allowable uid", func(t *testing.T) {
			p := user.Preparer{UID: math.MaxUint32 - 1, Username: "test"}
			_, err := p.Prepare(&fr)

			assert.NoError(t, err)
		})

		t.Run("no group", func(t *testing.T) {
			p := user.Preparer{UID: 1234, Username: "test", Name: "test", HomeDir: "tmp", State: user.StateAbsent}
			_, err := p.Prepare(&fr)

			assert.NoError(t, err)
		})

		t.Run("with groupname", func(t *testing.T) {
			p := user.Preparer{UID: 1234, Username: "test", GroupName: currGroupName, Name: "test", HomeDir: "tmp", State: user.StateAbsent}
			_, err := p.Prepare(&fr)

			assert.NoError(t, err)
		})

		t.Run("with gid", func(t *testing.T) {
			gid, _ := strconv.ParseUint(currGID, 10, 32)
			p := user.Preparer{UID: 1234, Username: "test", GID: uint32(gid), Name: "test", HomeDir: "tmp", State: user.StateAbsent}
			_, err := p.Prepare(&fr)

			assert.NoError(t, err)
		})

		t.Run("max allowable gid", func(t *testing.T) {
			p := user.Preparer{GID: math.MaxUint32 - 1, Username: "test"}
			_, err := p.Prepare(&fr)

			assert.NoError(t, err)
		})

		t.Run("no name parameter", func(t *testing.T) {
			p := user.Preparer{UID: 1234, GID: 123, Username: "test", HomeDir: "tmp", State: user.StateAbsent}
			_, err := p.Prepare(&fr)

			assert.NoError(t, err)
		})

		t.Run("no home_dir parameter", func(t *testing.T) {
			p := user.Preparer{UID: 1234, GID: 123, Username: "test", Name: "test", State: user.StateAbsent}
			_, err := p.Prepare(&fr)

			assert.NoError(t, err)
		})

		t.Run("no state parameter", func(t *testing.T) {
			p := user.Preparer{UID: 1234, GID: 123, Username: "test", Name: "test", HomeDir: "tmp"}
			_, err := p.Prepare(&fr)

			assert.NoError(t, err)
		})
	})

	t.Run("invalid", func(t *testing.T) {
		t.Run("uid out of range", func(t *testing.T) {
			p := user.Preparer{UID: math.MaxUint32, Username: "test"}
			_, err := p.Prepare(&fr)

			assert.EqualError(t, err, fmt.Sprintf("user \"uid\" parameter out of range"))
		})

		t.Run("gid out of range", func(t *testing.T) {
			p := user.Preparer{GID: math.MaxUint32, Username: "test"}
			_, err := p.Prepare(&fr)

			assert.EqualError(t, err, fmt.Sprintf("user \"gid\" parameter out of range"))
		})
	})
}
