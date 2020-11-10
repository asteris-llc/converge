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
				executor:   &ExecutorMock{},
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
				executor:   &ExecutorMock{},
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
				executor:   &ExecutorMock{},
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
				executor:   &ExecutorMock{},
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
				executor:   &ExecutorMock{},
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
				executor:     &ExecutorMock{},
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
				executor:     &ExecutorMock{},
			}
			_, err := p.Prepare(context.Background(), fakerenderer.New())
			require.Error(t, err)
		})
	})
	t.Run("sets-fields", func(t *testing.T) {
		t.Parallel()
		untrue := false
		res, err := (&Preparer{
			Name:          "test1",
			State:         "state1",
			Reload:        true,
			Enable:        &untrue,
			EnableRuntime: &untrue,
			executor:      &ExecutorMock{},
		}).Prepare(context.Background(), fakerenderer.New())
		require.NoError(t, err)
		assert.Equal(t, "test1", res.(*Resource).Name)
		assert.Equal(t, "state1", res.(*Resource).State)
		assert.True(t, res.(*Resource).Reload)
		assert.False(t, res.(*Resource).sendSignal)
		assert.Equal(t, "", res.(*Resource).SignalName)
		assert.Equal(t, uint(0), res.(*Resource).SignalNumber)
		assert.False(t, *res.(*Resource).enableChange)
		assert.False(t, *res.(*Resource).enableRuntimeChange)
	})
	t.Run("handles-enable-disable", func(t *testing.T) {
		t.Parallel()
		valTrue := true
		valFalse := false
		t.Run("when-true-true", func(t *testing.T) {
			t.Parallel()
			_, err := (&Preparer{
				Enable:        &valTrue,
				EnableRuntime: &valTrue,
				executor:      &ExecutorMock{},
			}).Prepare(context.Background(), fakerenderer.New())
			assert.NoError(t, err)
		})
		t.Run("when-false-true", func(t *testing.T) {
			t.Parallel()
			_, err := (&Preparer{
				Enable:        &valFalse,
				EnableRuntime: &valTrue,
				executor:      &ExecutorMock{},
			}).Prepare(context.Background(), fakerenderer.New())
			assert.NoError(t, err)
		})
		t.Run("when-true-false", func(t *testing.T) {
			t.Parallel()
			_, err := (&Preparer{
				Enable:        &valTrue,
				EnableRuntime: &valFalse,
				executor:      &ExecutorMock{},
			}).Prepare(context.Background(), fakerenderer.New())
			assert.NoError(t, err)
		})
		t.Run("when-false-false", func(t *testing.T) {
			t.Parallel()
			_, err := (&Preparer{
				Enable:        &valFalse,
				EnableRuntime: &valFalse,
				executor:      &ExecutorMock{},
			}).Prepare(context.Background(), fakerenderer.New())
			assert.NoError(t, err)
		})
		t.Run("when-true-nil", func(t *testing.T) {
			t.Parallel()
			_, err := (&Preparer{
				Enable:   &valTrue,
				executor: &ExecutorMock{},
			}).Prepare(context.Background(), fakerenderer.New())
			assert.NoError(t, err)
		})
		t.Run("when-false-nil", func(t *testing.T) {
			t.Parallel()
			_, err := (&Preparer{
				Enable:   &valFalse,
				executor: &ExecutorMock{},
			}).Prepare(context.Background(), fakerenderer.New())
			assert.NoError(t, err)
		})
		t.Run("when-nil-true", func(t *testing.T) {
			t.Parallel()
			_, err := (&Preparer{
				EnableRuntime: &valTrue,
				executor:      &ExecutorMock{},
			}).Prepare(context.Background(), fakerenderer.New())
			assert.NoError(t, err)
		})
		t.Run("when-nil-alse", func(t *testing.T) {
			t.Parallel()
			_, err := (&Preparer{
				EnableRuntime: &valFalse,
				executor:      &ExecutorMock{},
			}).Prepare(context.Background(), fakerenderer.New())
			assert.NoError(t, err)
		})
	})

}
