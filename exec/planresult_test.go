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

var result1 = &exec.PlanResult{
	Path:          "moduleA/submodule1",
	CurrentStatus: "status",
	WillChange:    true,
}

var result2 = &exec.PlanResult{
	Path:          "moduleB/submodule1",
	CurrentStatus: "status",
	WillChange:    false,
}

func TestResultString(t *testing.T) {
	t.Parallel()

	expected1 := "moduleA/submodule1:\n\tCurrently: status\n\tWill Change: true"
	expected2 := "moduleB/submodule1:\n\tCurrently: status\n\tWill Change: false"
	assert.Equal(t, expected1, result1.String())
	assert.Equal(t, expected2, result2.String())
}

func TestResultPretty(t *testing.T) {
	t.Parallel()

	expected1 := "\x1b[1;30mmoduleA/submodule1\x1b[0m:\n\tCurrently: \x1b[33mstatus\x1b[0m\n\tWill Change: \x1b[33mtrue\x1b[0m"
	expected2 := "\x1b[1;30mmoduleB/submodule1\x1b[0m:\n\tCurrently: \x1b[34mstatus\x1b[0m\n\tWill Change: \x1b[34mfalse\x1b[0m"
	assert.Equal(t, expected1, result1.Pretty())
	assert.Equal(t, expected2, result2.Pretty())
}

func TestResultsString(t *testing.T) {
	t.Parallel()

	rs := exec.Results{result1, result2}
	expected := "moduleA/submodule1:\n\tCurrently: status\n\tWill Change: true\n"
	expected += "moduleB/submodule1:\n\tCurrently: status\n\tWill Change: false"
	assert.Equal(t, expected, rs.String())
}

func TestResultsPretty(t *testing.T) {
	t.Parallel()

	rs := exec.Results{result1, result2}
	expected := "\x1b[1;30mmoduleA/submodule1\x1b[0m:\n\tCurrently: \x1b[33mstatus\x1b[0m\n\tWill Change: \x1b[33mtrue\x1b[0m\n\x1b[1;30mmoduleB/submodule1\x1b[0m:\n\tCurrently: \x1b[34mstatus\x1b[0m\n\tWill Change: \x1b[34mfalse\x1b[0m"
	assert.Equal(t, expected, rs.Pretty())
}

func TestResultsSorting(t *testing.T) {
	t.Parallel()

	// the results should be printed in the opposite order in which they appear in
	// this slice, they should be sorted by path
	rs := exec.Results{result2, result1}
	expected := "moduleA/submodule1:\n\tCurrently: status\n\tWill Change: true\n"
	expected += "moduleB/submodule1:\n\tCurrently: status\n\tWill Change: false"
	assert.Equal(t, expected, rs.Pretty())
}
