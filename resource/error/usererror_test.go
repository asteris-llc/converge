// Copyright © 2017 Asteris, LLC
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

package usererror

import (
	"testing"

	"github.com/asteris-llc/converge/helpers/fakerenderer"
	"github.com/asteris-llc/converge/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
)

const (
	_testRuntimeErrorMsg string = "runtime error: explicit error encountered"
)

func TestInterfaces(t *testing.T) {
	assert.Implements(t, (*resource.Resource)(nil), new(Preparer))
	assert.Implements(t, (*resource.Resource)(nil), new(ApplyPreparer))
	assert.Implements(t, (*resource.Task)(nil), new(UserError))
}

func TestPrepare(t *testing.T) {
	err := "error1"
	t.Run("Preparer", func(t *testing.T) {
		p := &Preparer{err}
		tsk, e := p.Prepare(context.Background(), fakerenderer.New())
		require.NoError(t, e)
		userErr, ok := tsk.(*UserError)
		require.True(t, ok)
		assert.Equal(t, err, userErr.Error)
	})
	t.Run("ApplyPreparer", func(t *testing.T) {
		p := &ApplyPreparer{err}
		tsk, e := p.Prepare(context.Background(), fakerenderer.New())
		require.NoError(t, e)
		userErr, ok := tsk.(*UserError)
		require.True(t, ok)
		assert.Equal(t, err, userErr.Error)
	})
}

func TestCheck(t *testing.T) {
	t.Run("skips-when-skip-plan-true", func(t *testing.T) {
		e := &UserError{SkipPlan: true}
		result, err := e.Check(context.Background(), fakerenderer.New())
		require.NoError(t, err)
		assert.True(t, result.HasChanges())
		assert.Equal(t, resource.StatusWillChange, result.StatusCode())
	})

	t.Run("adds-message-with-user-error", func(t *testing.T) {
		expected := "error1"
		e := &UserError{Error: expected}
		result, err := e.Check(context.Background(), fakerenderer.New())
		require.NoError(t, err)
		assert.Equal(t, resource.StatusFatal, result.StatusCode())
		assert.Equal(t, []string{_testRuntimeErrorMsg, expected}, result.Messages())
	})
}

func TestApply(t *testing.T) {
	expected := "error1"
	e := &UserError{Error: expected}
	result, err := e.Apply(context.Background())
	require.NoError(t, err)
	assert.Equal(t, resource.StatusFatal, result.StatusCode())
	assert.Equal(t, []string{_testRuntimeErrorMsg, expected}, result.Messages())
	assert.True(t, e.changed)
}
