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

package owner_test

import (
	"os/user"
	"testing"

	"github.com/asteris-llc/converge/resource/file/owner"
	"github.com/stretchr/testify/assert"
)

// TestOwnershipDiff tests things related to the owernship diff
func TestOwernshipDiff(t *testing.T) {

	users := []*user.User{
		fakeUser("1", "1", "user-1"),
		fakeUser("2", "2", "user-2"),
		fakeUser("3", "3", "user-3"),
	}
	groups := []*user.Group{
		fakeGroup("1", "group-1"),
		fakeGroup("2", "group-2"),
		fakeGroup("3", "group-3"),
	}
	m := newMockOS(nil, users, groups, nil, nil)

	t.Run("Original", func(t *testing.T) {

		t.Run("uid", func(t *testing.T) {
			o := (&owner.OwnershipDiff{UIDs: &[2]int{1, 2}}).SetProxy(m)
			assert.Equal(t, "user: user-1 (1)", o.Original())
		})
		t.Run("gid", func(t *testing.T) {
		})
		t.Run("both", func(t *testing.T) {
		})
	})

	t.Run("Current", func(t *testing.T) {
	})

	t.Run("Changes", func(t *testing.T) {
	})
}
