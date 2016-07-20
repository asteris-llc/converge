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

package mode_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"golang.org/x/net/context"

	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/helpers"
	"github.com/asteris-llc/converge/load"
	"github.com/asteris-llc/converge/parse"
	"github.com/asteris-llc/converge/plan"
	"github.com/asteris-llc/converge/render"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/builtin/file/mode"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTemplateInterface(t *testing.T) {
	t.Parallel()

	assert.Implements(t, (*resource.Task)(nil), new(mode.Mode))
}

func TestParsing(t *testing.T) {
	defer helpers.HideLogs(t)()

	tmpfile, err := ioutil.TempFile("", "mode_test")
	assert.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	module := fmt.Sprintf(
		`file.mode "x" {
    destination = %q
    mode = 777
  }`, tmpfile.Name())

	resourced, err := getResourcesGraph(t, []byte(module))
	assert.NoError(t, err)
	item := resourced.Get("root/file.mode.x")
	preparer, ok := item.(*mode.Preparer)

	require.True(t, ok, fmt.Sprintf("preparer was %T, not *mode.Preparer"), item)
	assert.Equal(t, tmpfile.Name(), preparer.Destination)
	assert.Equal(t, "777", preparer.Mode)
}

func TestPlan(t *testing.T) {
	defer helpers.HideLogs(t)()

	tmpfile, err := ioutil.TempFile("", "mode_test")
	assert.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	module := fmt.Sprintf(
		`file.mode "x" {
		destination = %q
		mode = 777
	}`, tmpfile.Name())

	tmpfile.Write([]byte(module))

	graph, err := load.Load(tmpfile.Name())
	assert.NoError(t, err)
	rendered, err := render.Render(graph, nil)
	assert.NoError(t, err)
	planned, err := plan.Plan(context.Background(), rendered)
	assert.NoError(t, err)

	result := getResult(t, planned, "root/file.mode.x")
	assert.Equal(t, "600", result.Status)
	assert.Equal(t, true, result.WillChange)
}

func getResourcesGraph(t *testing.T, content []byte) (*graph.Graph, error) {
	resources, err := parse.Parse(content)
	require.NoError(t, err)

	g := graph.New()
	g.Add("root", nil)
	for _, resource := range resources {
		id := graph.ID("root", resource.String())
		g.Add(id, resource)
		g.Connect("root", id)
	}
	require.NoError(t, g.Validate())

	return load.SetResources(g)
}

func getResult(t *testing.T, src *graph.Graph, key string) *plan.Result {
	val := src.Get(key)
	result, ok := val.(*plan.Result)
	if !ok {
		t.Logf("needed a %T for %q, got a %T\n", result, key, val)
		t.FailNow()
	}

	return result
}
