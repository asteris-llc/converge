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

	"github.com/asteris-llc/converge/resource"
	"github.com/pkg/errors"

	"io"
	"io/ioutil"
	"log"
	"os/exec"
	"syscall"
	"time"
)

// NB: Known Bug with timed script execution:

// Currently when a script executes beyond it's alloted time a timeout will
// occur and nil is returned by timeoutExec.  The goroutine running the script
// will continue to drain stdout and stderr from the process sockets and they
// will be GCed when the script finally finishes.  This means that there is no
// mechanism for getting the output of a script when it has timed out.  A
// proposed solution to this would be to implement a ReadUntilWouldBlock-type
// function that would allow us to read into a buffer from a ReadCloser until
// the read operation would block, then return the contents of the buffer (along
// with some marker if we recevied an error or EOF).  Then exec function would
// then take in pointers to buffers for stdout and stderr and populate them
// directly, so that if the script execution timed out we would still have a
// reference to those buffers.
var (
	ScriptTimedOutError = errors.New("execution timed out")
)

var outOfOrderMessage = "[WARNING] shell has no status code (maybe ran out-of-order)"

type CommandIOContext struct {
	Command *exec.Cmd
	Stdin   io.WriteCloser
	Stdout  io.ReadCloser
	Stderr  io.ReadCloser
}

type CommandGenerator struct {
	Interpreter string
	Flags       []string
	Timeout     *time.Duration
}

// Shell is a structure representing a task.
type Shell struct {
	CmdGenerator *CommandGenerator
	CheckStmt    string
	ApplyStmt    string
	Description  string
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

// GetDescription returns the description of the health check
func (s *Shell) GetDescription() string {
	if s.Description == "" {
		return "Unnamed Health Check"
	}
	return s.Description
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

func (cmd *CommandGenerator) Run(script string) (*CommandResults, error) {
	ctx, err := cmd.start()
	if err != nil {
		return nil, err
	}
	return ctx.Run(script, cmd.Timeout)
}

// Run wraps exec and timeoutExec, executing the script with or without a
// timeout depending whether or not timeout is nil.
func (c *CommandIOContext) Run(script string, timeout *time.Duration) (results *CommandResults, err error) {
	if timeout == nil {
		results, err = c.exec(script)
	} else {
		results, err = c.timeoutExec(script, *timeout)
	}
	return
}

// timeoutExec will run the given script with a timelimit specified by
// timeout. If the script does not return with that time duration a
// ScriptTimeoutError is returned.
func (c *CommandIOContext) timeoutExec(script string, timeout time.Duration) (*CommandResults, error) {
	timeoutChannel := make(chan []interface{}, 1)
	go func() {
		cmdResults, err := c.exec(script)
		timeoutChannel <- []interface{}{cmdResults, err}
	}()
	select {
	case result := <-timeoutChannel:
		return result[0].(*CommandResults), result[1].(error)
	case <-time.After(timeout):
		return nil, ScriptTimedOutError
	}
}

func (c *CommandIOContext) exec(script string) (results *CommandResults, err error) {
	results = &CommandResults{
		Stdin:    script,
		Timedout: false,
	}

	if err = c.Command.Start(); err != nil {
		return
	}
	if _, err = c.Stdin.Write([]byte(script)); err != nil {
		return
	}
	if err = c.Stdin.Close(); err != nil {
		return
	}

	if data, readErr := ioutil.ReadAll(c.Stdout); readErr == nil {
		results.Stdout = string(data)
	} else {
		log.Printf("[WARNING] cannot read stdout from script")
	}

	if data, readErr := ioutil.ReadAll(c.Stderr); readErr == nil {
		results.Stderr = string(data)
	} else {
		log.Printf("[WARNING] cannot read stderr from script")
	}

	if waitErr := c.Command.Wait(); waitErr == nil {
		results.ExitStatus = 0
	} else {
		exitErr, ok := waitErr.(*exec.ExitError)
		if !ok {
			err = errors.Wrap(waitErr, "failed to wait on process")
			return
		}
		status, ok := exitErr.Sys().(syscall.WaitStatus)
		if !ok {
			err = errors.New("unexpected error getting exit status")
		}
		results.ExitStatus = uint32(status.ExitStatus())
	}
	results.State = c.Command.ProcessState
	return
}

//func start(interpreter string, flags []string) (*CommandIOContext, error) {
func (cmd *CommandGenerator) start() (*CommandIOContext, error) {
	command := newCommand(cmd.Interpreter, cmd.Flags)
	stdin, stdout, stderr, err := cmdGetPipes(command)
	return &CommandIOContext{
		Command: command,
		Stdin:   stdin,
		Stdout:  stdout,
		Stderr:  stderr,
	}, err
}

func newCommand(interpreter string, flags []string) *exec.Cmd {
	if interpreter == "" {
		if len(flags) > 0 {
			log.Println("[INFO] passing flasg to default interpreter (/bin/sh)")
			return exec.Command(defaultInterpreter, flags...)
		}
		return exec.Command(defaultInterpreter, defaultExecFlags...)
	}
	return exec.Command(interpreter, flags...)
}

func (s *Shell) Warning() bool {
	if s == nil || s.Status == nil {
		return false
	}
	return s.Status.ExitStatus == 1
}

func (s *Shell) Error() bool {
	if s == nil || s.Status == nil {
		return true
	}
	return s.Status.ExitStatus > 1
}

func (s *Shell) Value() string {
	var value bytes.Buffer
	value.WriteString(s.Description + "\n")
	for _, message := range s.Messages() {
		value.WriteString(message)
	}
	return value.String()
}

func (s *Shell) Diffs() map[string]resource.Diff {
	return nil
}

func (s *Shell) StatusCode() int {
	if s.Status == nil {
		fmt.Println(outOfOrderMessage)
		return resource.StatusFatal
	}
	return int(s.Status.ExitStatus)
}

func (s *Shell) Messages() (messages []string) {
	if s.Status == nil {
		fmt.Println(outOfOrderMessage)
		return
	}
	uniqStatuses := s.Status.Reverse().UniqOp()

	exitCodes := uniqStatuses.ExitStrings()
	messages = append(messages, fmt.Sprintf("Exit Code(s): %v", exitCodes))

	messages = append(messages, uniqStatuses.GetMessages()...)
	return
}

func (s *Shell) Changes() bool {
	if s.Status == nil {
		fmt.Println(outOfOrderMessage)
		return false
	}
	return (s.Status.ExitStatus != 0)
}
