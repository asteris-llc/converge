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
	"testing"

	"github.com/asteris-llc/converge/exec"
	"github.com/asteris-llc/converge/load"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlan(t *testing.T) {
	t.Parallel()

	graph, err := load.Load("../samples/basic.hcl")
	require.NoError(t, err)

	results, err := exec.Plan(graph)
	assert.NoError(t, err)
	assert.Equal(
		t,
		[]*exec.PlanResult{{
			Path:          "basic.hcl/render",
			CurrentStatus: "cat: test.txt: No such file or directory\n",
			WillChange:    true,
		}},
		results,
	)
}
