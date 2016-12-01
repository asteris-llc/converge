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

package plan_test

import (
	"testing"

	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/graph/node"
	"github.com/asteris-llc/converge/graph/node/conditional"
	"github.com/asteris-llc/converge/helpers/faketask"
	"github.com/asteris-llc/converge/helpers/testing/graphutils"
	"github.com/asteris-llc/converge/parse/preprocessor/switch"
	"github.com/asteris-llc/converge/plan"
	"github.com/asteris-llc/converge/render"
	"github.com/asteris-llc/converge/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"golang.org/x/net/context"
)

// TestResolveConditional tests for handling conditional nodes during planning
func TestResolveConditional(t *testing.T) {
	t.Run("when-non-conditional-node", func(t *testing.T) {
		g := sampleGraph()
		factory, _ := render.NewFactory(context.Background(), g)
		p := plan.Pipeline(context.Background(), g, "root/a", factory)
		meta, _ := g.Get("root/a")
		result, err := p.Exec(context.Background(), meta.Value())
		require.NoError(t, err)
		asPlanResult, ok := result.(*plan.Result)
		require.True(t, ok)
		_, ok = asPlanResult.Task.(*faketask.FakeTask)
		assert.True(t, ok)
	})

	t.Run("when-should-evaluate", func(t *testing.T) {
		g := sampleGraph()
		factory, _ := render.NewFactory(context.Background(), g)
		p := plan.Pipeline(context.Background(), g, "root/a", factory)
		meta, _ := g.Get("root/a")
		graphutils.AddMetadata(g, "root/a", conditional.MetaPredicate, true)
		result, err := p.Exec(context.Background(), meta.Value())
		require.NoError(t, err)
		asPlanResult, ok := result.(*plan.Result)
		require.True(t, ok)
		_, ok = asPlanResult.Task.(*faketask.FakeTask)
		assert.True(t, ok)
	})

	t.Run("when-should-not-evaluate", func(t *testing.T) {
		g := sampleGraph()
		factory, _ := render.NewFactory(context.Background(), g)
		p := plan.Pipeline(context.Background(), g, "root/a", factory)
		meta, _ := g.Get("root/a")
		graphutils.AddMetadata(g, "root/a", conditional.MetaBranchName, "branch1")
		graphutils.AddMetadata(g, "root/a", conditional.MetaPredicate, false)
		result, err := p.Exec(context.Background(), meta.Value())
		require.NoError(t, err)
		asPlanResult, ok := result.(*plan.Result)
		require.True(t, ok)
		_, ok = asPlanResult.Task.(*control.NopTask)
		assert.True(t, ok)
	})
}

// TestExportedFields tests that we put exported fields in the struct
func TestExportedFields(t *testing.T) {
	g := sampleGraph()
	factory, _ := render.NewFactory(context.Background(), g)
	p := plan.Pipeline(context.Background(), g, "root/a", factory)
	meta, _ := g.Get("root/a")
	result, err := p.Exec(context.Background(), meta.Value())
	require.NoError(t, err)
	asPlanResult, ok := result.(*plan.Result)
	require.True(t, ok)
	_, ok = asPlanResult.Task.(*faketask.FakeTask)
	assert.True(t, ok)
	planStatus, ok := asPlanResult.Status.(*resource.Status)
	require.True(t, ok)
	statusMap := planStatus.ExportedFields()
	assert.Equal(t, "status1", statusMap["status"])
}

func sampleGraph() *graph.Graph {
	g := graph.New()
	g.Add(node.New(graph.ID("root"), nil))
	g.Add(node.New(graph.ID("root", "a"), &faketask.FakeTask{Status: "status1"}))
	return g
}
