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

package apply_test

import (
	"context"
	"testing"

	"github.com/asteris-llc/converge/apply"
	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/graph/node"
	"github.com/asteris-llc/converge/helpers/faketask"
	"github.com/asteris-llc/converge/plan"
	"github.com/asteris-llc/converge/render"
	"github.com/asteris-llc/converge/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExportedFields tests that we put the exported fields in the struct
func TestExportedFields(t *testing.T) {
	g := sampleGraph()
	factory, _ := render.NewFactory(context.Background(), g)
	p := apply.Pipeline(g, "root/a", factory)
	meta, _ := g.Get("root/a")
	result, err := p.Exec(context.Background(), meta.Value())
	require.NoError(t, err)
	asResult, ok := result.(*apply.Result)
	require.True(t, ok)
	_, ok = asResult.Task.(*faketask.FakeTask)
	assert.True(t, ok)
	applyStatus, ok := asResult.Status.(*resource.Status)
	if !ok {
		t.Logf("unable to convert result status to a resource status, it's type is :: %T\n", asResult.Status)
	}
	require.True(t, ok)
	statusMap := applyStatus.ExportedFields()
	assert.Equal(t, "changed", statusMap["status"])
}

func sampleGraph() *graph.Graph {
	planResult := &plan.Result{
		Task:   faketask.WillChange(),
		Status: &resource.Status{Level: resource.StatusWillChange},
	}
	g := graph.New()
	g.Add(node.New(graph.ID("root"), nil))
	g.Add(node.New(graph.ID("root", "a"), planResult))
	return g
}
