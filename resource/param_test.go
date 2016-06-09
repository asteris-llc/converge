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

func TestParamPreparePresent(t *testing.T) {
	t.Parallel()

	var (
		p = &resource.Param{
			Name:    "test",
			Default: "x",
			Type:    "string",
		}

		value = "a"

		m = &resource.Module{
			ModuleTask: resource.ModuleTask{
				Args: map[string]resource.Value{p.Name: "a"},
			},
			Resources: []resource.Resource{p},
		}
	)

	err := p.Prepare(m)
	assert.NoError(t, err)

	assert.EqualValues(t, value, p.Value())
}

func TestParamPrepareDefault(t *testing.T) {
	t.Parallel()

	var (
		p = &resource.Param{
			Name:    "test",
			Default: "x",
			Type:    "string",
		}

		m = &resource.Module{
			Resources: []resource.Resource{p},
		}
	)

	err := p.Prepare(m)
	assert.NoError(t, err)

	assert.EqualValues(t, p.Default, p.Value())
}
