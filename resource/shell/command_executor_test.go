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

package shell_test

import (
	"testing"
	"time"

	"github.com/asteris-llc/converge/resource/shell"
	"github.com/stretchr/testify/assert"
)

func Test_Run_WhenScriptTimesOut_ReturnsTimeoutError(t *testing.T) {
	script := "sleep 100"
	timeout := 1 * time.Millisecond
	generator := &shell.CommandGenerator{
		Interpreter: "/bin/sh",
		Timeout:     &timeout,
	}
	_, err := generator.Run(script)
	assert.Error(t, err)
}

func Test_Run_WhenTimeotuSetScriptDoesNotTimeout_DoesNotReturnError(t *testing.T) {
	script := "true"
	timeout := 5 * time.Second
	generator := &shell.CommandGenerator{
		Interpreter: "/bin/sh",
		Timeout:     &timeout,
	}
	_, err := generator.Run(script)
	assert.NoError(t, err)
}

func Test_Run_WhenNoTimeout_RunsScript(t *testing.T) {
	script := "true"
	generator := &shell.CommandGenerator{Interpreter: "/bin/sh"}
	_, err := generator.Run(script)
	assert.NoError(t, err)
}

func Test_Run_ReturnsCommandResultsWithCorrectData(t *testing.T) {
	script := `echo -n "stdout";
echo -n "stderr" >&2;
exit 7`
	generator := &shell.CommandGenerator{Interpreter: "/bin/bash"}
	result, err := generator.Run(script)
	assert.NoError(t, err)
	assert.Equal(t, uint32(7), result.ExitStatus)
	assert.Equal(t, "stdout", result.Stdout)
	assert.Equal(t, "stderr", result.Stderr)
}

func Test_Run_RunsWithSpecifiedInterpreter(t *testing.T) {
	script := "puts RUBY_VERSION"
	generator := &shell.CommandGenerator{Interpreter: "/usr/bin/ruby"}
	result, err := generator.Run(script)
	assert.NoError(t, err)
	assert.NotEqual(t, "", result.Stdout)
}
