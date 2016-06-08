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
	"sync"
	"testing"

	"github.com/asteris-llc/converge/load"
	"github.com/asteris-llc/converge/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	lock := new(sync.Mutex)

	err = graph.Walk(func(path string, r resource.Resource) error {
		lock.Lock()
		defer lock.Unlock()
		results = append(results, path)
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(
		t,
		[]string{
			"test",
			"test/test2",
			"test/test2/template.template2",
			"test/test2/task.task2",
			"test/template.template",
			"test/task.task",
		},
		results,
	)
}

func TestGraphParent(t *testing.T) {
	t.Parallel()

	mod := &resource.Module{
		ModuleTask: resource.ModuleTask{
			ModuleName: "test",
		},
		Resources: []resource.Resource{
			&resource.ShellTask{TaskName: "task"},
		},
	}

	graph, err := load.NewGraph(mod)
	assert.NoError(t, err)

	parent, err := graph.Parent("test/task")
	assert.NoError(t, err)
	assert.Equal(t, mod, parent)
}

func TestRequirementsOrdering(t *testing.T) {
	/*
		Tree in the form

								a
							 / \
							b		c
							 \ /
							  d

			A proper dependency order search would always result in a being the last
			element processed
	*/
	graph, err := load.Load("../samples/requirementsOrderSmall.hcl", resource.Values{})
	require.NoError(t, err)

	lock := new(sync.Mutex)
	paths := []string{}

	assert.NoError(t, graph.Walk(func(path string, res resource.Resource) error {
		lock.Lock()
		defer lock.Unlock()
		paths = append(paths, path)

		return nil
	}))

	assert.Equal(t, "requirementsOrderSmall.hcl/task.d", paths[1])
	assert.Equal(t, "requirementsOrderSmall.hcl/task.a", paths[len(paths)-1])
}
