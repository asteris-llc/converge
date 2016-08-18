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
	"fmt"

	"github.com/asteris-llc/converge/resource"
	"github.com/pkg/errors"
)

var outOfOrderMessage = "[WARNING] shell has no status code (maybe ran out-of-order)"

// Shell is a structure representing a task.
type Shell struct {
	CmdGenerator CommandExecutor
	CheckStmt    string
	ApplyStmt    string
	Status       *CommandResults
}

// Check passes through to shell.Shell.Check() and then sets the health status
func (s *Shell) Check() (resource.TaskStatus, error) {
	results, err := s.CmdGenerator.Run(s.CheckStmt)
	if err != nil {
		return nil, err
	}
	s.Status = s.Status.Cons("check", results)
	return s, nil
}

// Apply is a NOP for health checks
func (s *Shell) Apply() (err error) {
	results, err := s.CmdGenerator.Run(s.ApplyStmt)
	if err == nil {
		s.Status = s.Status.Cons("apply", results)
	}
	return err
}

// Healthy returns the health status of the node.  If a health check has not
// been run then Health() will call Check() before returning.  If a call to
// Check() fails Healthy() will return the error.
func (s *Shell) Healthy() (bool, error) {
	if s.Status == nil || s.Status.State == nil {
		return false, errors.New(outOfOrderMessage)
	}
	return s.Status.State.Success(), nil
}

// Warning returns true if the exit code of the last executed command (from
// check or apply) was 1.
func (s *Shell) Warning() bool {
	if s == nil || s.Status == nil {
		return false
	}
	return s.Status.ExitStatus == 1
}

// Error returns true if the exit code of the last executed command was greater
// than 1
func (s *Shell) Error() bool {
	if s == nil || s.Status == nil {
		return true
	}
	return s.Status.ExitStatus > 1
}

// Value provides a value for the shell, which is the stdout data from the last
// executed command.
func (s *Shell) Value() string {
	return s.Status.Stdout
}

// Diffs is required to implement resource.TaskStatus but there is no mechanism
// for defining diffs for shell operations, so returns a nil map.
func (s *Shell) Diffs() map[string]resource.Diff {
	return nil
}

// StatusCode returns the status code of the most recently executed command
func (s *Shell) StatusCode() int {
	if s.Status == nil {
		fmt.Println(outOfOrderMessage)
		return resource.StatusFatal
	}
	return int(s.Status.ExitStatus)
}

// Messages returns a summary of the first execution of check and/or apply.
// Subsequent runs are surpressed.
func (s *Shell) Messages() (messages []string) {
	if s.Status == nil {
		fmt.Println(outOfOrderMessage)
		return
	}
	messages = append(messages, s.Status.Reverse().UniqOp().SummarizeAll()...)
	return
}

// Changes returns true if changes are required as determined by the the most
// recent run of check.
func (s *Shell) Changes() bool {
	if s.Status == nil {
		fmt.Println(outOfOrderMessage)
		return false
	}
	return (s.Status.ExitStatus != 0)
}
