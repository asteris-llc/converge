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
	"github.com/asteris-llc/converge/helpers/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGet(t *testing.T) {
	// Get should return the value put into the graph
	t.Parallel()

	g := graph.New()
	g.Add("one", 1)

	assert.Equal(t, 1, g.Get("one").(int))
}

func TestGetNothing(t *testing.T) {
	// Get should return nil for a nonexistent node
	t.Parallel()

	g := graph.New()

	assert.Nil(t, g.Get("nothing"))
}

func TestMaybeGet(t *testing.T) {
	t.Parallel()
	g := graph.New()
	g.Add("one", 1)
	val, found := g.MaybeGet("one")
	assert.Equal(t, 1, val)
	assert.True(t, found)
	val, found = g.MaybeGet("two")
	assert.Nil(t, val)
	assert.False(t, found)
}

func TestContains(t *testing.T) {
	t.Parallel()
	g := graph.New()
	g.Add("one", 1)
	assert.True(t, g.Contains("one"))
	assert.False(t, g.Contains("two"))
}

func BenchmarkAddThenGet(b *testing.B) {
	g := graph.New()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			id := rand.Int()
			g.Add(strconv.Itoa(id), id)
			g.Get(strconv.Itoa(id))
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
	g.Add("one", 1)
	g.Remove("one")

	assert.Nil(t, g.Get("one"))
}

func TestDownEdges(t *testing.T) {
	// DownEdges should return string IDs for the downward edges of a given node
	t.Parallel()

	g := graph.New()
	g.Add("one", 1)
	g.Add("two", 2)
	g.Connect("one", "two")

	assert.Equal(t, []string{"two"}, graph.Targets(g.DownEdges("one")))
	assert.Equal(t, 0, len(g.DownEdges("two")))
}

func TestUpEdges(t *testing.T) {
	// UpEdges should return string IDs for the upward edges of a given node
	t.Parallel()

	g := graph.New()
	g.Add("one", 1)
	g.Add("two", 2)
	g.Connect("one", "two")

	assert.Equal(t, []string{"one"}, graph.Sources(g.UpEdges("two")))
	assert.Equal(t, 0, len(g.UpEdges("one")))
}

func TestDisconnect(t *testing.T) {
	t.Parallel()

	g := graph.New()
	g.Add("one", 1)
	g.Add("two", 2)
	g.Connect("one", "two")
	g.Disconnect("one", "two")

	assert.NotContains(t, g.DownEdges("one"), "two")
}

func TestDescendents(t *testing.T) {
	t.Parallel()

	g := graph.New()
	g.Add("one", 1)
	g.Add("one/two", 2)
	g.Connect("one", "one/two")

	assert.Equal(t, []string{"one/two"}, g.Descendents("one"))
}

// TestChildren tests to ensure the correct behavior when getting children
func TestChildren(t *testing.T) {
	defer logging.HideLogs(t)()

	g := graph.New()
	g.Add("root", nil)
	g.Add("child1", nil)
	g.Add("child2", nil)

	g.Add("child1.1", nil)
	g.Add("child1.2", nil)
	g.Add("child1.3", nil)

	g.Add("child.1.1.1", nil)

	g.Add("child2.1", nil)
	g.Add("child2.2", nil)
	g.Add("child2.3", nil)

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
	g.Add("root", nil)
	g.Add("child1", nil)
	g.Add("child2", nil)

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
	g.Add("a", nil)
	g.Add("b", nil)
	g.Add("c", nil)
	g.Add("d", nil)

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
	g.Add("root", 1)
	g.Add("dependent", 1)
	g.Add("dependency", 1)
	g.Add("dependent/child", 1)
	g.Add("dependency/child", 1)

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
			func(id string, _ interface{}) error {
				exLock.Lock()
				defer exLock.Unlock()

				execution = append(execution, id)

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

	g.Add("a", nil)
	g.Add("b", nil)
	g.Add("c", nil)

	g.ConnectParent("a", "b")
	g.ConnectParent("b", "c")

	err := g.Walk(
		context.Background(),
		func(id string, _ interface{}) error {
			if id == "c" {
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
	g.Add("a", nil)
	g.Add("b", nil)
	g.Add("c", nil)

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
	g.Add("a", nil)
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
	g.Add("int", 1)

	transformed, err := g.Transform(
		context.Background(),
		func(id string, dest *graph.Graph) error {
			dest.Add(id, 2)

			return nil
		},
	)

	assert.NoError(t, err)
	assert.Equal(t, 2, transformed.Get("int").(int))
}

func TestParent(t *testing.T) {
	// the graph should return the parent with ID
	t.Parallel()

	g := graph.New()
	g.Add(graph.ID("root"), 1)
	g.Add(graph.ID("root", "child"), 2)

	g.ConnectParent(graph.ID("root"), graph.ID("root", "child"))

	require.NoError(t, g.Validate())

	parent := g.GetParent(graph.ID("root", "child"))
	assert.Equal(t, g.Get(graph.ID("root")), parent)
}

func TestRootFirstWalk(t *testing.T) {
	// the graph should walk nodes root-to-leaf
	defer logging.HideLogs(t)()

	g := graph.New()
	g.Add("root", nil)
	g.Add("root/child", nil)
	g.Connect("root", "root/child")

	var out []string
	assert.NoError(
		t,
		g.RootFirstWalk(
			context.Background(),
			func(id string, _ interface{}) error {
				out = append(out, id)
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
	g.Add("root", nil)
	g.Add("root/child", nil)
	g.Add("root/sibling", nil)

	g.Connect("root", "root/child")
	g.Connect("root", "root/sibling")
	g.Connect("root/child", "root/sibling")

	var out []string
	assert.NoError(
		t,
		g.RootFirstWalk(
			context.Background(),
			func(id string, _ interface{}) error {
				out = append(out, id)
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
	g.Add("int", 1)

	transformed, err := g.RootFirstTransform(
		context.Background(),
		func(id string, dest *graph.Graph) error {
			dest.Add(id, 2)

			return nil
		},
	)

	assert.NoError(t, err)
	assert.Equal(t, 2, transformed.Get("int").(int))
}

func idsInOrderOfExecution(g *graph.Graph) ([]string, error) {
	lock := new(sync.Mutex)
	out := []string{}

	err := g.Walk(
		context.Background(),
		func(id string, _ interface{}) error {
			lock.Lock()
			defer lock.Unlock()

			out = append(out, id)

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
