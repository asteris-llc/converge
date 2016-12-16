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
	"errors"
	"os/user"
	"testing"

	"github.com/asteris-llc/converge/resource/file/owner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			o := (&owner.OwnershipDiff{GIDs: &[2]int{1, 2}}).SetProxy(m)
			assert.Equal(t, "group: group-1 (1)", o.Original())
		})
		t.Run("both", func(t *testing.T) {
			o := (&owner.OwnershipDiff{UIDs: &[2]int{1, 2}, GIDs: &[2]int{1, 2}}).SetProxy(m)
			assert.Equal(t, "user: user-1 (1); group: group-1 (1)", o.Original())
		})
		t.Run("heterogenous", func(t *testing.T) {
			t.Run("mismatched-uid", func(t *testing.T) {
				o := (&owner.OwnershipDiff{UIDs: &[2]int{1, 2}, GIDs: &[2]int{1, 1}}).SetProxy(m)
				assert.Equal(t, "user: user-1 (1)", o.Original())
			})
			t.Run("mismatched-gid", func(t *testing.T) {
				o := (&owner.OwnershipDiff{UIDs: &[2]int{1, 1}, GIDs: &[2]int{1, 2}}).SetProxy(m)
				assert.Equal(t, "group: group-1 (1)", o.Original())
			})
		})
	})
	t.Run("Current", func(t *testing.T) {
		t.Run("uid", func(t *testing.T) {
			o := (&owner.OwnershipDiff{UIDs: &[2]int{1, 2}}).SetProxy(m)
			assert.Equal(t, "user: user-2 (2)", o.Current())
		})
		t.Run("gid", func(t *testing.T) {
			o := (&owner.OwnershipDiff{GIDs: &[2]int{1, 2}}).SetProxy(m)
			assert.Equal(t, "group: group-2 (2)", o.Current())
		})
		t.Run("both", func(t *testing.T) {
			o := (&owner.OwnershipDiff{UIDs: &[2]int{1, 2}, GIDs: &[2]int{1, 2}}).SetProxy(m)
			assert.Equal(t, "user: user-2 (2); group: group-2 (2)", o.Current())
		})
		t.Run("heterogenous", func(t *testing.T) {
			t.Run("mismatched-uid", func(t *testing.T) {
				o := (&owner.OwnershipDiff{UIDs: &[2]int{1, 2}, GIDs: &[2]int{1, 1}}).SetProxy(m)
				assert.Equal(t, "user: user-2 (2)", o.Current())
			})
			t.Run("mismatched-gid", func(t *testing.T) {
				o := (&owner.OwnershipDiff{UIDs: &[2]int{1, 1}, GIDs: &[2]int{1, 2}}).SetProxy(m)
				assert.Equal(t, "group: group-2 (2)", o.Current())
			})
		})
	})
	t.Run("Changes", func(t *testing.T) {
		t.Run("uid", func(t *testing.T) {
			o := (&owner.OwnershipDiff{UIDs: &[2]int{1, 2}}).SetProxy(m)
			assert.True(t, o.Changes())
		})
		t.Run("gid", func(t *testing.T) {
			o := (&owner.OwnershipDiff{GIDs: &[2]int{1, 2}}).SetProxy(m)
			assert.True(t, o.Changes())
		})
		t.Run("both", func(t *testing.T) {
			o := (&owner.OwnershipDiff{UIDs: &[2]int{1, 2}, GIDs: &[2]int{1, 2}}).SetProxy(m)
			assert.True(t, o.Changes())
		})
		t.Run("heterogenous", func(t *testing.T) {
			t.Run("mismatched-uid", func(t *testing.T) {
				o := (&owner.OwnershipDiff{UIDs: &[2]int{1, 2}, GIDs: &[2]int{1, 1}}).SetProxy(m)
				assert.True(t, o.Changes())
			})
			t.Run("mismatched-gid", func(t *testing.T) {
				o := (&owner.OwnershipDiff{UIDs: &[2]int{1, 1}, GIDs: &[2]int{1, 2}}).SetProxy(m)
				assert.True(t, o.Changes())
			})
		})
		t.Run("neither", func(t *testing.T) {
			o := (&owner.OwnershipDiff{}).SetProxy(m)
			assert.False(t, o.Changes())
		})
	})
	t.Run("NewOwnershipDiff", func(t *testing.T) {
		ownershipRecords := []ownershipRecord{
			makeOwned("foo", "user-1", "1", "group-1", "1"),
		}
		m := newMockOS(ownershipRecords, users, groups, nil, nil)
		t.Run("when-matching", func(t *testing.T) {
			o := &owner.Ownership{UID: intRef(1), GID: intRef(1)}
			d, err := owner.NewOwnershipDiff(m, "foo", o)
			require.NoError(t, err)
			assert.False(t, d.Changes())
		})
		t.Run("when-mismatched", func(t *testing.T) {
			o := &owner.Ownership{UID: intRef(2), GID: intRef(2)}
			d, err := owner.NewOwnershipDiff(m, "foo", o)
			require.NoError(t, err)
			assert.True(t, d.Changes())
		})
		t.Run("when-uid-match", func(t *testing.T) {
			o := &owner.Ownership{UID: intRef(1), GID: intRef(2)}
			d, err := owner.NewOwnershipDiff(m, "foo", o)
			require.NoError(t, err)
			assert.True(t, d.Changes())
		})
		t.Run("when-gid-match", func(t *testing.T) {
			o := &owner.Ownership{UID: intRef(2), GID: intRef(1)}
			d, err := owner.NewOwnershipDiff(m, "foo", o)
			require.NoError(t, err)
			assert.True(t, d.Changes())
		})
		t.Run("when-only-uid", func(t *testing.T) {
			t.Run("when-matches", func(t *testing.T) {
				o := &owner.Ownership{UID: intRef(1)}
				d, err := owner.NewOwnershipDiff(m, "foo", o)
				require.NoError(t, err)
				assert.False(t, d.Changes())
			})
			t.Run("when-not-matches", func(t *testing.T) {
				o := &owner.Ownership{UID: intRef(2)}
				d, err := owner.NewOwnershipDiff(m, "foo", o)
				require.NoError(t, err)
				assert.True(t, d.Changes())
			})
		})
		t.Run("when-only-gid", func(t *testing.T) {
			t.Run("when-matches", func(t *testing.T) {
				o := &owner.Ownership{GID: intRef(1)}
				d, err := owner.NewOwnershipDiff(m, "foo", o)
				require.NoError(t, err)
				assert.False(t, d.Changes())
			})
			t.Run("when-not-matches", func(t *testing.T) {
				o := &owner.Ownership{GID: intRef(2)}
				d, err := owner.NewOwnershipDiff(m, "foo", o)
				require.NoError(t, err)
				assert.True(t, d.Changes())
			})
		})
	})
	t.Run("when-syscall-errors", func(t *testing.T) {
		expectedError := errors.New("error")
		o := &owner.Ownership{UID: intRef(1), GID: intRef(1)}
		t.Run("GetUID", func(t *testing.T) {
			m := failingMockOS(map[string]error{"GetUID": expectedError})
			_, err := owner.NewOwnershipDiff(m, "foo", o)
			assert.Equal(t, expectedError, err)
		})
		t.Run("GetGID", func(t *testing.T) {
			m := failingMockOS(map[string]error{"GetGID": expectedError})
			_, err := owner.NewOwnershipDiff(m, "foo", o)
			assert.Equal(t, expectedError, err)
		})
	})
}

