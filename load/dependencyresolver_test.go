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
