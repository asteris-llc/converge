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
	"testing"

	"github.com/asteris-llc/converge/helpers/fakerenderer"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/group"
	"github.com/stretchr/testify/assert"
)

// TestPreparerInterface tests that the Preparer interface is properly implemeted
func TestPreparerInterface(t *testing.T) {
	t.Parallel()

	assert.Implements(t, (*resource.Resource)(nil), new(group.Preparer))
}

// TestPrepare tests the valid and invalid cases of Prepare
func TestPrepare(t *testing.T) {
	t.Parallel()

	fr := fakerenderer.FakeRenderer{}
	var invalidGID = uint32(math.MaxUint32)
	var maxGID = uint32(math.MaxUint32 - 1)
	var minGID uint32
	var testGID uint32 = 123

	t.Run("valid", func(t *testing.T) {
		t.Run("all parameters", func(t *testing.T) {
			p := group.Preparer{GID: &testGID, Name: "test", NewName: "test2", State: group.StateAbsent}
			_, err := p.Prepare(&fr)

			assert.NoError(t, err)
		})

		t.Run("no new_name parameter", func(t *testing.T) {
			p := group.Preparer{GID: &testGID, Name: "test", State: group.StatePresent}
			_, err := p.Prepare(&fr)

			assert.NoError(t, err)
		})

		t.Run("no state parameter", func(t *testing.T) {
			p := group.Preparer{GID: &testGID, Name: "test", NewName: "test2"}
			_, err := p.Prepare(&fr)

			assert.NoError(t, err)
		})

		t.Run("no gid parameter", func(t *testing.T) {
			p := group.Preparer{Name: "test", NewName: "test2", State: group.StateAbsent}
			_, err := p.Prepare(&fr)

			assert.NoError(t, err)
		})

		t.Run("min allowable gid", func(t *testing.T) {
			p := group.Preparer{GID: &minGID, Name: "test"}
			_, err := p.Prepare(&fr)

			assert.NoError(t, err)
		})

		t.Run("max allowable gid", func(t *testing.T) {
			p := group.Preparer{GID: &maxGID, Name: "test"}
			_, err := p.Prepare(&fr)

			assert.NoError(t, err)
		})
	})

	t.Run("invalid", func(t *testing.T) {
		t.Run("gid out of range", func(t *testing.T) {
			p := group.Preparer{GID: &invalidGID, Name: "test"}
			_, err := p.Prepare(&fr)

			assert.EqualError(t, err, fmt.Sprintf("group \"gid\" parameter out of range"))
		})
	})
}
