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
	"context"
	"os/user"
	"testing"

	"github.com/asteris-llc/converge/helpers/fakerenderer"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/file/owner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInterface ensures that owner.Owner implements the resource.Task interface
func TestInterface(t *testing.T) {
	t.Parallel()
	assert.Implements(t, (*resource.Task)(nil), new(owner.Owner))
}

// TestCheck tests the behavior of Check
func TestCheck(t *testing.T) {
	t.Parallel()

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

	t.Run("when-user-and-group-no-change", func(t *testing.T) {
		t.Parallel()

		o := (&owner.Owner{
			Destination: "foo",
			Username:    "user-1",
			UID:         "1",
			Group:       "group-1",
			GID:         "1",
		}).SetOSProxy(m)
		status, err := o.Check(context.Background(), fakerenderer.New())
		require.NoError(t, err)
		m.AssertCalled(t, "GetGID", any)
		m.AssertCalled(t, "GetUID", any)
		assert.False(t, status.HasChanges())
	})

	t.Run("when-missing", func(t *testing.T) {
		t.Parallel()
		m := missingMockOS()
		o := (&owner.Owner{
			Destination: "foo",
			Username:    "user-1",
			UID:         "1",
			Group:       "group-1",
			GID:         "1",
		}).SetOSProxy(m)
		status, err := o.Check(context.Background(), fakerenderer.New())
		require.NoError(t, err)
		assert.True(t, status.HasChanges())
		m.AssertNotCalled(t, "GetGID", any)
		m.AssertNotCalled(t, "GetUID", any)
	})

	t.Run("when-user-no-change", func(t *testing.T) {
		t.Parallel()

		o := (&owner.Owner{
			Destination: "foo",
			Username:    "user-1",
			UID:         "1",
		}).SetOSProxy(m)
		status, err := o.Check(context.Background(), fakerenderer.New())
		require.NoError(t, err)
		m.AssertCalled(t, "GetGID", any)
		m.AssertCalled(t, "GetUID", any)
		assert.False(t, status.HasChanges())
	})

	t.Run("when-group-no-change", func(t *testing.T) {
		t.Parallel()

		o := (&owner.Owner{
			Destination: "foo",
			Group:       "group-1",
			GID:         "1",
		}).SetOSProxy(m)
		status, err := o.Check(context.Background(), fakerenderer.New())
		require.NoError(t, err)
		m.AssertCalled(t, "GetGID", any)
		m.AssertCalled(t, "GetUID", any)
		assert.False(t, status.HasChanges())
	})

	t.Run("when-user-and-group-change", func(t *testing.T) {
		t.Parallel()

		o := (&owner.Owner{
			Destination: "foo",
			Username:    "user-2",
			UID:         "2",
			Group:       "group-2",
			GID:         "2",
		}).SetOSProxy(m)
		status, err := o.Check(context.Background(), fakerenderer.New())
		require.NoError(t, err)
		m.AssertCalled(t, "GetGID", any)
		m.AssertCalled(t, "GetUID", any)
		assert.True(t, status.HasChanges())
		resStat, ok := status.(*resource.Status)
		require.True(t, ok)
		fooDiffs, ok := resStat.Differences["foo"]
		require.True(t, ok)
		diff, ok := fooDiffs.(*owner.OwnershipDiff)
		require.True(t, ok)
		require.NotNil(t, diff.UIDs)
		require.NotNil(t, diff.GIDs)
		assert.Equal(t, [2]int{1, 2}, *(diff.UIDs))
		assert.Equal(t, [2]int{1, 2}, *(diff.GIDs))
	})

	t.Run("when-user-change-and-group-no-change", func(t *testing.T) {
		t.Parallel()

		o := (&owner.Owner{
			Destination: "foo",
			Username:    "user-2",
			UID:         "2",
			Group:       "group-1",
			GID:         "1",
		}).SetOSProxy(m)
		status, err := o.Check(context.Background(), fakerenderer.New())
		require.NoError(t, err)
		m.AssertCalled(t, "GetGID", any)
		m.AssertCalled(t, "GetUID", any)
		assert.True(t, status.HasChanges())
		resStat, ok := status.(*resource.Status)
		require.True(t, ok)
		fooDiffs, ok := resStat.Differences["foo"]
		require.True(t, ok)
		diff, ok := fooDiffs.(*owner.OwnershipDiff)
		require.True(t, ok)
		require.NotNil(t, diff.UIDs)
		require.Nil(t, diff.GIDs)
		assert.Equal(t, [2]int{1, 2}, *(diff.UIDs))
	})

	t.Run("when-user-change-and-group-unspecified", func(t *testing.T) {
		t.Parallel()

		o := (&owner.Owner{
			Destination: "foo",
			Username:    "user-2",
			UID:         "2",
			Group:       "",
			GID:         "",
		}).SetOSProxy(m)
		status, err := o.Check(context.Background(), fakerenderer.New())
		require.NoError(t, err)
		m.AssertCalled(t, "GetGID", any)
		m.AssertCalled(t, "GetUID", any)
		assert.True(t, status.HasChanges())
		resStat, ok := status.(*resource.Status)
		require.True(t, ok)
		fooDiffs, ok := resStat.Differences["foo"]
		require.True(t, ok)
		diff, ok := fooDiffs.(*owner.OwnershipDiff)
		require.True(t, ok)
		require.NotNil(t, diff.UIDs)
		require.Nil(t, diff.GIDs)
		assert.Equal(t, [2]int{1, 2}, *(diff.UIDs))
	})

	t.Run("when-group-change-and-user-no-change", func(t *testing.T) {
		t.Parallel()

		o := (&owner.Owner{
			Destination: "foo",
			Username:    "user-1",
			UID:         "1",
			Group:       "group-2",
			GID:         "2",
		}).SetOSProxy(m)
		status, err := o.Check(context.Background(), fakerenderer.New())
		require.NoError(t, err)
		m.AssertCalled(t, "GetGID", any)
		m.AssertCalled(t, "GetUID", any)
		assert.True(t, status.HasChanges())
		resStat, ok := status.(*resource.Status)
		require.True(t, ok)
		fooDiffs, ok := resStat.Differences["foo"]
		require.True(t, ok)
		diff, ok := fooDiffs.(*owner.OwnershipDiff)
		require.True(t, ok)
		require.NotNil(t, diff.GIDs)
		require.Nil(t, diff.UIDs)
		assert.Equal(t, [2]int{1, 2}, *(diff.GIDs))
	})

	t.Run("when-group-change-and-user-unspecified", func(t *testing.T) {
		t.Parallel()

		o := (&owner.Owner{
			Destination: "foo",
			Username:    "",
			UID:         "",
			Group:       "group-2",
			GID:         "2",
		}).SetOSProxy(m)
		status, err := o.Check(context.Background(), fakerenderer.New())
		require.NoError(t, err)
		m.AssertCalled(t, "GetGID", any)
		m.AssertCalled(t, "GetUID", any)
		assert.True(t, status.HasChanges())
		resStat, ok := status.(*resource.Status)
		require.True(t, ok)
		fooDiffs, ok := resStat.Differences["foo"]
		require.True(t, ok)
		diff, ok := fooDiffs.(*owner.OwnershipDiff)
		require.True(t, ok)
		require.NotNil(t, diff.GIDs)
		require.Nil(t, diff.UIDs)
		assert.Equal(t, [2]int{1, 2}, *(diff.GIDs))
	})

	t.Run("when-recursive", func(t *testing.T) {
		t.Parallel()
		t.Run("when-all-change", func(t *testing.T) {
			t.Parallel()
			ownershipRecords := []ownershipRecord{
				makeOwnedFull("root", "user-1", "1", "group-1", "1", true, 0, 0755),
				makeOwnedFull("root/dir1", "user-1", "1", "group-1", "1", true, 0, 0755),
				makeOwnedFull("root/dir1/a", "user-1", "1", "group-1", "1", false, 0, 0755),
				makeOwnedFull("root/dir2", "user-1", "1", "group-1", "1", true, 0, 0755),
				makeOwnedFull("root/dir2/a", "user-1", "1", "group-1", "1", false, 0, 0755),
				makeOwnedFull("root/dir2/b", "user-1", "1", "group-1", "1", false, 0, 0755),
				makeOwnedFull("root/file3", "user-1", "1", "group-1", "1", false, 0, 0755),
			}
			m := newMockOS(ownershipRecords, users, groups, nil, nil)
			o := (&owner.Owner{
				Destination: "root",
				Username:    "user-2",
				UID:         "2",
				Group:       "group-2",
				GID:         "2",
				Recursive:   true,
			}).SetOSProxy(m)
			status, err := o.Check(context.Background(), fakerenderer.New())
			require.NoError(t, err)
			m.AssertCalled(t, "Walk", "root", any)
			m.AssertNumberOfCalls(t, "Walk", 1)
			for _, rec := range ownershipRecords {
				m.AssertCalled(t, "GetGID", rec.Path)
				m.AssertCalled(t, "GetUID", rec.Path)
			}
			resStatus, ok := status.(*resource.Status)
			require.True(t, ok)
			diffs := resStatus.Differences
			assert.Equal(t, len(diffs), len(ownershipRecords))
		})
		t.Run("when-some-change", func(t *testing.T) {
			t.Parallel()
			ownershipRecords := []ownershipRecord{
				makeOwnedFull("root", "user-1", "1", "group-1", "1", true, 0, 0755),
				makeOwnedFull("root/dir1", "user-1", "1", "group-1", "1", true, 0, 0755),
				makeOwnedFull("root/dir1/a", "user-1", "1", "group-1", "1", false, 0, 0755),

				makeOwnedFull("root/file3", "user-1", "1", "group-1", "1", false, 0, 0755),
			}
			changingRecords := []ownershipRecord{
				makeOwnedFull("root/dir2", "user-1", "1", "group-2", "2", true, 0, 0755),
				makeOwnedFull("root/dir2/a", "user-1", "1", "group-2", "2", false, 0, 0755),
				makeOwnedFull("root/dir2/b", "user-1", "1", "group-2", "2", false, 0, 0755),
			}
			allRecords := append(ownershipRecords, changingRecords...)
			m := newMockOS(allRecords, users, groups, nil, nil)
			o := (&owner.Owner{
				Destination: "root",
				Username:    "user-1",
				UID:         "1",
				Group:       "group-1",
				GID:         "1",
				Recursive:   true,
			}).SetOSProxy(m)
			status, err := o.Check(context.Background(), fakerenderer.New())
			require.NoError(t, err)
			m.AssertCalled(t, "Walk", "root", any)
			m.AssertNumberOfCalls(t, "Walk", 1)
			for _, rec := range allRecords {
				m.AssertCalled(t, "GetGID", rec.Path)
				m.AssertCalled(t, "GetUID", rec.Path)
			}
			resStatus, ok := status.(*resource.Status)
			require.True(t, ok)
			diffs := resStatus.Differences
			for _, shouldChange := range changingRecords {
				_, ok := diffs[shouldChange.Path]
				assert.True(t, ok)
			}
			assert.Equal(t, len(diffs), len(changingRecords))
		})
	})
}

