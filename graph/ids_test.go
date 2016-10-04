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

func TestParentID(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "x", graph.ParentID("x/y"))
}

func TestSiblingID(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "x/z", graph.SiblingID("x/y", "z"))
}

func TestAreSiblingIDs(t *testing.T) {
	t.Parallel()

	assert.True(t, graph.AreSiblingIDs("x/y", "x/z"))
}

func TestAreSiblingIDsNot(t *testing.T) {
	t.Parallel()

	assert.False(t, graph.AreSiblingIDs("a/b", "x/y"))
}

// TestIsNibling tests various scenarios where we want to know if a node is a
// nibling of the source node.
func TestIsNibling(t *testing.T) {
	t.Parallel()

	t.Run("are siblings", func(t *testing.T) {
		assert.True(t, graph.IsNibling("a/b", "a/c"))
	})
	t.Run("is direct nibling", func(t *testing.T) {
		assert.True(t, graph.IsNibling("a/b", "a/c/d"))
	})
	t.Run("is nibling child of nibling", func(t *testing.T) {
		assert.True(t, graph.IsNibling("a/b", "a/c/d/e"))
	})
	t.Run("unrelated", func(t *testing.T) {
		assert.False(t, graph.IsNibling("a/b", "x/c"))
	})
	t.Run("cousins", func(t *testing.T) {
		assert.False(t, graph.IsNibling("a/b/c", "a/x"))
	})
	t.Run("parent", func(t *testing.T) {
		assert.False(t, graph.IsNibling("a/b/c", "a/b"))
	})

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
