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

package wait_test

import (
	"testing"

	"github.com/asteris-llc/converge/helpers/fakerenderer"
	"github.com/asteris-llc/converge/helpers/logging"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/wait"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
)

// TestPreparerInterface tests that the Preparer interface is properly
// implemeted
func TestPreparerInterface(t *testing.T) {
	t.Parallel()
	defer logging.HideLogs(t)()
	assert.Implements(t, (*resource.Resource)(nil), new(wait.Preparer))
}

// TestPreparerPrepare tests that Prepare initializes a Wait correctly
func TestPreparerPrepare(t *testing.T) {
	t.Parallel()
	defer logging.HideLogs(t)()

	t.Run("invalid check", func(t *testing.T) {
		p := &wait.Preparer{Check: ""}
		_, err := p.Prepare(context.Background(), fakerenderer.New())
		assert.Error(t, err)
	})

	t.Run("initializes task", func(t *testing.T) {
		var waitTask *wait.Wait
		p := &wait.Preparer{Check: "test"}
		r, err := p.Prepare(context.Background(), fakerenderer.New())
		require.NoError(t, err)
		require.IsType(t, (*wait.Wait)(nil), r)
		waitTask = r.(*wait.Wait)

		t.Run("shell", func(t *testing.T) {
			assert.NotNil(t, waitTask.Shell)
		})

		t.Run("initalizes retrier", func(t *testing.T) {
			assert.NotNil(t, waitTask.Retrier)
		})
	})

}
