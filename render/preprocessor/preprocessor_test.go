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

package preprocessor_test

import (
	"testing"

	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/graph/node"
	"github.com/asteris-llc/converge/render/preprocessor"
	"github.com/stretchr/testify/assert"
)

// TestInitsWhenEmptySlice ensures that an emptly slice returns an empty slice
func TestInitsWhenEmptySlice(t *testing.T) {
	t.Parallel()
	assert.Nil(t, preprocessor.Inits([]string{}))
}

// TestVertexSplit ensures the correct behavior when stripping off trailing
// fields from a qualified vertex name
func TestVertexSplit(t *testing.T) {
	t.Parallel()
	t.Run("TestVertexSplitWhenMatchingSubstringReturnsPrefixAndRest", func(t *testing.T) {
		s := "a.b.c.d.e"
		g := graph.New()
		g.Add(node.New("a", "a"))
		g.Add(node.New("a.b", "a.b."))
		g.Add(node.New("a.c.d.", "a.c.d"))
		g.Add(node.New("a.b.c", "a.b.c"))
		pfx, rest, found := preprocessor.VertexSplit(g, s)
		assert.Equal(t, "a.b.c", pfx)
		assert.Equal(t, "d.e", rest)
		assert.True(t, found)
	})

	t.Run("TestVertexSplitWhenExactMatchReturnsPrefix", func(t *testing.T) {
		s := "a.b.c"
		g := graph.New()
		g.Add(node.New("a.b.c", "a.b.c"))
		pfx, rest, found := preprocessor.VertexSplit(g, s)
		assert.Equal(t, "a.b.c", pfx)
		assert.Equal(t, "", rest)
		assert.True(t, found)
	})

	t.Run("TestVertexSplitWhenNoMatchReturnsRest", func(t *testing.T) {
		s := "x.y.z"
		g := graph.New()
		g.Add(node.New("a.b.c", "a.b.c"))
		pfx, rest, found := preprocessor.VertexSplit(g, s)
		assert.Equal(t, "", pfx)
		assert.Equal(t, "x.y.z", rest)
		assert.False(t, found)
	})
}
