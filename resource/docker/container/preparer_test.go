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

package container_test

import (
	"testing"

	"github.com/asteris-llc/converge/helpers/fakerenderer"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/docker/container"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
)

// TestPreparerInterface tests that the Preparer interface is properly
// implemented
func TestPreparerInterface(t *testing.T) {
	t.Parallel()
	assert.Implements(t, (*resource.Resource)(nil), new(container.Preparer))
}

// TestPrepare tests Prepare
func TestPrepare(t *testing.T) {
	t.Parallel()

	t.Run("default network mode", func(t *testing.T) {
		p := &container.Preparer{Name: "test", Image: "nginx"}
		task, err := p.Prepare(context.Background(), fakerenderer.New())
		require.NoError(t, err)
		con := task.(*container.Container)
		assert.Equal(t, container.DefaultNetworkMode, con.NetworkMode)
	})
}
