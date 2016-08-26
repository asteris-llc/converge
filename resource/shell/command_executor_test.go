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
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/asteris-llc/converge/resource/shell"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func Test_Run_WhenTimeoutSetScriptDoesNotTimeout_DoesNotReturnError(t *testing.T) {
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
	script := "echo -n 'foo'"
	generator := &shell.CommandGenerator{Interpreter: "/bin/bash"}
	result, err := generator.Run(script)
	assert.NoError(t, err)
	assert.False(t, strings.HasPrefix(result.Stdout, "-n"))
}

func Test_Run_RunsInDir(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "test-shell-task-check-in-dir")
	require.NoError(t, err)
	defer func() { require.NoError(t, os.RemoveAll(tmpdir)) }()

	script := "echo -n $(pwd)"
	generator := &shell.CommandGenerator{Interpreter: "/bin/bash", Dir: tmpdir}
	result, err := generator.Run(script)
	assert.NoError(t, err)

	pwd := result.Stdout
	pwd = strings.TrimPrefix(pwd, "/private") // on osx, /tmp is a symlink to private/tmp
	assert.Equal(t, tmpdir, pwd)
}

func Test_Run_RunsWithEnv(t *testing.T) {
	script := "echo -n \"Role: $ROLE, Version: $VERSION\""
	generator := &shell.CommandGenerator{
		Interpreter: "/bin/bash",
		Env:         []string{"ROLE=test", "VERSION=0.1"},
	}
	result, err := generator.Run(script)
	assert.NoError(t, err)
	assert.Equal(t, "Role: test, Version: 0.1", result.Stdout)
}
