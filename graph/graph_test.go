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
	"sync"
	"testing"

	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/helpers"
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

func TestDownEdges(t *testing.T) {
	// DownEdges should return string IDs for the downward edges of a given node
	t.Parallel()

	g := graph.New()
	g.Add("one", 1)
	g.Add("two", 2)
	g.Connect("one", "two")

	assert.Equal(t, []string{"two"}, g.DownEdges("one"))
	assert.Equal(t, 0, len(g.DownEdges("two")))
}

func TestWalkOrder(t *testing.T) {
	// the walk order should start with leaves and head towards the root
	defer helpers.HideLogs(t)()

	g := graph.New()
	g.Add("root", nil)
	g.Add("child1", nil)
	g.Add("child2", nil)

	g.Connect("root", "child1")
	g.Connect("root", "child2")

	out, err := idsInOrderOfExecution(g)

	assert.NoError(t, err)
	assert.Equal(t, "root", out[len(out)-1])
}

func TestWalkOrderDiamond(t *testing.T) {
	/*
		Tree in the form

								a
							 / \
							b		c
							 \ /
								d

		A proper dependency order search would always result in `d` being the first
		element processed
	*/
	defer helpers.HideLogs(t)()

	g := graph.New()
	g.Add("a", nil)
	g.Add("b", nil)
	g.Add("c", nil)
	g.Add("d", nil)

	g.Connect("a", "b")
	g.Connect("b", "d")
	g.Connect("c", "d")

	out, err := idsInOrderOfExecution(g)

	assert.NoError(t, err)
	assert.Equal(t, "d", out[0])
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
		assert.EqualError(t, err, "1 error(s) occurred:\n\n* Cycle: c, b")
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

func TestValidateDifferingTypes(t *testing.T) {
	t.Parallel()

	g := graph.New()
	g.Add("int", 1)
	g.Add("string", "string")

	// just so we don't have multiple roots
	g.Connect("int", "string")

	err := g.Validate()
	if assert.Error(t, err) {
		assert.EqualError(t, err, "differing types in graph vertices: int, string")
	}
}

func TestTransform(t *testing.T) {
	// Transforming in the same type should work
	t.Parallel()

	g := graph.New()
	g.Add("int", 1)

	transformed, err := g.Transform(
		func(id string, _ interface{}, edges []string) (interface{}, []string, error) {
			return 2, edges, nil
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

	g.Connect(graph.ID("root"), graph.ID("root", "child"))

	require.NoError(t, g.Validate())

	parent := g.GetParent(graph.ID("root", "child"))
	assert.Equal(t, g.Get(graph.ID("root")), parent)
}

func idsInOrderOfExecution(g *graph.Graph) ([]string, error) {
	lock := new(sync.Mutex)
	out := []string{}

	err := g.Walk(func(id string, _ interface{}) error {
		lock.Lock()
		defer lock.Unlock()

		out = append(out, id)

		return nil
	})

	return out, err
}
