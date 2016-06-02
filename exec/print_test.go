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
	"github.com/stretchr/testify/assert"
)

// some values to test with

var result1 = &exec.PlanResult{
	Path:          "test1.hcl/result1",
	CurrentStatus: "status1",
	WillChange:    true,
}

var result2 = &exec.PlanResult{
	Path:          "test2.hcl/result2",
	CurrentStatus: "status2",
	WillChange:    true,
}

var result3 = &exec.PlanResult{
	Path:          "test2.hcl/result2",
	CurrentStatus: "status2",
	WillChange:    false,
}

var results = exec.Results{result1, result2, result3}

// this implies the success of Sort
func TestResultsLess(t *testing.T) {
	assert.True(t, results.Less(0, 1))
	assert.True(t, results.Less(1, 2))
}
