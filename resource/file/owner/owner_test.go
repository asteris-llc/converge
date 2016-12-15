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

func TestInterface(t *testing.T) {
	assert.Implements(t, (*resource.Task)(nil), new(owner.Owner))
}

func TestCheck(t *testing.T) {
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
	t.Run("when-user-no-change", func(t *testing.T) {
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
}
