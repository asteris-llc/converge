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

var applyResult1 = &exec.ApplyResult{
	Path:      "moduleC/submodule1",
	OldStatus: "old",
	NewStatus: "new",
	Success:   true,
}

var applyResult2 = &exec.ApplyResult{
	Path:      "moduleD/submodule1",
	OldStatus: "old",
	NewStatus: "old",
	Success:   false,
}

func TestApplyResultString(t *testing.T) {
	t.Parallel()

	expected1 := "+moduleC/submodule1:\n\tStatus: \"old\" => \"new\"\n\tSuccess: true"
	expected2 := "-moduleD/submodule1:\n\tStatus: \"old\" => \"old\"\n\tSuccess: false"
	assert.Equal(t, expected1, applyResult1.String())
	assert.Equal(t, expected2, applyResult2.String())
}

func TestApplyResultPretty(t *testing.T) {
	t.Parallel()

	expected1 := "\x1b[32m+\x1b[0m\x1b[32mmoduleC/submodule1\x1b[0m:\n\tStatus: \"old\" => \"new\"\n\tSuccess: \x1b[32mtrue\x1b[0m"
	expected2 := "\x1b[31m-\x1b[0m\x1b[31mmoduleD/submodule1\x1b[0m:\n\tStatus: \"old\" => \"old\"\n\tSuccess: \x1b[31mfalse\x1b[0m"
	assert.Equal(t, expected1, applyResult1.Pretty())
	assert.Equal(t, expected2, applyResult2.Pretty())
}