func TestApplyOwnershipDiff(t *testing.T) {
	users := []*user.User{
		fakeUser("1", "1", "user-1"),
		fakeUser("2", "2", "user-2"),
	}
	groups := []*user.Group{
		fakeGroup("1", "group-1"),
		fakeGroup("2", "group-2"),
	}
	ownershipRecords := []ownershipRecord{
		makeOwned("foo", "user-1", "1", "group-1", "1"),
	}
	m := newMockOS(ownershipRecords, users, groups, nil, nil)
	t.Run("no-changes", func(t *testing.T) {
		o := &owner.Ownership{UID: intRef(1), GID: intRef(1)}
		diff, err := owner.NewOwnershipDiff(m, "foo", o)
		require.NoError(t, err)
		err = diff.Apply()
		require.NoError(t, err)
		m.AssertNotCalled(t, "Chown", any, any, any)
	})
	t.Run("uid-changes", func(t *testing.T) {
		o := &owner.Ownership{UID: intRef(2), GID: intRef(1)}
		diff, err := owner.NewOwnershipDiff(m, "foo", o)
		require.NoError(t, err)
		err = diff.Apply()
		require.NoError(t, err)
		m.AssertCalled(t, "Chown", "foo", 2, 1)
	})
	t.Run("gid-changes", func(t *testing.T) {
		o := &owner.Ownership{UID: intRef(1), GID: intRef(2)}
		diff, err := owner.NewOwnershipDiff(m, "foo", o)
		require.NoError(t, err)
		err = diff.Apply()
		require.NoError(t, err)
		m.AssertCalled(t, "Chown", "foo", 1, 2)
	})
	t.Run("uid-and-gid-changes", func(t *testing.T) {
		o := &owner.Ownership{UID: intRef(2), GID: intRef(2)}
		diff, err := owner.NewOwnershipDiff(m, "foo", o)
		require.NoError(t, err)
		err = diff.Apply()
		require.NoError(t, err)
		m.AssertCalled(t, "Chown", "foo", 2, 2)
	})
	t.Run("chown-error-needs-changes", func(t *testing.T) {
		expected := errors.New("error1")
		m := failingMockOS(map[string]error{"Chown": expected})
		o := &owner.Ownership{UID: intRef(2), GID: intRef(2)}
		diff, err := owner.NewOwnershipDiff(m, "foo", o)
		require.NoError(t, err)
		err = diff.Apply()
		m.AssertCalled(t, "Chown", any, any, any)
		assert.Equal(t, expected, err)
	})
}

func intRef(i int) *int {
	return &i
}
