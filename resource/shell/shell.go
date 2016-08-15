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

	"github.com/asteris-llc/converge/resource"
)

// Shell task
type Shell struct {
	Interpreter string
	CheckStmt   string
	ApplyStmt   string
}

// Check system using CheckStmt
func (s *Shell) Check() (resource.TaskStatus, error) {
	out, code, err := s.exec(s.CheckStmt)
	status := &resource.Status{
		WarningLevel: exitCodeToWarningLevel(code),
		Output:       messageMapToStringSlice(out),
		WillChange:   code != 0,
	}
	return status, err
}

// Apply ApplyStmt stanza to system
func (s *Shell) Apply() (err error) {
	out, code, err := s.exec(s.ApplyStmt)
	if code != 0 {
		return fmt.Errorf("exit code %d, stdout: %q, stderr: %q", code, out["stdout"], out["stderr"])
	}

	return err
}

func exitCodeToWarningLevel(exitCode uint32) int {
	switch exitCode {
	case 0:
		return resource.StatusOK
	case 1:
		return resource.StatusWarning
	default:
		return resource.StatusError
	}
}

func (s *Shell) exec(script string) (map[string]string, uint32, error) {
	messages := make(map[string]string)
	var code uint32
	command := exec.Command(s.Interpreter)
	stdin, err := command.StdinPipe()
	if err != nil {
		return messages, 0, err
	}

	// TODO: does this create a race condition?
	var stdoutBuffer bytes.Buffer
	var stderrBuffer bytes.Buffer
	command.Stdout = &stdoutBuffer
	command.Stderr = &stderrBuffer

	if err = command.Start(); err != nil {
		return messages, 0, err
	}

	if _, err = stdin.Write([]byte(script)); err != nil {
		return messages, 0, err
	}

	if err = stdin.Close(); err != nil {
		return messages, 0, err
	}

	err = command.Wait()
	if _, ok := err.(*exec.ExitError); !ok && err != nil {
		return messages, 0, err
	}

	switch result := command.ProcessState.Sys().(type) {
	case syscall.WaitStatus:
		code = uint32(result)
	default:
		panic(fmt.Sprintf("unknown type %+v", result))
	}

	if stdout := stdoutBuffer.String(); stdout != "" {
		messages["stdout"] = stdout
	}

	if stderr := stderrBuffer.String(); stderr != "" {
		messages["stderr"] = stderr
	}
	return messages, code, nil
}

func messageMapToStringSlice(m map[string]string) []string {
	var messages []string
	for k, v := range m {
		messages = append(messages, fmt.Sprintf("%s: %s", k, v))
	}
	return messages
}
