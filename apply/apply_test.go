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
	"github.com/asteris-llc/converge/helpers/faketask"
	"github.com/asteris-llc/converge/helpers/logging"
	"github.com/asteris-llc/converge/plan"
	"github.com/asteris-llc/converge/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyNoOp(t *testing.T) {
	defer logging.HideLogs(t)()

	g := graph.New()
	task := faketask.Swapper()
	g.Add("root", &plan.Result{Status: &resource.Status{WarningLevel: resource.StatusWillChange}, Task: task})

	require.NoError(t, g.Validate())

	// test that applying applies the vertex
	applied, err := apply.Apply(context.Background(), g)
	assert.NoError(t, err)

	result := getResult(t, applied, "root")
	assert.Equal(t, task.Status, result.Status.Messages()[0])
	assert.True(t, result.Ran)
}

func TestApplyNoRun(t *testing.T) {
	defer logging.HideLogs(t)()

	g := graph.New()
	task := faketask.NoOp()
	g.Add("root", &plan.Result{Status: &resource.Status{WarningLevel: resource.StatusWontChange}, Task: task})

	require.NoError(t, g.Validate())

	// test that running does not apply the vertex
	applied, err := apply.Apply(context.Background(), g)
	assert.NoError(t, err)

	result := getResult(t, applied, "root")
	assert.False(t, result.Ran)
}

func TestApplyErrorsBelow(t *testing.T) {
	defer logging.HideLogs(t)()

	g := graph.New()
	g.Add("root", &plan.Result{Status: &resource.Status{WarningLevel: resource.StatusWillChange}, Task: faketask.NoOp()})
	g.Add("root/err", &plan.Result{Status: &resource.Status{WarningLevel: resource.StatusWillChange}, Task: faketask.Error()})

	g.ConnectParent("root", "root/err")

	require.NoError(t, g.Validate())

	// applying will return an error if anything errors, and will set an error
	// in vertices that are higher up. This test should show an error in both
	// nodes.
	out, err := apply.Apply(context.Background(), g)
	assert.Equal(t, apply.ErrTreeContainsErrors, err)

	errNode, ok := out.Get("root/err").(*apply.Result)
	require.True(t, ok)
	assert.EqualError(t, errNode.Error(), "error applying root/err: error")

	rootNode, ok := out.Get("root").(*apply.Result)
	require.True(t, ok)
	assert.EqualError(t, rootNode.Error(), `error in dependency "root/err"`)
}

func TestApplyStillChange(t *testing.T) {
	defer logging.HideLogs(t)()

	g := graph.New()
	g.Add("root", &plan.Result{Status: &resource.Status{WarningLevel: resource.StatusWillChange}, Task: faketask.WillChange()})

	require.NoError(t, g.Validate())

	// applying should result in an error since the task will report that it still
	// needs to change
	_, err := apply.Apply(context.Background(), g)
	if assert.Error(t, err) {
		assert.Equal(t, err, apply.ErrTreeContainsErrors)
	}
}

func getResult(t *testing.T, src *graph.Graph, key string) *apply.Result {
	val := src.Get(key)
	result, ok := val.(*apply.Result)
	if !ok {
		t.Logf("needed a %T for %q, got a %T\n", result, key, val)
		t.FailNow()
	}

	return result
}
