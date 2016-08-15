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

	"github.com/asteris-llc/converge/helpers/fakerenderer"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/shell"
	"github.com/stretchr/testify/assert"
)

func Test_Preparer_ImplementsResourceInterface(t *testing.T) {
	t.Parallel()
	assert.Implements(t, (*resource.Resource)(nil), new(shell.Preparer))
}

func Test_Prepare_ReturnsError_WhenScriptFailsSyntaxCheck(t *testing.T) {
	t.Parallel()
	p := shPreparer("if [[ -x")
	_, err := p.Prepare(fakerenderer.New())
	assert.Error(t, err)
}

func Test_Prepare_ReturnsNilError_WhenScriptPassesSyntaxCheck(t *testing.T) {
	t.Parallel()
	p := shPreparer("true")
	_, err := p.Prepare(fakerenderer.New())
	assert.NoError(t, err)
}

func Test_Prepare_ReturnsCheckWithShellTask_WhenSyntaxOK(t *testing.T) {
	t.Parallel()
	checkFlag := []string{"-n"}
	expectedInterpreter := "/bin/sh"
	expectedStatement := "true"
	p := &shell.Preparer{
		Interpreter: expectedInterpreter,
		CheckFlags:  checkFlag,
		Check:       expectedStatement,
	}
	returnedCheck, _ := p.Prepare(fakerenderer.New())
	actualCheck, ok := returnedCheck.(*shell.Shell)
	assert.True(t, ok)
	fmt.Println(actualCheck)
	_, ok = actualCheck.ExecutionShell.(resource.Task)
	assert.True(t, ok)
}

func Test_Prepare_ReturnsError_WhenSyntaxError(t *testing.T) {
	t.Parallel()
	checkFlag := []string{"-n"}
	expectedInterpreter := "/bin/sh"
	expectedStatement := "if [[ -x"
	p := &shell.Preparer{
		Interpreter: expectedInterpreter,
		CheckFlags:  checkFlag,
		Check:       expectedStatement,
	}
	_, err := p.Prepare(fakerenderer.New())
	assert.Error(t, err)
}

func shPreparer(script string) *shell.Preparer {
	syntaxFlag := []string{"-n"}
	return &shell.Preparer{
		Interpreter: "/bin/sh",
		CheckFlags:  syntaxFlag,
		Check:       script,
	}
}
