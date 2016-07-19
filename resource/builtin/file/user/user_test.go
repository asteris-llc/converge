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

package user_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/helpers"
	"github.com/asteris-llc/converge/load"
	"github.com/asteris-llc/converge/parse"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/builtin/file/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTemplateInterface(t *testing.T) {
	t.Parallel()

	assert.Implements(t, (*resource.Task)(nil), new(user.User))
}

func TestFunctionality(t *testing.T) {
	defer helpers.HideLogs(t)()

	tmpfile, err := ioutil.TempFile("", "user_test")
	assert.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	module := fmt.Sprintf(
		`file.user "x" {
    destination = %q
    user = "nobody"
  }`, tmpfile.Name())
	resourced, err := getResourcesGraph(t, []byte(module))
	assert.NoError(t, err)
	item := resourced.Get("root/file.user.x")
	preparer, ok := item.(*user.Preparer)

	require.True(t, ok, fmt.Sprintf("preparer was %T, not *user.Preparer"), item)
	assert.Equal(t, tmpfile.Name(), preparer.Destination)
	assert.Equal(t, "nobody", preparer.User)
}

//TODO test plan and apply when implemented

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
