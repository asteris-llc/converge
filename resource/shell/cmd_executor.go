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

package shell

import (
	"bytes"
	"fmt"
	"os/exec"
	"syscall"

	"github.com/pkg/errors"
)

// CommandExecuter provides an implementation of ScriptExecuter that passes
// through to exec.Command
type CommandExecuter struct{}

// CheckSyntax is a passthrough to cmd.Execute
func (c *CommandExecuter) CheckSyntax(interpreter string, flags []string, command string) error {
	cmd := exec.Command(interpreter, flags...)
	in, err := cmd.StdinPipe()
	if err != nil {
		return errors.Wrap(err, "Unable to create pipe to command")
	}
	if err = cmd.Start(); err != nil {
		return errors.Wrap(err, "failed to start interpreter at: "+interpreter)
	}
	if _, err := in.Write([]byte(command)); err != nil {
		return errors.Wrap(err, "unable to write to interpreter")
	}
	if err := in.Close(); err != nil {
		return errors.Wrap(err, "failed to close pipe")
	}
	if err := cmd.Wait(); err != nil {
		return errors.Wrap(err, "syntax error")
	}
	return nil
}

// ExecuteCommand is a passthrough to cmd.Execute
func (c *CommandExecuter) ExecuteCommand(interpreter string, flags []string, script string) (string, int, error) {
	var code int

	command := exec.Command(interpreter, flags...)
	stdin, err := command.StdinPipe()

	if err != nil {
		return "", 0, err
	}

	// TODO: does this create a race condition?
	var sink bytes.Buffer
	command.Stdout = &sink
	command.Stderr = &sink

	if err = command.Start(); err != nil {
		return "", 0, err
	}

	if _, err = stdin.Write([]byte(script)); err != nil {
		return "", 0, err
	}

	if err = stdin.Close(); err != nil {
		return "", 0, err
	}

	err = command.Wait()
	if _, ok := err.(*exec.ExitError); !ok && err != nil {
		return "", 0, err
	}

	switch result := command.ProcessState.Sys().(type) {
	case syscall.WaitStatus:
		code = result.ExitStatus()
	default:
		panic(fmt.Sprintf("unknown type %+v", result))
	}

	return sink.String(), code, nil
}
