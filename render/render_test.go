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

package render_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/graph/node"
	"github.com/asteris-llc/converge/helpers/logging"
	"github.com/asteris-llc/converge/render"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/file/content"
	"github.com/asteris-llc/converge/resource/param"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderSingleNode(t *testing.T) {
	defer logging.HideLogs(t)()

	g := graph.New()
	g.Add(node.New(
		"root/file.content.x",
		resource.NewPreparerWithSource(
			new(content.Preparer),
			map[string]interface{}{
				"destination": "{{1}}",
				"content":     "{{2}}",
			},
		),
	))

	rendered, err := render.Render(context.Background(), g, render.Values{})
	assert.NoError(t, err)

	meta, ok := rendered.Get("root/file.content.x")
	assert.True(t, ok, `"root/file.content.x" was missing from the graph`)

	wrapper, ok := meta.Value().(*resource.TaskWrapper)
	require.True(t, ok, fmt.Sprintf("expected root to be a %T, but it was %T", wrapper, meta.Value()))

	fileContent, ok := wrapper.Task.(*content.Content)
	require.True(t, ok, fmt.Sprintf("expected root to be a %T, but it was %T", fileContent, wrapper.Task))

	assert.Equal(t, "1", fileContent.Destination)
	assert.Equal(t, "2", fileContent.Content)
}

func TestRenderParam(t *testing.T) {
	defer logging.HideLogs(t)()

	g := graph.New()
	g.Add(node.New("root", nil))

	g.Add(node.New(
		"root/file.content.x",
		resource.NewPreparerWithSource(
			new(content.Preparer),
			map[string]interface{}{"destination": "{{param `destination`}}"},
		),
	))

	g.Add(node.New(
		"root/param.destination",
		resource.NewPreparerWithSource(
			new(param.Preparer),
			map[string]interface{}{"default": "1"},
		),
	))

	g.ConnectParent("root", "root/file.content.x")
	g.ConnectParent("root", "root/param.destination")
	g.Connect("root/file.content.x", "root/param.destination")

	rendered, err := render.Render(context.Background(), g, render.Values{})
	require.NoError(t, err)

	meta, ok := rendered.Get("root/file.content.x")
	assert.True(t, ok, `"root/file.content.x" was missing from the graph`)

	wrapper, ok := meta.Value().(*resource.TaskWrapper)
	require.True(t, ok, fmt.Sprintf("expected root to be a %T, but it was %T", wrapper, meta.Value()))

	fileContent, ok := wrapper.Task.(*content.Content)
	require.True(t, ok, fmt.Sprintf("expected root to be a %T, but it was %T", fileContent, wrapper.Task))

	assert.Equal(t, "1", fileContent.Destination)
}

func TestRenderValues(t *testing.T) {
	defer logging.HideLogs(t)()

	g := graph.New()
	g.Add(node.New("root", nil))
	g.Add(node.New(
		"root/file.content.x",
		resource.NewPreparerWithSource(
			new(content.Preparer),
			map[string]interface{}{"destination": "{{param `destination`}}"},
		),
	))
	g.Add(node.New(
		"root/param.destination",
		resource.NewPreparerWithSource(
			new(param.Preparer),
			map[string]interface{}{"default": "1"},
		),
	))

	g.ConnectParent("root", "root/file.content.x")
	g.ConnectParent("root", "root/param.destination")
	g.Connect("root/file.content.x", "root/param.destination")

	rendered, err := render.Render(context.Background(), g, render.Values{"destination": 2})
	require.NoError(t, err)

	meta, ok := rendered.Get("root/file.content.x")
	assert.True(t, ok, `"root/file.content.x" was missing from the graph`)

	wrapper, ok := meta.Value().(*resource.TaskWrapper)
	require.True(t, ok, fmt.Sprintf("expected root to be a %T, but it was a %T", wrapper, meta.Value()))

	content, ok := wrapper.Task.(*content.Content)
	require.True(t, ok, fmt.Sprintf("expected root to be a %T, but it was a %T", content, wrapper.Task))

	assert.Equal(t, "2", content.Destination)
}
