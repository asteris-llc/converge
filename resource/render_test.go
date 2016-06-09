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

package resource_test

import (
	"testing"

	"github.com/asteris-llc/converge/resource"
	"github.com/stretchr/testify/assert"
)

func TestRendererRenderValid(t *testing.T) {
	t.Parallel()

	mod := &resource.Module{
		ModuleTask: resource.ModuleTask{ModuleName: "test"},
	}
	renderer, err := resource.NewRenderer(mod)
	assert.NoError(t, err)

	result, err := renderer.Render("", "{{.String}}")
	assert.NoError(t, err)
	assert.Equal(t, mod.String(), result)
}

func TestRendererRenderInvalid(t *testing.T) {
	t.Parallel()

	mod := &resource.Module{
		ModuleTask: resource.ModuleTask{ModuleName: "test"},
	}
	renderer, err := resource.NewRenderer(mod)
	assert.NoError(t, err)

	_, err = renderer.Render("", "{{")
	if assert.Error(t, err) {
		assert.EqualError(t, err, "template: :1: unexpected unclosed action in command")
	}
}

func TestRendererRenderParam(t *testing.T) {
	t.Parallel()

	param := &resource.Param{
		Name:    "test_parameter",
		Default: "test_default",
	}
	mod := &resource.Module{
		ModuleTask: resource.ModuleTask{
			ModuleName: "test_module",
			Args:       map[string]resource.Value{"test_parameter": "test_value"},
		},
		Resources: []resource.Resource{param},
	}

	err := param.Prepare(mod)
	assert.NoError(t, err)

	renderer, err := resource.NewRenderer(mod)
	assert.NoError(t, err)

	result, err := renderer.Render("test", "{{param `test_parameter`}}")
	assert.NoError(t, err)
	assert.EqualValues(t, "test_value", result)
}

func TestRenderMissingParam(t *testing.T) {
	t.Parallel()

	renderer, err := resource.NewRenderer(&resource.Module{})
	assert.NoError(t, err)

	_, err = renderer.Render("x", "{{param `nonexistent`}}")
	if assert.Error(t, err) {
		assert.EqualError(t, err, "template: x:1:2: executing \"x\" at <param `nonexistent`>: error calling param: no such param \"nonexistent\"")
	}
}
