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
	"testing"

	"github.com/asteris-llc/converge/load"
	"github.com/asteris-llc/converge/resource"
	"github.com/stretchr/testify/assert"
)

func TestNewGraph(t *testing.T) {
	t.Parallel()

	mod := &resource.Module{
		Resources: []resource.Resource{
			new(resource.ShellTask),
		},
	}

	_, err := load.NewGraph(mod)
	assert.NoError(t, err)
}

func TestGraphWalk(t *testing.T) {
	t.Parallel()

	mod2 := &resource.Module{
		ModuleTask: resource.ModuleTask{
			ModuleName: "test2",
		},
		Resources: []resource.Resource{
			&resource.ShellTask{TaskName: "task2"},
			&resource.Template{TemplateName: "template2"},
		},
	}

	mod := &resource.Module{
		ModuleTask: resource.ModuleTask{
			ModuleName: "test",
		},
		Resources: []resource.Resource{
			&resource.ShellTask{TaskName: "task"},
			&resource.Template{TemplateName: "template"},
			mod2,
		},
	}

	graph, err := load.NewGraph(mod)
	assert.NoError(t, err)

	results := []string{}

	err = graph.Walk(func(path string, r resource.Resource) error {
		results = append(results, path)
		return nil
	})
	assert.NoError(t, err)

	assert.Equal(
		t,
		results,
		[]string{
			"test",
			"test.test2",
			"test.test2.template2",
			"test.test2.task2",
			"test.template",
			"test.task",
		},
	)
}
