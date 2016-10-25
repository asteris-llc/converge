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
	"context"
	"errors"
	"math/rand"
	"sort"

	"strconv"
	"sync"
	"testing"

	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/graph/node"
	"github.com/asteris-llc/converge/helpers/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGet(t *testing.T) {
	t.Parallel()

	g := graph.New()
	g.Add(node.New("one", 1))

	t.Run("found", func(t *testing.T) {
		val, found := g.Get("one")
		assert.Equal(t, 1, val.Value())
		assert.True(t, found)
	})

	t.Run("not found", func(t *testing.T) {
		val, found := g.Get("two")
		assert.False(t, found)
		assert.Nil(t, val)
	})
}

func TestContains(t *testing.T) {
	t.Parallel()
	g := graph.New()
	g.Add(node.New("one", 1))
	assert.True(t, g.Contains("one"))
	assert.False(t, g.Contains("two"))
}

func BenchmarkAddThenGet(b *testing.B) {
	g := graph.New()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			id := strconv.Itoa(rand.Int())
			g.Add(node.New(id, id))
			g.Get(id)
		}
	})
}

func BenchmarkCopyParallel(b *testing.B) {
	g := graph.New()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			g.Copy()
		}
	})
}

func TestRemove(t *testing.T) {
	// Remove should remove a vertex
	t.Parallel()

	g := graph.New()
	g.Add(node.New("one", 1))
	g.Remove("one")

	_, ok := g.Get("one")
	assert.False(t, ok)
}

func TestDownEdges(t *testing.T) {
	// DownEdges should return string IDs for the downward edges of a given node
	t.Parallel()

	g := graph.New()
	g.Add(node.New("one", 1))
	g.Add(node.New("two", 2))
	g.Connect("one", "two")

	assert.Equal(t, []string{"two"}, graph.Targets(g.DownEdges("one")))
	assert.Equal(t, 0, len(g.DownEdges("two")))
}

// TestDownEdgesInGroup tests that DownEdgesInGroup returns only edges with
// values in a group
func TestDownEdgesInGroup(t *testing.T) {
	t.Parallel()

	group := "group"
	g := graph.New()
	g.Add(newGroupNode("one", group, 1))
	g.Add(newGroupNode("two", group, 2))
	g.Add(node.New("three", 3))
	g.Connect("one", "two")
	g.Connect("one", "three")

	assert.Equal(t, []string{"two"}, g.DownEdgesInGroup("one", group))
	assert.Equal(t, 0, len(g.DownEdges("two")))
}

func TestUpEdges(t *testing.T) {
	// UpEdges should return string IDs for the upward edges of a given node
	t.Parallel()

	g := graph.New()
	g.Add(node.New("one", 1))
	g.Add(node.New("two", 2))
	g.Connect("one", "two")

	assert.Equal(t, []string{"one"}, graph.Sources(g.UpEdges("two")))
	assert.Equal(t, 0, len(g.UpEdges("one")))
}

// TestUpEdgesInGroup tests that UpEdgesInGroup returns only edges with values
// in a group
func TestUpEdgesInGroup(t *testing.T) {
	t.Parallel()

	group := "test"
	g := graph.New()
	g.Add(newGroupNode("one", group, 1))
	g.Add(newGroupNode("two", group, 2))
	g.Add(node.New("three", 2)) // no group
	g.Connect("one", "two")
	g.Connect("three", "two")

	assert.Equal(t, []string{"one"}, g.UpEdgesInGroup("two", group))
}

// TestGroupNodes tests that GroupNodes only returns nodes in a specific group
func TestGroupNodes(t *testing.T) {
	t.Parallel()

	group := "test"
	newNode := func(id string, val int) *node.Node {
		n := node.New(id, val)
		n.Group = group
		return n
	}

	g := graph.New()
	g.Add(newNode("one", 1))
	g.Add(newNode("two", 2))
	g.Add(node.New("three", 2)) // no group
	g.Connect("one", "two")
	g.Connect("three", "two")

	assert.Equal(t, []string{"one"}, g.UpEdgesInGroup("two", group))
}

// TestSafeConnect tests that calling SafeConnect on a an invalid graph will
// return an error
func TestSafeConnect(t *testing.T) {
	t.Parallel()
	t.Run("valid", func(t *testing.T) {
		g := graph.New()
		g.Add(node.New("a", nil))
		g.Add(node.New("b", nil))
		err := g.SafeConnect("a", "b")

		assert.NoError(t, err)
	})

	t.Run("invalid", func(t *testing.T) {
		g := invalidGraph()
		g.Add(node.New("a", nil))
		g.Add(node.New("b", nil))
		err := g.SafeConnect("a", "b")

		assert.Error(t, err)
	})
}