// TestApply tests the behavior of Apply
func TestApply(t *testing.T) {
	t.Parallel()
	t.Run("single-diff", func(t *testing.T) {
		users := []*user.User{fakeUser("1", "1", "user-1"), fakeUser("2", "2", "user-2")}
		groups := []*user.Group{fakeGroup("1", "group-1"), fakeGroup("2", "group-2")}
		ownershipRecords := []ownershipRecord{makeOwned("foo", "user-1", "1", "group-1", "1")}
		m := newMockOS(ownershipRecords, users, groups, nil, nil)
		o := (&owner.Owner{
			Destination: "foo",
			Username:    "user-2",
			UID:         "2",
			Group:       "group-2",
			GID:         "2",
		}).SetOSProxy(m)
		_, err := o.Check(context.Background(), fakerenderer.New())
		require.NoError(t, err)
		_, err = o.Apply(context.Background())
		require.NoError(t, err)
		m.AssertCalled(t, "Chown", "foo", 2, 2)
	})
	t.Run("multiple-diff", func(t *testing.T) {
		users := []*user.User{fakeUser("1", "1", "user-1"), fakeUser("2", "2", "user-2")}
		groups := []*user.Group{fakeGroup("1", "group-1"), fakeGroup("2", "group-2")}
		ownershipRecords := []ownershipRecord{makeOwned("foo", "user-1", "1", "group-1", "1"), makeOwned("bar", "user-1", "1", "group-1", "1")}
		m := newMockOS(ownershipRecords, users, groups, nil, nil)
		o := (&owner.Owner{
			Destination: "foo",
			Username:    "user-2",
			Recursive:   true,
			UID:         "2",
			Group:       "group-2",
			GID:         "2",
		}).SetOSProxy(m)
		_, err := o.Check(context.Background(), fakerenderer.New())
		require.NoError(t, err)
		_, err = o.Apply(context.Background())
		require.NoError(t, err)
		m.AssertCalled(t, "Chown", "foo", 2, 2)
		m.AssertCalled(t, "Chown", "bar", 2, 2)
		m.AssertNumberOfCalls(t, "Chown", 2)
	})
	t.Run("when-missing", func(t *testing.T) {
		t.Parallel()
		users := []*user.User{fakeUser("1", "1", "user-1"), fakeUser("2", "2", "user-2")}
		groups := []*user.Group{fakeGroup("1", "group-1"), fakeGroup("2", "group-2")}
		ownershipRecords := []ownershipRecord{makeOwned("foo", "user-1", "1", "group-1", "1")}
		m := newMockOS(ownershipRecords, users, groups, nil, nil)
		missing := missingMockOS()
		o := (&owner.Owner{
			Destination: "foo",
			Username:    "user-1",
			UID:         "1",
			Group:       "group-1",
			GID:         "1",
		}).SetOSProxy(missing)
		_, err := o.Check(context.Background(), fakerenderer.New())
		require.NoError(t, err)
		o.SetOSProxy(m)
		_, err = o.Apply(context.Background())
		m.AssertCalled(t, "GetGID", any)
		m.AssertCalled(t, "GetUID", any)
	})
}
