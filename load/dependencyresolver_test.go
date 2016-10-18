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

		assert.Empty(t, graph.Targets(resolved.DownEdges("root/task.install-build-essential")))
		assert.Equal(
			t,
			[]string{"root/task.install-build-essential"},
			graph.Targets(resolved.DownEdges("root/task.install-jq")),
		)
		assert.Equal(
			t,
			[]string{"root/task.install-jq"},
			graph.Targets(resolved.DownEdges("root/task.install-tree")),
		)
	})

	t.Run("inter-module", func(t *testing.T) {
		nodes, err := load.Nodes(context.Background(), "../samples/groupedIncludeModule.hcl", false)
		require.NoError(t, err)

		resolved, err := load.ResolveDependencies(context.Background(), nodes)
		assert.NoError(t, err)

		// first module is not dependent on other modules
		assert.NotContains(t, resolved.Dependencies("root/module.test1"), "root/module.test2")
		assert.NotContains(t, resolved.Dependencies("root/module.test1"), "root/module.test3")

		// other modules should depend on each other
		assert.Contains(t, resolved.Dependencies("root/module.test2"), "root/module.test1")
		assert.Contains(t, resolved.Dependencies("root/module.test3"), "root/module.test2")
	})
}
