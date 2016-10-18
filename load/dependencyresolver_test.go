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

package load_test

import (
	"context"
	"testing"

	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/helpers/logging"
	"github.com/asteris-llc/converge/load"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDependencyResolverResolvesDependencies(t *testing.T) {
	defer logging.HideLogs(t)()

	nodes, err := load.Nodes(context.Background(), "../samples/basicDependencies.hcl", false)
	require.NoError(t, err)

	resolved, err := load.ResolveDependencies(context.Background(), nodes)
	assert.NoError(t, err)
	assert.Contains(
		t,
		graph.Targets(resolved.DownEdges("root/task.render")),
		"root/task.directory",
	)
}

// TestDependencyResolverWithLocks tests the dependency resolution when nodes
// have a lock
func TestDependencyResolverWithLocks(t *testing.T) {
	defer logging.HideLogs(t)()

	nodes, err := load.Nodes(context.Background(), "../samples/locks.hcl", false)
	require.NoError(t, err)

	resolved, err := load.ResolveDependencies(context.Background(), nodes)
	assert.NoError(t, err)

	t.Run("lock nodes added", func(t *testing.T) {
		// lock and unlock nodes should be added to the graph
		lockNodeID := "root/lock.lock.mylock"
		unlockNodeID := "root/lock.unlock.mylock"
		assert.Contains(t, graph.Targets(resolved.DownEdges("root")), lockNodeID)
		assert.Contains(t, graph.Targets(resolved.DownEdges("root")), unlockNodeID)
	})

	t.Run("locked dependencies", func(t *testing.T) {
		resolved, err = load.ResolveDependenciesInLocks(context.Background(), resolved)
		assert.NoError(t, err)

		// each locked node should have an edge to another locked node or to the lock
		// itself
		lockedIDs := []string{"root/task.lockme1", "root/task.lockme2", "root/task.lockme3"}
		hasLockedEdge := func(id string) bool {
			lockNodeID := "root/lock.lock.mylock"
			edges := graph.Targets(resolved.DownEdges(id))
			for _, lockedID := range lockedIDs {
				for _, edge := range edges {
					if edge == lockedID || edge == lockNodeID {
						return true
					}
				}
			}
			return false
		}

		for _, lockedID := range lockedIDs {
			assert.True(t, hasLockedEdge(lockedID))
		}
	})
}

func TestDependencyResolverBadDependency(t *testing.T) {
	defer logging.HideLogs(t)()

	nodes, err := load.Nodes(context.Background(), "../samples/errors/bad_requirement.hcl", false)
	require.NoError(t, err)

	_, err = load.ResolveDependencies(context.Background(), nodes)
	if assert.Error(t, err) {
		assert.EqualError(t, err, "nonexistent vertices in edges: root/task.nonexistent")
	}
}

func TestDependencyResolverResolvesParam(t *testing.T) {
	defer logging.HideLogs(t)()

	nodes, err := load.Nodes(context.Background(), "../samples/basicDependencies.hcl", false)
	require.NoError(t, err)

	resolved, err := load.ResolveDependencies(context.Background(), nodes)
	assert.NoError(t, err)

	assert.Contains(
		t,
		graph.Targets(resolved.DownEdges("root/task.directory")),
		"root/param.filename",
	)

	assert.Contains(
		t,
		graph.Targets(resolved.DownEdges("root/task.render")),
		"root/param.filename",
	)
	assert.Contains(
		t,
		graph.Targets(resolved.DownEdges("root/task.render")),
		"root/param.message",
	)
}
