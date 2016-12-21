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

// TestPreparer tests Prepare
func TestPreparer(t *testing.T) {
	users := []*user.User{fakeUser("1", "1", "user-1")}
	groups := []*user.Group{fakeGroup("1", "group-1")}
	t.Run("implements-resource", func(t *testing.T) {
		assert.Implements(t, (*resource.Resource)(nil), new(owner.Preparer))
	})
	t.Run("normalizes-data", func(t *testing.T) {
		m := newMockOS(nil, users, groups, nil, nil)
		t.Run("when-username", func(t *testing.T) {
			p := (&owner.Preparer{Username: "user-1"}).SetOSProxy(m)
			oRes, err := p.Prepare(context.Background(), fakerenderer.New())
			require.NoError(t, err)
			o, ok := oRes.(*owner.Owner)
			require.True(t, ok)
			assert.Equal(t, "user-1", o.Username)
			assert.Equal(t, "1", o.UID)
			assert.Equal(t, "", o.Group)
			assert.Equal(t, "", o.GID)
		})
		t.Run("when-uid", func(t *testing.T) {
			u := new(int)
			*u = 1
			p := (&owner.Preparer{UID: u}).SetOSProxy(m)
			oRes, err := p.Prepare(context.Background(), fakerenderer.New())
			require.NoError(t, err)
			o, ok := oRes.(*owner.Owner)
			require.True(t, ok)
			assert.Equal(t, "user-1", o.Username)
			assert.Equal(t, "1", o.UID)
			assert.Equal(t, "", o.Group)
			assert.Equal(t, "", o.GID)
		})
		t.Run("when-groupname", func(t *testing.T) {
			p := (&owner.Preparer{Groupname: "group-1"}).SetOSProxy(m)
			oRes, err := p.Prepare(context.Background(), fakerenderer.New())
			require.NoError(t, err)
			o, ok := oRes.(*owner.Owner)
			require.True(t, ok)
			assert.Equal(t, "", o.Username)
			assert.Equal(t, "", o.UID)
			assert.Equal(t, "group-1", o.Group)
			assert.Equal(t, "1", o.GID)
		})
		t.Run("when-gid", func(t *testing.T) {
			g := new(int)
			*g = 1
			p := (&owner.Preparer{GID: g}).SetOSProxy(m)
			oRes, err := p.Prepare(context.Background(), fakerenderer.New())
			require.NoError(t, err)
			o, ok := oRes.(*owner.Owner)
			require.True(t, ok)
			assert.Equal(t, "", o.Username)
			assert.Equal(t, "", o.UID)
			assert.Equal(t, "group-1", o.Group)
			assert.Equal(t, "1", o.GID)
		})
	})
}
