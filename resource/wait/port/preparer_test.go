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

package port_test

import (
	"testing"

	"github.com/asteris-llc/converge/helpers/fakerenderer"
	"github.com/asteris-llc/converge/helpers/logging"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/wait/port"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPreparerInterface tests that the Preparer interface is properly
// implemeted
func TestPreparerInterface(t *testing.T) {
	t.Parallel()
	assert.Implements(t, (*resource.Resource)(nil), new(port.Preparer))
}

// TestPreparerPrepare tests that Prepare initializes a Port correctly
func TestPreparerPrepare(t *testing.T) {
	t.Parallel()
	defer logging.HideLogs(t)()

	t.Run("initializes port", func(t *testing.T) {
		var portTask *port.Port
		p := &port.Preparer{Port: 8080, Host: "hostname"}
		r, err := p.Prepare(fakerenderer.New())
		require.NoError(t, err)
		require.IsType(t, (*port.Port)(nil), r)
		portTask = r.(*port.Port)

		t.Run("sets host", func(t *testing.T) {
			assert.Equal(t, "hostname", portTask.Host)
		})

		t.Run("sets port", func(t *testing.T) {
			assert.Equal(t, 8080, portTask.Port)
		})

		t.Run("initalizes retrier", func(t *testing.T) {
			assert.NotNil(t, portTask.Retrier)
		})
	})

	t.Run("invalid port", func(t *testing.T) {
		p := &port.Preparer{Port: 0, Host: "hostname"}
		_, err := p.Prepare(fakerenderer.New())
		assert.Error(t, err)
	})
}
