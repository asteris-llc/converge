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
	"log"
	"testing"

	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMergeDuplicatesRemovesDuplicates(t *testing.T) {
	defer helpers.HideLogs(t)()

	g := baseDupGraph()

	transformed, err := graph.MergeDuplicates(context.Background(), g, neverSkip)

	require.NoError(t, err)

	var nodesLeft int
	if transformed.Get("root/first") != nil {
		nodesLeft++
	}
	if transformed.Get("root/one") != nil {
		nodesLeft++
	}

	assert.Equal(t, 1, nodesLeft, "only one of root/first or root/one should remain")
}

func TestMergeDuplicatesMigratesDependencies(t *testing.T) {
	defer helpers.HideLogs(t)()

	g := baseDupGraph()
	g.Add("root/two", 2)
	g.Connect("root", "root/two")
	g.Connect("root/two", "root/first")

	for i := 1; i <= 5; i++ {
		transformed, err := graph.MergeDuplicates(context.Background(), g, neverSkip)

		// we need to get a result where root/first is removed so we can test
		// dependency migration. So if root/first still exists, we need to skip
		if transformed.Get("root/first") != nil {
			log.Printf("[DEBUG] retrying test after failing %d times\n", i)
			continue
		}

		require.NoError(t, err)
		assert.Contains(t, transformed.DownEdges("root/two"), "root/one")

		return
	}

	assert.FailNow(t, "didn't get a testable result in five tries")
}

func TestMergeDuplicatesRemovesChildren(t *testing.T) {
	defer helpers.HideLogs(t)()

	g := baseDupGraph()

	for _, node := range []string{"root/one", "root/first"} {
		g.Add(graph.ID(node, "x"), node)
		g.Connect(node, graph.ID(node, "x"))
	}

	transformed, err := graph.MergeDuplicates(context.Background(), g, neverSkip)
	require.NoError(t, err)

	var removed string
	if transformed.Get("root/one") == nil {
		removed = "root/one"
	} else if transformed.Get("root/first") == nil {
		removed = "root/first"
	} else {
		assert.FailNow(t, `neither "root/one" nor "root/first" was removed`)
	}

	assert.Nil(t, transformed.Get(graph.ID(removed, "x")))
}

func baseDupGraph() *graph.Graph {
	g := graph.New()
	g.Add("root", nil)
	g.Add("root/one", 1)
	g.Add("root/first", 1)

	g.Connect("root", "root/one")
	g.Connect("root", "root/first")

	return g
}

func neverSkip(string) bool {
	return false
}
