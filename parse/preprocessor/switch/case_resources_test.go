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
	"github.com/stretchr/testify/require"
)

// TestPrepare tests that CasePreparer works as expected
func TestPrepare(t *testing.T) {
	t.Run("sets the predicate", func(t *testing.T) {
		mockRenderer := &MockRenderer{}
		mockRenderer.On("Render", any, any).Return("predicate1", nil)

		prep := &control.CasePreparer{Predicate: "something"}
		result, err := prep.Prepare(mockRenderer)
		assert.NoError(t, err)
		mockRenderer.AssertCalled(t, "Render", "predicate", "something")
		caseTask, ok := result.(*control.CaseTask)
		assert.True(t, ok)
		assert.Equal(t, "predicate1", caseTask.Predicate)
	})

	t.Run("sets the name", func(t *testing.T) {
		prep := &control.CasePreparer{Name: "name1"}
		result, err := prep.Prepare(defaultMockRenderer())
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

// TestShouldEvaluate ensures that we correctly understand when to evaluate and
// when to avoid evaluation.
func TestShouldEvaluate(t *testing.T) {
	trueCase := &control.CaseTask{Name: "trueCase", Predicate: "true"}
	falseCase := &control.CaseTask{Name: "false", Predicate: "false"}
	thisCase := &control.CaseTask{Name: "thisCase", Predicate: "true"}

	t.Run("when parent is nil", func(t *testing.T) {
		thisCase.SetParent(nil)
		assert.False(t, thisCase.ShouldEvaluate())
	})
	t.Run("when a previous case is true", func(t *testing.T) {
		parent := &control.SwitchTask{Branches: []string{"trueCase", "thisCase"}}
		parent.AppendCase(trueCase)
		parent.AppendCase(thisCase)
		assert.False(t, thisCase.ShouldEvaluate())
	})
	t.Run("when no previous case is true", func(t *testing.T) {
		parent := &control.SwitchTask{Branches: []string{"falseCase", "thisCase"}}
		parent.AppendCase(falseCase)
		parent.AppendCase(thisCase)
		assert.True(t, thisCase.ShouldEvaluate())
	})
	t.Run("when not in parent branches", func(t *testing.T) {
		parent := &control.SwitchTask{}
		thisCase.SetParent(parent)
		assert.False(t, thisCase.ShouldEvaluate())
	})
}

// TestIsTrue ensures that truth is correctly reported for predicates
func TestIsTrue(t *testing.T) {
	t.Run("when case is nil", func(t *testing.T) {
		var c *control.CaseTask
		isTrue, error := c.IsTrue()
		assert.Error(t, error)
		assert.False(t, isTrue)
	})
	t.Run("when parent is nil", func(t *testing.T) {
		c := &control.CaseTask{}
		c.SetParent(nil)
		isTrue, error := c.IsTrue()
		assert.Error(t, error)
		assert.False(t, isTrue)
	})
	t.Run("previous case is true", func(t *testing.T) {
		trueCase := &control.CaseTask{Name: "trueCase", Predicate: "true"}
		thisCase := &control.CaseTask{Name: "thisCase", Predicate: "true"}

		parent := &control.SwitchTask{Branches: []string{"trueCase", "thisCase"}}
		parent.AppendCase(trueCase)
		parent.AppendCase(thisCase)

		isTrue, error := thisCase.IsTrue()
		assert.NoError(t, error)
		assert.False(t, isTrue)
	})
	t.Run("no previous case is true", func(t *testing.T) {
		falseCase := &control.CaseTask{Name: "falseCase", Predicate: "false"}
		thisCase := &control.CaseTask{Name: "thisCase", Predicate: "true"}

		parent := &control.SwitchTask{Branches: []string{"falseCase", "thisCase"}}
		parent.AppendCase(falseCase)
		parent.AppendCase(thisCase)

		isTrue, error := thisCase.IsTrue()
		assert.NoError(t, error)
		assert.True(t, isTrue)
	})
	t.Run("unevaluated status", func(t *testing.T) {
		t.Run("returns true when predicate is true", func(t *testing.T) {
			thisCase := &control.CaseTask{Name: "thisCase", Predicate: "true"}
			parentCase(thisCase)
			isTrue, error := thisCase.IsTrue()
			assert.NoError(t, error)
			assert.True(t, isTrue)
		})
		t.Run("returns false when predicate is false", func(t *testing.T) {
			thisCase := &control.CaseTask{Name: "thisCase", Predicate: "false"}
			parentCase(thisCase)
			isTrue, error := thisCase.IsTrue()
			assert.NoError(t, error)
			assert.False(t, isTrue)
		})
		t.Run("caches results", func(t *testing.T) {
			thisCase := &control.CaseTask{Name: "thisCase", Predicate: "false", Status: nil}
			parentCase(thisCase)
			isTrue, error := thisCase.IsTrue()
			assert.NoError(t, error)
			assert.False(t, isTrue)
			require.NotNil(t, thisCase.Status)
			assert.False(t, *(thisCase.Status))
			thisCase.Predicate = "true"
			isTrue, error = thisCase.IsTrue()
			assert.NoError(t, error)
			assert.False(t, isTrue)
		})
	})
}

// TestEvaluatePredicate tests predicate evaluation
func TestEvaluatePredicate(t *testing.T) {
	t.Run("returns an error when invalid predicate", func(t *testing.T) {
		truth, err := control.EvaluatePredicate("foo | bar")
		assert.Error(t, err)
		assert.False(t, truth)
	})
	t.Run("returns an error when empty predicate", func(t *testing.T) {
		truth, err := control.EvaluatePredicate("")
		assert.Error(t, err)
		assert.False(t, truth)
	})
	t.Run("returns an error when invalid truth value", func(t *testing.T) {
		truth, err := control.EvaluatePredicate("\"foo\"")
		assert.Error(t, err)
		assert.False(t, truth)
	})
	t.Run("returns true when true", func(t *testing.T) {
		truth, err := control.EvaluatePredicate("true")
		assert.NoError(t, err)
		assert.True(t, truth)
	})
	t.Run("returns false when false", func(t *testing.T) {
		truth, err := control.EvaluatePredicate("false")
		assert.NoError(t, err)
		assert.False(t, truth)
	})
}

// TestCheck provides basic assurances about the operation of check
func TestCheck(t *testing.T) {
	c := &control.CaseTask{}
	stat, err := c.Check(fakerenderer.New())
	assert.NoError(t, err)
	assert.Equal(t, &resource.Status{}, stat)
}

// TestApply provides basic assurances about the operation of apply
func TestApply(t *testing.T) {
	c := &control.CaseTask{}
	stat, err := c.Apply()
	assert.NoError(t, err)
	assert.Equal(t, &resource.Status{}, stat)
}

func parentCase(c *control.CaseTask) {
	s := &control.SwitchTask{Branches: []string{c.Name}}
	s.AppendCase(c)
}