func TestDisconnect(t *testing.T) {
	t.Parallel()

	g := graph.New()
	g.Add(node.New("one", 1))
	g.Add(node.New("two", 2))
	g.Connect("one", "two")
	g.Disconnect("one", "two")

	assert.NotContains(t, g.DownEdges("one"), "two")
}

// TestSafeDisconnect tests that calling SafeDisconnect on a an invalid graph
// will return an error
func TestSafeDisconnect(t *testing.T) {
	t.Parallel()
	t.Run("valid", func(t *testing.T) {
		g := graph.New()
		g.Add(node.New("one", 1))
		g.Add(node.New("two", 2))
		g.Add(node.New("three", 2))
		g.Connect("one", "two")
		g.Connect("one", "three")
		g.Connect("two", "three")

		err := g.SafeDisconnect("two", "three")
		assert.NoError(t, err)
	})

	t.Run("invalid", func(t *testing.T) {
		g := invalidGraph()
		g.Add(node.New("one", 1))
		g.Add(node.New("two", 2))
		g.Add(node.New("three", 2))
		g.Connect("one", "two")
		g.Connect("one", "three")
		g.Connect("two", "three")
		err := g.SafeDisconnect("two", "three")

		assert.Error(t, err)
	})
}

func TestDescendents(t *testing.T) {
	t.Parallel()

	g := graph.New()
	g.Add(node.New("one", 1))
	g.Add(node.New("one/two", 2))
	g.Connect("one", "one/two")

	assert.Equal(t, []string{"one/two"}, g.Descendents("one"))
}

// TestChildren tests to ensure the correct behavior when getting children
func TestChildren(t *testing.T) {
	t.Parallel()
	defer logging.HideLogs(t)()

	g := graph.New()
	g.Add(node.New("root", nil))
	g.Add(node.New("child1", nil))
	g.Add(node.New("child2", nil))

	g.Add(node.New("child1.1", nil))
	g.Add(node.New("child1.2", nil))
	g.Add(node.New("child1.3", nil))

	g.Add(node.New("child.1.1.1", nil))

	g.Add(node.New("child2.1", nil))
	g.Add(node.New("child2.2", nil))
	g.Add(node.New("child2.3", nil))

	g.ConnectParent("root", "child1")
	g.ConnectParent("root", "child2")

	g.ConnectParent("child1", "child1.1")
	g.ConnectParent("child1", "child1.2")
	g.ConnectParent("child1", "child1.3")

	g.ConnectParent("child2", "child2.1")
	g.ConnectParent("child2", "child2.2")
	g.ConnectParent("child2", "child2.3")

	g.ConnectParent("child1.1", "child.1.1.1")

	g.Connect("child1", "child2.1")
	g.Connect("child1", "child2.2")
	g.Connect("child1", "child2.3")

	children := g.Children("child1")

	expected := []string{"child1.1", "child1.2", "child1.3"}
	sort.Strings(expected)
	sort.Strings(children)
	assert.Equal(t, expected, children)
}

func TestWalkOrder(t *testing.T) {
	// the walk order should start with leaves and head towards the root
	defer logging.HideLogs(t)()

	g := graph.New()
	g.Add(node.New("root", nil))
	g.Add(node.New("child1", nil))
	g.Add(node.New("child2", nil))

	g.ConnectParent("root", "child1")
	g.ConnectParent("root", "child2")

	out, err := idsInOrderOfExecution(g)

	assert.NoError(t, err)
	assert.Equal(t, "root", out[len(out)-1])
}

func TestWalkOrderDiamond(t *testing.T) {
	/*
		Tree in the form

		|    a    |
		|   / \   |
		|  b   c  |
		|   \ /   |
		|    d    |

		A proper dependency order search would always result in `d` being the first
		element processed
	*/
	defer logging.HideLogs(t)()

	g := graph.New()
	g.Add(node.New("a", nil))
	g.Add(node.New("b", nil))
	g.Add(node.New("c", nil))
	g.Add(node.New("d", nil))

	g.ConnectParent("a", "b")
	g.ConnectParent("a", "c")
	g.ConnectParent("b", "d")
	g.ConnectParent("c", "d")

	out, err := idsInOrderOfExecution(g)

	assert.NoError(t, err)
	if assert.True(t, len(out) > 3, "out was %s", out) {
		assert.Equal(t, "a", out[3])
		assert.Equal(t, "d", out[0])
	}
}

