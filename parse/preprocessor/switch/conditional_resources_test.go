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

package control_test

import (
	"errors"
	"testing"

	"github.com/asteris-llc/converge/helpers/fakerenderer"
	"github.com/asteris-llc/converge/parse/preprocessor/switch"
	"github.com/asteris-llc/converge/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
)

// TestExecutionController ensures that execution controllers are respected
func TestExecutionController(t *testing.T) {
	mockTask := newMockTask(&resource.Status{}, errors.New("error1"))
	c := &control.ConditionalTask{
		Task: mockTask,
	}
	t.Run("When controller returns false does not call check", func(t *testing.T) {
		ctrl := newMockExecutionController(false)
		expected := &resource.Status{}
		c.SetExecutionController(ctrl)
		actual, err := c.Check(context.Background(), fakerenderer.New())
		require.NoError(t, err)
		assert.Equal(t, expected, actual)
		ctrl.AssertNotCalled(t, "Check", any)
	})
	t.Run("When controller returns false does not call apply", func(t *testing.T) {
		ctrl := newMockExecutionController(false)
		expected := &resource.Status{}
		c.SetExecutionController(ctrl)
		actual, err := c.Apply(context.Background())
		require.NoError(t, err)
		assert.Equal(t, expected, actual)
		ctrl.AssertNotCalled(t, "Apply")
	})
	t.Run("When controller returns true calls check", func(t *testing.T) {
		ctrl := newMockExecutionController(true)
		c.SetExecutionController(ctrl)
		c.Check(context.Background(), fakerenderer.New())
		mockTask.AssertCalled(t, "Check", any)
	})
	t.Run("When controller returns true calls check", func(t *testing.T) {
		ctrl := newMockExecutionController(true)
		c.SetExecutionController(ctrl)
		c.Apply(context.Background())
		mockTask.AssertCalled(t, "Apply")
	})
}
