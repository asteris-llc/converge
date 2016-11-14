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
	"fmt"
	"sort"
	"testing"

	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/graph/node"
	"github.com/asteris-llc/converge/helpers/logging"
	"github.com/asteris-llc/converge/helpers/testing/graphutils"
	"github.com/asteris-llc/converge/load"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNodesBasic tests loading basic.hcl
func TestNodesBasic(t *testing.T) {
	t.Parallel()
	defer logging.HideLogs(t)()

	_, err := load.Nodes(context.Background(), "../samples/basic.hcl", false)
	assert.NoError(t, err)
}

// TestNodesSourceFile tests loading from a source file
func TestNodesSourceFile(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
	defer logging.HideLogs(t)()
	g, err := load.Nodes(context.Background(), "../samples/conditionalLanguages.hcl", false)
	require.NoError(t, err)

	t.Run("creation", func(t *testing.T) {
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
		_, found = g.Get("root/macro.switch.test-switch/macro.case.spanish/file.content.greeting")
		assert.True(t, found)
		_, found = g.Get("root/macro.switch.test-switch/macro.case.french/file.content.greeting")
		assert.True(t, found)
		_, found = g.Get("root/macro.switch.test-switch/macro.case.japanese/file.content.greeting")
		assert.True(t, found)
		_, found = g.Get("root/macro.switch.test-switch/macro.case.default/file.content.greeting")
		assert.True(t, found)
	})

	t.Run("metadata-switch-name", func(t *testing.T) {
		node, found := g.Get("root/macro.switch.test-switch/macro.case.spanish/file.content.greeting")
		require.True(t, found)
		assertMetadataMatches(t, node, "conditional-switch-name", "test-switch")
	})

	t.Run("metadata-predicates", func(t *testing.T) {
		langs := []string{"spanish", "french", "japanese"}
		for _, lang := range langs {
			node, found := g.Get(fmt.Sprintf("root/macro.switch.test-switch/macro.case.%s/file.content.greeting", lang))
			require.True(t, found)
			assertMetadataMatches(t, node, "conditional-predicate-raw", fmt.Sprintf("eq `%s` `{{param `lang`}}`", lang))
		}
	})

	t.Run("metadata-default-predicate", func(t *testing.T) {
		node, found := g.Get("root/macro.switch.test-switch/macro.case.default/file.content.greeting")
		require.True(t, found)
		assertMetadataMatches(t, node, "conditional-predicate-raw", "true")
	})

	t.Run("metadata-name", func(t *testing.T) {
		names := []string{"spanish", "french", "japanese", "default"}
		for _, name := range names {
			node, found := g.Get(fmt.Sprintf("root/macro.switch.test-switch/macro.case.%s/file.content.greeting", name))
			require.True(t, found)
			assertMetadataMatches(t, node, "conditional-name", name)
		}
	})

	t.Run("metadata-peers", func(t *testing.T) {
		names := []string{"spanish", "french", "japanese", "default"}
		expected := []string{"macro.case.spanish", "macro.case.french", "macro.case.japanese", "macro.case.default"}
		for _, name := range names {
			node, found := g.Get(fmt.Sprintf("root/macro.switch.test-switch/macro.case.%s/file.content.greeting", name))
			require.True(t, found)
			assertMetadataMatches(t, node, "conditional-peers", expected)
		}
	})

	t.Run("peer-dependencies", func(t *testing.T) {
		peers := []string{"spanish", "french", "japanese", "default"}
		for idx := 1; idx < len(peers); idx++ {
			me := fmt.Sprintf("root/macro.switch.test-switch/macro.case.%s", peers[idx])
			parent := fmt.Sprintf("root/macro.switch.test-switch/macro.case.%s", peers[idx-1])
			assert.True(t, graphutils.DependsOn(g, me, parent))
		}
	})
}

func assertMetadataMatches(t *testing.T, node *node.Node, key string, expected interface{}) {
	actual, ok := node.LookupMetadata(key)
	assert.True(t, ok)
	assert.Equal(t, expected, actual)
}