func TestWalkOrderParentDependency(t *testing.T) {
	defer logging.HideLogs(t)()

	g := graph.New()
	g.Add(node.New("root", 1))
	g.Add(node.New("dependent", 1))
	g.Add(node.New("dependency", 1))
	g.Add(node.New("dependent/child", 1))
	g.Add(node.New("dependency/child", 1))

	g.ConnectParent("root", "dependent")
	g.ConnectParent("root", "dependency")
	g.ConnectParent("dependent", "dependent/child")
	g.ConnectParent("dependency", "dependency/child")

	g.Connect("dependent", "dependency")

	exLock := new(sync.Mutex)
	var execution []string

	require.NoError(t,
		g.Walk(
			context.Background(),
			func(meta *node.Node) error {
				exLock.Lock()
				defer exLock.Unlock()

				execution = append(execution, meta.ID)

				return nil
			},
		),
	)

	assert.Equal(
		t,
		[]string{
			"dependency/child",
			"dependency",
			"dependent/child",
			"dependent",
			"root",
		},
		execution,
	)
}

func TestWalkError(t *testing.T) {
	g := graph.New()

	g.Add(node.New("a", nil))
	g.Add(node.New("b", nil))
	g.Add(node.New("c", nil))

	g.ConnectParent("a", "b")
	g.ConnectParent("b", "c")

	err := g.Walk(
		context.Background(),
		func(meta *node.Node) error {
			if meta.ID == "c" {
				return errors.New("test")
			}
			return nil
		},
	)

	if assert.Error(t, err) {
		assert.EqualError(t, err, "1 error(s) occurred:\n\n* c: test")
	}
}

func TestValidateNoRoot(t *testing.T) {
	// Validate should error if there is no root
	t.Parallel()

	g := graph.New()

	err := g.Validate()
	if assert.Error(t, err) {
		assert.EqualError(t, err, "no roots found")
	}
}

func TestValidateCycle(t *testing.T) {
	// Validate should error if there is a cyle
	t.Parallel()

	g := graph.New()
	g.Add(node.New("a", nil))
	g.Add(node.New("b", nil))
	g.Add(node.New("c", nil))

	// a is just a root
	g.Connect("a", "b")

	// now the cycle
	g.Connect("b", "c")
	g.Connect("c", "b")

	err := g.Validate()
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "1 error(s) occurred:\n\n* Cycle: ")
	}
}

func TestValidateDanglingEdge(t *testing.T) {
	t.Parallel()

	g := graph.New()
	g.Add(node.New("a", nil))
	g.Connect("a", "nope")

	err := g.Validate()
	if assert.Error(t, err) {
		assert.EqualError(t, err, "nonexistent vertices in edges: nope")
	}
}

func TestTransform(t *testing.T) {
	// Transforming in the same type should work
	defer logging.HideLogs(t)()

	g := graph.New()
	g.Add(node.New("int", 1))

	transformed, err := g.Transform(
		context.Background(),
		func(meta *node.Node, dest *graph.Graph) error {
			dest.Add(node.New(meta.ID, 2))

			return nil
		},
	)

	assert.NoError(t, err)
	meta, ok := transformed.Get("int")
	require.True(t, ok, "node was not present in graph")
	assert.Equal(t, 2, meta.Value().(int))
}

func TestParent(t *testing.T) {
	// the graph should return the parent with ID
	t.Parallel()

	g := graph.New()
	g.Add(node.New(graph.ID("root"), 1))
	g.Add(node.New(graph.ID("root", "child"), 2))

	g.ConnectParent(graph.ID("root"), graph.ID("root", "child"))

	require.NoError(t, g.Validate())

	actual, ok := g.GetParent(graph.ID("root", "child"))
	require.True(t, ok)
	should, _ := g.Get(graph.ID("root"))
	assert.Equal(t, should, actual)
}

func TestParentID(t *testing.T) {
	t.Parallel()
	t.Run("valid", func(t *testing.T) {
		parentID := graph.ID("root")
		childID := graph.ID("root", "child")
		g := graph.New()
		g.Add(node.New(parentID, 1))
		g.Add(node.New(childID, 2))
		g.ConnectParent(parentID, childID)

		actualParentID, ok := g.GetParentID(childID)
		require.True(t, ok)
		assert.Equal(t, parentID, actualParentID)
	})
	t.Run("root", func(t *testing.T) {
		g := graph.New()
		g.Add(node.New("root", 1))
		_, ok := g.GetParentID("root")
		assert.False(t, ok)
	})
	t.Run("one-two-three", func(t *testing.T) {
		g := graph.New()
		g.Add(node.New("one", 1))
		g.Add(node.New("two", 2))
		g.Add(node.New("three", 2))
		g.Connect("one", "two")
		g.Connect("one", "three")
		g.Connect("two", "three")
		_, ok := g.GetParentID("one")
		assert.False(t, ok)
		_, ok = g.GetParentID("two")
		assert.False(t, ok)
		_, ok = g.GetParentID("three")
		assert.False(t, ok)
	})
}

