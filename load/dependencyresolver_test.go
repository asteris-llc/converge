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
	"testing"

	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/helpers/logging"
	"github.com/asteris-llc/converge/load"
	"github.com/asteris-llc/converge/parse"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
)

// TestDependencyResolverResolvesDependencies tests dependency resolution
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

// TestDependencyResolverResolvesExplicitDepsInBranch tests explicit
// dependencies inside of case branch nodes
func TestDependencyResolverResolvesExplicitDepsInBranch(t *testing.T) {
	defer logging.HideLogs(t)()

	nodes, err := load.Nodes(context.Background(), "../samples/conditionalDeps.hcl", false)
	require.NoError(t, err)

	resolved, err := load.ResolveDependencies(context.Background(), nodes)
	assert.NoError(t, err)
	assert.Contains(
		t,
		graph.Targets(resolved.DownEdges("root/macro.switch.sample/macro.case.true/file.content.foo-output")),
		"root/task.query.foo",
	)
}

func TestDependencyResolverBadDependency(t *testing.T) {
	defer logging.HideLogs(t)()

	nodes, err := load.Nodes(context.Background(), "../samples/errors/bad_requirement.hcl", false)
	require.NoError(t, err)

	_, err = load.ResolveDependencies(context.Background(), nodes)
	if assert.Error(t, err) {
		assert.EqualError(t, err, "1 error(s) occurred:\n\n* root/task.bad_requirement: nonexistent vertices in edges: task.nonexistent")
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

// TestDependencyResolverResolvesGroupDependencies tests whether group
// dependencies are wired correctly
func TestDependencyResolverResolvesGroupDependencies(t *testing.T) {
	t.Parallel()
	defer logging.HideLogs(t)()

	t.Run("intra-module", func(t *testing.T) {
		nodes, err := load.Nodes(context.Background(), "../samples/groups.hcl", false)
		require.NoError(t, err)

		resolved, err := load.ResolveDependencies(context.Background(), nodes)
		assert.NoError(t, err)

		group := "apt"
		groupNodes := resolved.GroupNodes(group)
		assert.NotEmpty(t, groupNodes)
		for _, node := range groupNodes {
			assert.True(t, len(resolved.DownEdgesInGroup(node.ID, group)) <= 1)
			assert.True(t, len(resolved.UpEdgesInGroup(node.ID, group)) <= 1)

			// find the highest node
			if len(resolved.UpEdges(node.ID)) == 1 {
				// it should depend on the other nodes
				assert.Equal(t, 2, len(resolved.Dependencies(node.ID)))
			}

		}
	})

	t.Run("inter-module", func(t *testing.T) {
		nodes, err := load.Nodes(context.Background(), "../samples/groupedIncludeModule.hcl", false)
		require.NoError(t, err)

		resolved, err := load.ResolveDependencies(context.Background(), nodes)
		assert.NoError(t, err)

		group := "groupedModule"
		groupNodes := resolved.GroupNodes(group)
		assert.NotEmpty(t, groupNodes)
		for _, node := range groupNodes {
			moduleID := graph.ParentID(node.ID)
			assert.True(t, len(resolved.DownEdgesInGroup(moduleID, group)) <= 1)
			assert.True(t, len(resolved.UpEdgesInGroup(moduleID, group)) <= 1)

			// find the highest node
			if len(resolved.UpEdges(moduleID)) == 1 {
				// it should depend on the other modules
				var moduleDeps []string
				for _, depID := range resolved.Dependencies(moduleID) {
					if dep, ok := resolved.Get(depID); ok {
						depNode, ok := dep.Value().(*parse.Node)
						if ok && depNode.IsModule() {
							moduleDeps = append(moduleDeps, dep.ID)
						}
					}
				}
				assert.Equal(t, 2, len(moduleDeps))
			}
		}
	})
}
