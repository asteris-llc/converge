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
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/pkg/errors"
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
	ErrTimedOut = errors.New("execution timed out")
)

// A CommandExecutor supports running a script and returning the results wrapped
// in a *CommandResults structure.
type CommandExecutor interface {
	Run(string) (*CommandResults, error)
}

// CommandGenerator provides a container to wrap generating a system command
type CommandGenerator struct {
	Interpreter string
	Flags       []string
	Dir         string
	Env         []string
	Timeout     *time.Duration
}

// Run will generate a new command and run it with optional timeout parameters
func (cmd *CommandGenerator) Run(script string) (*CommandResults, error) {
	ctx, err := cmd.start()
	if err != nil {
		return nil, err
	}
	return ctx.Run(script, cmd.Timeout)
}

func (cmd *CommandGenerator) start() (*commandIOContext, error) {
	command := newCommand(cmd)
	stdin, stdout, stderr, err := cmdGetPipes(command)
	return &commandIOContext{
		Command: command,
		Stdin:   stdin,
		Stdout:  stdout,
		Stderr:  stderr,
	}, err
}

// commandIOContext provides the context for a command that includes it's stdin,
// stdout, and stderr pipes along with the underlying command.
type commandIOContext struct {
	Command *exec.Cmd
	Stdin   io.WriteCloser
	Stdout  io.ReadCloser
	Stderr  io.ReadCloser
}

// Run wraps exec and timeoutExec, executing the script with or without a
// timeout depending whether or not timeout is nil.
func (c *commandIOContext) Run(script string, timeout *time.Duration) (results *CommandResults, err error) {
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
func (c *commandIOContext) timeoutExec(script string, timeout time.Duration) (*CommandResults, error) {
	timeoutChannel := make(chan []interface{}, 1)
	go func() {
		cmdResults, err := c.exec(script)
		timeoutChannel <- []interface{}{cmdResults, err}
	}()
	select {
	case result := <-timeoutChannel:
		var errResult error
		if result[1] != nil {
			errResult = result[1].(error)
		}
		return result[0].(*CommandResults), errResult
	case <-time.After(timeout):
		return nil, ErrTimedOut
	}
}

func (c *commandIOContext) exec(script string) (results *CommandResults, err error) {
	results = &CommandResults{
		Stdin: script,
	}

	// if working dir does not exist, we want the check to return a non-zero
	// result. otherwise, Command.Start will return an error and short-circuit
	// plan/apply
	if c.Command.Dir != "" {
		_, err := os.Stat(c.Command.Dir)
		if os.IsNotExist(err) {
			results.ExitStatus = 1
			results.Stdout = err.Error()
			return results, nil
		}
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

func newCommand(cmd *CommandGenerator) *exec.Cmd {
	var command *exec.Cmd
	if cmd.Interpreter == "" {
		if len(cmd.Flags) > 0 {
			log.Println("[INFO] passing flags to default interpreter (/bin/sh)")
			command = exec.Command(defaultInterpreter, cmd.Flags...)
		} else {
			command = exec.Command(defaultInterpreter, defaultExecFlags...)
		}
	} else {
		command = exec.Command(cmd.Interpreter, cmd.Flags...)
	}

	command.Dir = cmd.Dir
	if len(cmd.Env) > 0 {
		env := os.Environ()
		for _, envvar := range cmd.Env {
			env = append(env, envvar)
		}
		command.Env = env
	}

	return command
}
