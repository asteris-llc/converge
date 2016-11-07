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
	"testing"

	"github.com/asteris-llc/converge/helpers/fakerenderer"
	"github.com/asteris-llc/converge/parse/preprocessor/switch"
	"github.com/asteris-llc/converge/resource"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

// TestPrepare tests that CasePreparer works as expected
func TestPrepare(t *testing.T) {
	t.Run("sets the predicate", func(t *testing.T) {
		mockRenderer := &MockRenderer{}
		mockRenderer.On("Render", any, any).Return("predicate1", nil)

		prep := &control.CasePreparer{}
		result, err := prep.Prepare(context.Background(), mockRenderer)
		assert.NoError(t, err)
		_, ok := result.(*control.CaseTask)
		assert.True(t, ok)
	})

	t.Run("sets the name", func(t *testing.T) {
		prep := &control.CasePreparer{Name: "name1"}
		result, err := prep.Prepare(context.Background(), defaultMockRenderer())
		assert.NoError(t, err)
		caseTask, ok := result.(*control.CaseTask)
		assert.True(t, ok)
		assert.Equal(t, "name1", caseTask.Name)
	})
}

// TestIsDefault ensures that we correctly understand when we are the default
// case.
func TestIsDefault(t *testing.T) {
	c := &control.CaseTask{}
	t.Run("when nil", func(t *testing.T) {
		var nilCase *control.CaseTask
		assert.False(t, nilCase.IsDefault())
	})
	t.Run("when default", func(t *testing.T) {
		c.Name = "default"
		assert.True(t, c.IsDefault())
	})
	t.Run("when notDefault", func(t *testing.T) {
		c.Name = "name1"
		assert.False(t, c.IsDefault())
	})
}

// TestSetGetParent ensures that setting and getting the parent switch task
// works as expected.
func TestSetGetParent(t *testing.T) {
	parent := &control.SwitchTask{}
	c := &control.CaseTask{}
	c.SetParent(parent)
	assert.Equal(t, parent, c.GetParent())
}

// TestCheck provides basic assurances about the operation of check
func TestCheck(t *testing.T) {
	c := &control.CaseTask{}
	stat, err := c.Check(context.Background(), fakerenderer.New())
	assert.NoError(t, err)
	assert.Equal(t, &resource.Status{}, stat)
}

// TestApply provides basic assurances about the operation of apply
func TestApply(t *testing.T) {
	c := &control.CaseTask{}
	stat, err := c.Apply(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, &resource.Status{}, stat)
}

func parentCase(c *control.CaseTask) {
	s := &control.SwitchTask{Branches: []string{c.Name}}
	s.AppendCase(c)
}
