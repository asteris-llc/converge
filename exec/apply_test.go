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

package exec_test

import (
	"errors"
	"fmt"
	"testing"

	"golang.org/x/net/context"

	"github.com/asteris-llc/converge/exec"
	"github.com/asteris-llc/converge/helpers"
	"github.com/asteris-llc/converge/load"
	"github.com/asteris-llc/converge/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApply(t *testing.T) {
	defer (helpers.HideLogs(t))()

	graph, err := load.Load("../samples/basic.hcl", resource.Values{})
	require.NoError(t, err)

	plan, err := exec.Plan(context.Background(), graph)
	assert.NoError(t, err)

	helpers.InTempDir(t, func() {
		results, err := exec.Apply(context.Background(), graph, plan)
		assert.NoError(t, err)
		assert.Equal(
			t,
			[]*exec.ApplyResult{{
				Path:      "module.basic.hcl/task.render",
				OldStatus: "cat: test.txt: No such file or directory\n",
				NewStatus: "Hello, World!\n",
				Success:   true,
			}},
			results,
		)
	})
}

// if a task's dependency fails due to an error, that task shouldn't run
func TestBlockingTaskError(t *testing.T) {
	defer (helpers.HideLogs(t))()

	task1 := &helpers.DummyTask{
		Name:       "fail",
		Change:     true,
		ApplyError: errors.New("failed applying dummy task"),
	}
	task2 := &helpers.DummyTask{Name: "dont_run", Change: true}
	task2.SetDepends([]string{"dummy_task.fail"})

	mod := &resource.Module{
		ModuleTask: resource.ModuleTask{ModuleName: "test_module"},
		Resources:  []resource.Resource{task1, task2},
	}

	graph, err := load.NewGraph(mod)
	assert.NoError(t, err)

	plan, err := exec.Plan(context.Background(), graph)
	assert.NoError(t, err)

	helpers.InTempDir(t, func() {
		results, err := exec.Apply(context.Background(), graph, plan)
		assert.Error(t, err)     // the first task should return an error
		assert.Empty(t, results) // the second task shouldn't be run
	})
}

// if a task's dependency fails because Check still reports WillChange, that
// task shouldn't run
func TestBlockingTaskFailure(t *testing.T) {
	defer (helpers.HideLogs(t))()

	task1 := &helpers.DummyTask{
		Name:             "fail",
		Change:           true,
		ChangeAfterApply: true, // even after applying, Check will report willChange
	}
	task2 := &helpers.DummyTask{Name: "dont_run", Change: true}
	task2.SetDepends([]string{"dummy_task.fail"})

	mod := &resource.Module{
		ModuleTask: resource.ModuleTask{ModuleName: "test_module"},
		Resources:  []resource.Resource{task1, task2},
	}

	graph, err := load.NewGraph(mod)
	assert.NoError(t, err)

	plan, err := exec.Plan(context.Background(), graph)
	assert.NoError(t, err)

	helpers.InTempDir(t, func() {
		results, err := exec.Apply(context.Background(), graph, plan)
		assert.NoError(t, err) // the first task should return an error
		fmt.Println(results)
		assert.Equal(
			t,
			[]*exec.ApplyResult{
				{
					Path:      "module.test_module/dummy_task.fail",
					OldStatus: "will change: true",
					NewStatus: "will change: true",
					Success:   false,
				},
				{
					Path:      "module.test_module/dummy_task.dont_run",
					OldStatus: "will change: true",
					NewStatus: "failed due to dependencies: module.test_module/dummy_task.fail",
					Success:   false,
				},
			},
			results,
		)
	})
}
