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
	"testing"

	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/graph/node"
	"github.com/asteris-llc/converge/helpers/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMergeDuplicatesRemovesDuplicates(t *testing.T) {
	defer logging.HideLogs(t)()

	g := baseDupGraph()

	transformed, err := graph.MergeDuplicates(context.Background(), g, neverSkip)

	require.NoError(t, err)

	var nodesLeft int
	if _, ok := transformed.Get("root/first"); ok {
		nodesLeft++
	}
	if _, ok := transformed.Get("root/one"); ok {
		nodesLeft++
	}

	assert.Equal(t, 1, nodesLeft, "only one of root/first or root/one should remain")
}

func TestMergeDuplicatesMigratesDependencies(t *testing.T) {
	defer logging.HideLogs(t)()

	g := baseDupGraph()
	g.Add(node.New("root/two", 2))
	g.ConnectParent("root", "root/two")
	g.Connect("root/two", "root/first")

	for i := 1; i <= 5; i++ {
		transformed, err := graph.MergeDuplicates(context.Background(), g, neverSkip)

		// we need to get a result where root/first is removed so we can test
		// dependency migration. So if root/first still exists, we need to skip
		if _, ok := transformed.Get("root/first"); ok {
			t.Logf("retrying test after failing %d times", i)
			continue
		}

		require.NoError(t, err)
		assert.Contains(t, graph.Targets(transformed.DownEdges("root/two")), "root/one")

		return
	}

	assert.FailNow(t, "didn't get a testable result in five tries")
}

func TestMergeDuplicatesRemovesChildren(t *testing.T) {
	defer logging.HideLogs(t)()

	g := baseDupGraph()

	for _, id := range []string{"root/one", "root/first"} {
		g.Add(node.New(graph.ID(id, "x"), id))
		g.Connect(id, graph.ID(id, "x"))
	}

	transformed, err := graph.MergeDuplicates(context.Background(), g, neverSkip)
	require.NoError(t, err)

	var removed string
	if _, ok := transformed.Get("root/one"); !ok {
		removed = "root/one"
	} else if _, ok := transformed.Get("root/first"); !ok {
		removed = "root/first"
	} else {
		assert.FailNow(t, `neither "root/one" nor "root/first" was removed`)
	}

	_, ok := transformed.Get(graph.ID(removed, "x"))
	assert.False(t, ok, "%q was still present", graph.ID(removed, "x"))
}

func baseDupGraph() *graph.Graph {
	g := graph.New()
	g.Add(node.New("root", nil))
	g.Add(node.New("root/one", 1))
	g.Add(node.New("root/first", 1))

	g.ConnectParent("root", "root/one")
	g.ConnectParent("root", "root/first")

	return g
}

func neverSkip(*node.Node) bool {
	return false
}
