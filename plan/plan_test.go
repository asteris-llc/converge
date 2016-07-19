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

package plan_test

import (
	"testing"

	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/helpers"
	"github.com/asteris-llc/converge/helpers/faketask"
	"github.com/asteris-llc/converge/plan"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlanNoOp(t *testing.T) {
	defer helpers.HideLogs(t)()

	g := graph.New()
	task := faketask.NoOp()
	g.Add("root", task)

	require.NoError(t, g.Validate())

	// test that running this results in an appropriate result
	planned, err := plan.Plan(g)
	assert.NoError(t, err)

	result := getResult(t, planned, "root")
	assert.Equal(t, task.Status, result.Status)
	assert.Equal(t, task.WillChange, result.WillChange)
	assert.Equal(t, task, result.Task)
}

func TestPlanErrorsBelow(t *testing.T) {
	defer helpers.HideLogs(t)()

	g := graph.New()
	g.Add("root", faketask.Error())
	g.Add("root/err", faketask.Error())

	g.Connect("root", "root/err")

	require.NoError(t, g.Validate())

	// planning will return an error if any of the leaves error, but it won't even
	// touch vertices that are higher up. This test should show an error in the
	// leafmost node, and not the root.
	_, err := plan.Plan(g)
	if assert.Error(t, err) {
		assert.EqualError(
			t,
			err,
			"1 error(s) occurred:\n\n* error checking root/err: error",
		)
	}
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
