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
	"sort"
	"testing"

	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/helpers/logging"
	"github.com/asteris-llc/converge/load"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNodesBasic tests loading basic.hcl
func TestNodesBasic(t *testing.T) {
	defer logging.HideLogs(t)()

	_, err := load.Nodes(context.Background(), "../samples/basic.hcl", false)
	assert.NoError(t, err)
}

// TestNodesSourceFile tests loading from a source file
func TestNodesSourceFile(t *testing.T) {
	defer logging.HideLogs(t)()

	g, err := load.Nodes(context.Background(), "../samples/sourceFile.hcl", false)
	require.NoError(t, err)

	assertPresent := func(id string) {
		_, ok := g.Get(id)
		assert.True(t, ok, "%q was missing from the graph", id)
	}

	assertPresent("root/param.message")
	assertPresent("root/module.basic")
	assertPresent("root/module.basic/param.message")
	assertPresent("root/module.basic/param.filename")
	assertPresent("root/module.basic/task.render")

	basicDeps := graph.Targets(g.DownEdges("root/module.basic"))
	sort.Strings(basicDeps)

	assert.Equal(
		t,
		[]string{
			"root/module.basic/param.filename",
			"root/module.basic/param.message",
			"root/module.basic/task.render",
		},
		basicDeps,
	)
}

// TestNodeWithConditionals tests loading when switch statements are present
func TestNodeWithConditionals(t *testing.T) {
	defer logging.HideLogs(t)()
	g, err := load.Nodes(context.Background(), "../samples/conditionalLanguages.hcl", false)
	require.NoError(t, err)
	_, found := g.Get("root/param.lang")
	assert.True(t, found)
	_, found = g.Get("root/macro.switch.test-switch")
	assert.True(t, found)
	_, found = g.Get("root/macro.switch.test-switch/macro.case.spanish")
	assert.True(t, found)
	_, found = g.Get("root/macro.switch.test-switch/macro.case.french")
	assert.True(t, found)
	_, found = g.Get("root/macro.switch.test-switch/macro.case.japanese")
	assert.True(t, found)
	_, found = g.Get("root/macro.switch.test-switch/macro.case.default")
	assert.True(t, found)
	_, found = g.Get("root/macro.switch.test-switch/macro.case.spanish/file.content.foo-file")
	assert.True(t, found)
	_, found = g.Get("root/macro.switch.test-switch/macro.case.french/file.content.foo-file")
	assert.True(t, found)
	_, found = g.Get("root/macro.switch.test-switch/macro.case.japanese/file.content.foo-file")
	assert.True(t, found)
	_, found = g.Get("root/macro.switch.test-switch/macro.case.default/file.content.foo-file")
	assert.True(t, found)
}

// TestNodesWithLocks tests that lock nodes are generated on load
func TestNodesWithLocks(t *testing.T) {
	defer logging.HideLogs(t)()

	g, err := load.Nodes(context.Background(), "../samples/locks.hcl", false)
	require.NoError(t, err)

	node, ok := g.Get("root/lock.lock.mylock")
	require.True(t, ok)
	assert.NotNil(t, node)

	node, ok = g.Get("root/lock.unlock.mylock")
	require.True(t, ok)
	assert.NotNil(t, node)
}
