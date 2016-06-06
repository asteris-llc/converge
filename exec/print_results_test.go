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
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestPlanResultsPrint(t *testing.T) {
	t.Parallel()

	rs := exec.Results{planResult1, planResult2}
	viper.Set("nocolor", true)
	assert.Equal(t, "moduleA/submodule1:\n\tCurrently: status\n\tWill Change: true\n\nmoduleB/submodule1:\n\tCurrently: status\n\tWill Change: false", rs.Print())
}

func TestApplyResultsPrint(t *testing.T) {
	t.Parallel()

	rs := exec.Results{applyResult1, applyResult2}
	viper.Set("nocolor", true)
	assert.Equal(t, "+ moduleC/submodule1:\n\tStatus: \"old\" => \"new\"\n\tSuccess: true\n\n- moduleD/submodule1:\n\tStatus: \"old\" => \"old\"\n\tSuccess: false", rs.Print())
}
