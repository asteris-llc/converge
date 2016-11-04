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

package network_test

import (
	"testing"

	"github.com/asteris-llc/converge/helpers/fakerenderer"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/docker/network"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
)

// TestPreparerInterface ensures that the correct interfaces are implemented by
// the preparer
func TestPreparerInterface(t *testing.T) {
	t.Parallel()
	assert.Implements(t, (*resource.Resource)(nil), new(network.Preparer))
}

// TestPreparerPrepare tests the Prepare function
func TestPreparerPrepare(t *testing.T) {
	t.Run("name is required", func(t *testing.T) {
		p := &network.Preparer{Name: ""}
		_, err := p.Prepare(context.Background(), fakerenderer.New())
		if assert.Error(t, err) {
			assert.EqualError(t, err, "name must be provided")
		}
	})

	t.Run("state defaults to present", func(t *testing.T) {
		p := &network.Preparer{Name: "test-network"}
		task, err := p.Prepare(context.Background(), fakerenderer.New())
		require.NoError(t, err)
		require.IsType(t, (*network.Network)(nil), task)
		nw := task.(*network.Network)
		assert.Equal(t, network.StatePresent, nw.State)
	})

	t.Run("driver defaults to bridge", func(t *testing.T) {
		p := &network.Preparer{Name: "test-network"}
		task, err := p.Prepare(context.Background(), fakerenderer.New())
		require.NoError(t, err)
		require.IsType(t, (*network.Network)(nil), task)
		nw := task.(*network.Network)
		assert.Equal(t, network.DefaultDriver, nw.Driver)
	})
}
