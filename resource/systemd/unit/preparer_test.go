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

package unit

import (
	"testing"

	"github.com/asteris-llc/converge/helpers/fakerenderer"
	"github.com/asteris-llc/converge/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
)

func TestPreparerInterface(t *testing.T) {
	t.Parallel()
	assert.Implements(t, (*resource.Resource)(nil), new(Preparer))
}

// TestPreparer runs a test
func TestPreparer(t *testing.T) {
	t.Parallel()

	t.Run("sets-signal-when-signal-name", func(t *testing.T) {
		t.Parallel()
		t.Run("when-uppercase", func(t *testing.T) {
			t.Parallel()
			p := &Preparer{
				SignalName: "KILL",
				executor:   LinuxExecutor{&DbusMock{}},
			}
			res, err := p.Prepare(context.Background(), fakerenderer.New())
			require.NoError(t, err)
			assert.Equal(t, "SIGKILL", res.(*Resource).SignalName)
			assert.Equal(t, uint(9), res.(*Resource).SignalNumber)
			assert.True(t, res.(*Resource).sendSignal)
		})
		t.Run("when-lowercase", func(t *testing.T) {
			t.Parallel()
			p := &Preparer{
				SignalName: "kill",
				executor:   LinuxExecutor{&DbusMock{}},
			}
			res, err := p.Prepare(context.Background(), fakerenderer.New())
			require.NoError(t, err)
			assert.Equal(t, "SIGKILL", res.(*Resource).SignalName)
			assert.Equal(t, uint(9), res.(*Resource).SignalNumber)
			assert.True(t, res.(*Resource).sendSignal)
		})
		t.Run("when-sig-prefix", func(t *testing.T) {
			t.Parallel()
			p := &Preparer{
				SignalName: "sigkill",
				executor:   LinuxExecutor{&DbusMock{}},
			}
			res, err := p.Prepare(context.Background(), fakerenderer.New())
			require.NoError(t, err)
			assert.Equal(t, "SIGKILL", res.(*Resource).SignalName)
			assert.Equal(t, uint(9), res.(*Resource).SignalNumber)
			assert.True(t, res.(*Resource).sendSignal)
		})
		t.Run("when-mixed-case", func(t *testing.T) {
			t.Parallel()
			p := &Preparer{
				SignalName: randomizeCase("sigkill"),
				executor:   LinuxExecutor{&DbusMock{}},
			}
			res, err := p.Prepare(context.Background(), fakerenderer.New())
			require.NoError(t, err)
			assert.Equal(t, "SIGKILL", res.(*Resource).SignalName)
			assert.Equal(t, uint(9), res.(*Resource).SignalNumber)
			assert.True(t, res.(*Resource).sendSignal)
		})
		t.Run("when-invalid", func(t *testing.T) {
			t.Parallel()
			p := &Preparer{
				SignalName: randomizeCase("badsignal1"),
				executor:   LinuxExecutor{&DbusMock{}},
			}
			_, err := p.Prepare(context.Background(), fakerenderer.New())
			require.Error(t, err)
		})
	})
	t.Run("sets-signal-when-signal-number", func(t *testing.T) {
		t.Parallel()
		t.Run("when-valid", func(t *testing.T) {
			t.Parallel()
			p := &Preparer{
				SignalNumber: 9,
				executor:     LinuxExecutor{&DbusMock{}},
			}
			res, err := p.Prepare(context.Background(), fakerenderer.New())
			require.NoError(t, err)
			assert.Equal(t, "SIGKILL", res.(*Resource).SignalName)
			assert.Equal(t, uint(9), res.(*Resource).SignalNumber)
			assert.True(t, res.(*Resource).sendSignal)
		})
		t.Run("when-invalid", func(t *testing.T) {
			t.Parallel()
			p := &Preparer{
				SignalNumber: 99,
				executor:     LinuxExecutor{&DbusMock{}},
			}
			_, err := p.Prepare(context.Background(), fakerenderer.New())
			require.Error(t, err)
		})
	})
	t.Run("sets-fields", func(t *testing.T) {
		t.Parallel()
		res, err := (&Preparer{
			Name:     "test1",
			State:    "state1",
			Reload:   true,
			executor: LinuxExecutor{&DbusMock{}},
		}).Prepare(context.Background(), fakerenderer.New())
		require.NoError(t, err)
		assert.Equal(t, "test1", res.(*Resource).Name)
		assert.Equal(t, "state1", res.(*Resource).State)
		assert.True(t, res.(*Resource).Reload)
		assert.False(t, res.(*Resource).sendSignal)
		assert.Equal(t, "", res.(*Resource).SignalName)
		assert.Equal(t, uint(0), res.(*Resource).SignalNumber)
	})

}
