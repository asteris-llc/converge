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
	"fmt"
	"testing"

	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/helpers"
	"github.com/asteris-llc/converge/load"
	"github.com/asteris-llc/converge/parse"
	"github.com/asteris-llc/converge/resource/shell"
	"github.com/asteris-llc/converge/resource/template"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetResourcesTask(t *testing.T) {
	defer helpers.HideLogs(t)()

	resourced, err := getResourcesGraph(
		t,
		[]byte(`
task x {
  check = "check"
  apply = "apply"
}`),
	)
	assert.NoError(t, err)

	item := resourced.Get("root/task.x")
	preparer, ok := item.(*shell.Preparer)

	require.True(t, ok, fmt.Sprintf("preparer was %T, not *shell.Preparer", item))
	assert.Equal(t, "check", preparer.Check)
	assert.Equal(t, "apply", preparer.Apply)
}

func TestSetResourcesTemplate(t *testing.T) {
	defer helpers.HideLogs(t)()

	resourced, err := getResourcesGraph(
		t,
		[]byte(`
template x {
  destination = "destination"
  content = "content"
}
`),
	)
	assert.NoError(t, err)

	item := resourced.Get("root/template.x")
	preparer, ok := item.(*template.Preparer)

	require.True(t, ok, fmt.Sprintf("preparer was %T, not *template.Preparer", item))
	assert.Equal(t, "destination", preparer.Destination)
	assert.Equal(t, "content", preparer.Content)
}

func TestSetResourcesBad(t *testing.T) {
	defer helpers.HideLogs(t)()

	_, err := getResourcesGraph(t, []byte("x x {}"))
	if assert.Error(t, err) {
		assert.EqualError(t, err, "1 error(s) occurred:\n\n* \"x\" is not a valid resource type in \"x.x\"")
	}
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
