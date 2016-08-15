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
	"fmt"
	"testing"

	"github.com/asteris-llc/converge/resource/shell"
	"github.com/stretchr/testify/assert"
)

var (
	executor      *shell.CommandExecuter
	shInterpreter = "/bin/sh"
	shCheckFlags  = []string{"-n"}
	shExecFlags   = []string{}
)

func Test_CheckSyntax_ReturnsErrorWhenSyntaxError(t *testing.T) {
	badScript := "if [ -x"
	assert.Error(t, checkSh(badScript))
}

func Test_CheckSyntax_ReturnsNoError_WhenSyntaxOkandReturnsExitFailure(t *testing.T) {
	script := "false"
	assert.NoError(t, checkSh(script))
}

func Test_CheckSyntax_CallsSpecifiedInterpreter(t *testing.T) {
	interpreter := "/usr/bin/ruby"
	goodScript := "if true then puts foo else puts bar end"
	badScript := "if true then puts foo else puts bar"
	flags := []string{"-c"}
	assert.NoError(t, executor.CheckSyntax(interpreter, flags, goodScript))
	assert.Error(t, executor.CheckSyntax(interpreter, flags, badScript))
}

func Test_ExecuteCommand_ReturnsStdoutFromCommand(t *testing.T) {
	expected := "foo\n"
	script := "echo foo"
	actual, _, _ := executeSh(script)
	assert.Equal(t, expected, actual)
}

func Test_ExecuteCommand_ReturnsStatusCodeFromCommand(t *testing.T) {
	returnCode := int(112)
	script := fmt.Sprintf("exit %d", returnCode)
	_, actualCode, _ := executeSh(script)
	assert.Equal(t, returnCode, actualCode)
}

func Test_ExecuteCommand_ReturnsError_WhenExecutionError(t *testing.T) {
	_, _, err := executor.ExecuteCommand("non-existant interpreter", shExecFlags, "")
	assert.Error(t, err)
}

func Test_ExecuteCommand_UsesSpecifiedInterpreter(t *testing.T) {
	badExitStatus := int(27)
	interpreter := "/usr/bin/ruby"
	goodScript := "exit EXIT_SUCCESS"
	badScript := fmt.Sprintf("exit %d", badExitStatus)
	_, returnCode, err := executor.ExecuteCommand(interpreter, shExecFlags, goodScript)
	assert.NoError(t, err)
	assert.Equal(t, 1, returnCode)

	_, returnCode, err = executor.ExecuteCommand(interpreter, shExecFlags, badScript)
	assert.NoError(t, err)
	assert.Equal(t, badExitStatus, returnCode)
}

func checkSh(cmd string) error {
	return executor.CheckSyntax(shInterpreter, shCheckFlags, cmd)
}

func executeSh(cmd string) (string, int, error) {
	return executor.ExecuteCommand(shInterpreter, shExecFlags, cmd)
}