func TestRootFirstWalk(t *testing.T) {
	// the graph should walk nodes root-to-leaf
	defer logging.HideLogs(t)()

	g := graph.New()
	g.Add(node.New("root", nil))
	g.Add(node.New("root/child", nil))
	g.ConnectParent("root", "root/child")

	var out []string
	assert.NoError(
		t,
		g.RootFirstWalk(
			context.Background(),
			func(meta *node.Node) error {
				out = append(out, meta.ID)
				return nil
			},
		),
	)

	assert.Equal(t, []string{"root", "root/child"}, out)
}

func TestRootFirstWalkSiblingDep(t *testing.T) {
	// the graph should resolve sibling dependencies before their dependers
	defer logging.HideLogs(t)()

	g := graph.New()
	g.Add(node.New("root", nil))
	g.Add(node.New("root/child", nil))
	g.Add(node.New("root/sibling", nil))

	g.ConnectParent("root", "root/child")
	g.ConnectParent("root", "root/sibling")
	g.Connect("root/child", "root/sibling")

	var out []string
	assert.NoError(
		t,
		g.RootFirstWalk(
			context.Background(),
			func(meta *node.Node) error {
				out = append(out, meta.ID)
				return nil
			},
		),
	)

	assert.Equal(
		t,
		[]string{"root", "root/sibling", "root/child"},
		out,
	)
}

func TestRootFirstTransform(t *testing.T) {
	// transforming depth first should work
	defer logging.HideLogs(t)()

	g := graph.New()
	g.Add(node.New("int", 1))

	transformed, err := g.RootFirstTransform(
		context.Background(),
		func(meta *node.Node, dest *graph.Graph) error {
			dest.Add(meta.WithValue(2))

			return nil
		},
	)

	assert.NoError(t, err)
	meta, ok := transformed.Get("int")
	require.True(t, ok, "\"int\" was not present in graph")
	assert.Equal(t, 2, meta.Value().(int))
}

// TestIsNibling tests various scenarios where we want to know if a node is a
// nibling of the source node.
func TestIsNibling(t *testing.T) {
	t.Parallel()

	g := graph.New()
	g.Add(node.New("a", struct{}{}))
	g.Add(node.New("a/b", struct{}{}))
	g.ConnectParent("a", "a/b")
	g.Add(node.New("a/b/c", struct{}{}))
	g.ConnectParent("a/b", "a/b/c")
	g.Add(node.New("a/b/c/d", struct{}{}))
	g.ConnectParent("a/b/c", "a/b/c/d")
	g.Add(node.New("a/c", struct{}{}))
	g.ConnectParent("a", "a/c")
	g.Add(node.New("a/c/d", struct{}{}))
	g.ConnectParent("a/c", "a/c/d")
	g.Add(node.New("a/c/d/e", struct{}{}))
	g.ConnectParent("a/c/d", "a/c/d/e")
	g.Add(node.New("x", struct{}{}))
	g.Add(node.New("x/c", struct{}{}))
	g.ConnectParent("x", "x/c")

	t.Run("are siblings", func(t *testing.T) {
		assert.True(t, g.IsNibling("a/b", "a/c"))
	})
	t.Run("is direct nibling", func(t *testing.T) {
		assert.True(t, g.IsNibling("a/b", "a/c/d"))
	})
	t.Run("is nibling child of nibling", func(t *testing.T) {
		assert.True(t, g.IsNibling("a/b", "a/c/d/e"))
	})
	t.Run("child", func(t *testing.T) {
		assert.False(t, g.IsNibling("a/b", "a/b/c"))
	})
	t.Run("grandchild", func(t *testing.T) {
		assert.False(t, g.IsNibling("a/b", "a/b/c/d"))
	})
	t.Run("unrelated", func(t *testing.T) {
		assert.False(t, g.IsNibling("a/b", "x/c"))
	})
	t.Run("cousins", func(t *testing.T) {
		assert.False(t, g.IsNibling("a/b/c", "a/x"))
	})
	t.Run("parent", func(t *testing.T) {
		assert.False(t, g.IsNibling("a/b/c", "a/b"))
	})
}

func idsInOrderOfExecution(g *graph.Graph) ([]string, error) {
	lock := new(sync.Mutex)
	out := []string{}

	err := g.Walk(
		context.Background(),
		func(meta *node.Node) error {
			lock.Lock()
			defer lock.Unlock()

			out = append(out, meta.ID)

			return nil
		},
	)

	return out, err
}

func invalidGraph() *graph.Graph {
	g := graph.New()
	g.Connect("Bad", "Nodes")
	return g
}

func newGroupNode(id, group string, val interface{}) *node.Node {
	n := node.New(id, val)
	n.Group = group
	return n
}
