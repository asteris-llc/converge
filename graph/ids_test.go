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

package graph_test

import (
	"testing"

	"github.com/asteris-llc/converge/graph"
	"github.com/stretchr/testify/assert"
)

func TestID(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "x/y", graph.ID("x", "y"))
}

func TestSiblingID(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "x/z", graph.SiblingID("x/y", "z"))
}

func TestBaseID(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "b", graph.BaseID("a/b"))
}

func TestIsDescendentID(t *testing.T) {
	t.Parallel()

	assert.True(t, graph.IsDescendentID("a", "a/b"))
}

func TestIsDescendentIDNot(t *testing.T) {
	t.Parallel()

	assert.False(t, graph.IsDescendentID("a/b", "a/c"))
}

// TestIsRoot tests the IsRoot function
func TestIsRoot(t *testing.T) {
	t.Parallel()

	t.Run("is root", func(t *testing.T) {
		assert.True(t, graph.IsRoot("root"))
	})

	t.Run("is not root", func(t *testing.T) {
		assert.False(t, graph.IsRoot("root/module.test"))
	})
}
