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
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/shell"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShellInterface(t *testing.T) {
	t.Parallel()

	assert.Implements(t, (*resource.Task)(nil), new(shell.Shell))
}

func TestShellTaskCheckNeedsChange(t *testing.T) {
	t.Parallel()

	s := shell.Shell{
		Interpreter: "sh",
		CheckStmt:   "echo test && exit 1",
	}

	current, change, err := s.Check()
	assert.Equal(t, "test\n", current)
	assert.True(t, change)
	assert.Nil(t, err)
}

func TestShellCheckNoChange(t *testing.T) {
	t.Parallel()

	s := shell.Shell{
		Interpreter: "sh",
		CheckStmt:   "echo test",
	}

	current, change, err := s.Check()
	assert.Equal(t, "test\n", current)
	assert.False(t, change)
	assert.Nil(t, err)
}

func TestShellApplySuccess(t *testing.T) {
	t.Parallel()

	s := shell.Shell{
		Interpreter: "sh",
		ApplyStmt:   "echo test",
	}

	assert.NoError(t, s.Apply())
}

func TestShellTaskApplyError(t *testing.T) {
	t.Parallel()

	s := shell.Shell{
		Interpreter: "sh",
		ApplyStmt:   "echo bad && exit 1",
	}

	err := s.Apply()
	if assert.Error(t, err) {
		assert.EqualError(
			t,
			err,
			`exit code 256, output: "bad\n"`,
		)
	}
}

func TestShellTaskCheckInDir(t *testing.T) {
	t.Parallel()

	tmpdir, err := ioutil.TempDir("", "test-shell-task-check-in-dir")
	require.NoError(t, err)
	defer func() { require.NoError(t, os.RemoveAll(tmpdir)) }()

	s := shell.Shell{
		Interpreter: "sh",
		CheckStmt:   "pwd",
		Dir:         tmpdir,
	}

	current, change, err := s.Check()

	// workaround: on osx /tmp is symlinked to /private/tmp.
	actual := strings.TrimPrefix(current, "/private")

	assert.Equal(t, fmt.Sprintf("%s\n", tmpdir), actual)
	assert.False(t, change)
	assert.NoError(t, err)
}

func TestShellTaskCheckWithEnv(t *testing.T) {
	t.Parallel()

	s := shell.Shell{
		Interpreter: "sh",
		CheckStmt:   "echo \"ROLE: $ROLE\"",
		Env:         []string{"ROLE=test"},
	}

	current, change, err := s.Check()

	assert.Equal(t, "ROLE: test\n", current)
	assert.False(t, change)
	assert.NoError(t, err)
}

func TestShellTaskApplyInDir(t *testing.T) {
	t.Parallel()

	tmpdir, err := ioutil.TempDir("", "test-shell-task-apply-in-dir")
	require.NoError(t, err)
	defer func() { require.NoError(t, os.RemoveAll(tmpdir)) }()

	s := shell.Shell{
		Interpreter: "sh",
		ApplyStmt:   "touch dir-test",
		Dir:         tmpdir,
	}

	err = s.Apply()
	assert.NoError(t, err)

	// verify file exists in working directory
	_, err = os.Stat(path.Join(tmpdir, "dir-test"))
	assert.NoError(t, err)
}

func TestShellTaskApplyWithEnv(t *testing.T) {
	t.Parallel()

	tmpdir, err := ioutil.TempDir("", "test-shell-task-apply-with-env")
	require.NoError(t, err)
	defer func() { require.NoError(t, os.RemoveAll(tmpdir)) }()

	s := shell.Shell{
		Interpreter: "sh",
		Dir:         tmpdir,
		ApplyStmt:   "echo \"ROLE: $ROLE\" > test-with-env",
		Env:         []string{"ROLE=test"},
	}

	err = s.Apply()
	assert.NoError(t, err)

	// verify file exists in working directory
	bytes, err := ioutil.ReadFile(path.Join(tmpdir, "test-with-env"))
	assert.NoError(t, err)

	assert.Equal(t, "ROLE: test\n", string(bytes))
}
