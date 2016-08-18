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
	"io"
	"io/ioutil"
	"log"
	"os/exec"
	"time"

	"github.com/pkg/errors"

	"github.com/asteris-llc/converge/resource"
)

var (
	defaultInterpreter = "/bin/sh"
	defaultCheckFlags  = []string{"-n"}
	defaultExecFlags   = []string{}
)

// Preparer for shell tasks
type Preparer struct {
	Interpreter string   `hcl:"interpreter"`
	CheckFlags  []string `hcl:"check_flags"`
	ExecFlags   []string `hcl:"run_flags"`
	Check       string   `hcl:"check"`
	Apply       string   `hcl:"apply"`
	Timeout     string   `hcl:"timeout"`
}

// Prepare a new shell task
func (p *Preparer) Prepare(render resource.Renderer) (resource.Task, error) {

	check, err := render.Render("check", p.Check)
	if err != nil {
		return nil, err
	}

	apply, err := render.Render("apply", p.Apply)
	if err != nil {
		return nil, err
	}

	interpreter, err := render.Render("interpreter", p.Interpreter)
	if err != nil {
		return nil, err
	}

	timeout, err := render.Render("timeout", p.Timeout)
	if err != nil {
		return nil, err
	}

	generator := &CommandGenerator{
		Interpreter: interpreter,
		Flags:       p.ExecFlags,
	}

	if duration, err := time.ParseDuration(timeout); err == nil {
		generator.Timeout = &duration
	}

	shell := &Shell{
		CmdGenerator: generator,
		CheckStmt:    check,
		ApplyStmt:    apply,
	}

	return shell, checkSyntax(interpreter, p.CheckFlags, check)
}

func checkSyntax(interpreter string, flags []string, script string) error {
	if interpreter == "" {
		interpreter = defaultInterpreter
		if len(flags) > 0 {
			log.Println("[ERROR] check_flags specified without an interpreter")
			return errors.New("custom syntax check_flags given without an interpreter")
		}
		flags = defaultCheckFlags
	} else {
		if len(flags) == 0 {
			log.Println("[INFO] no check_flags specified for interpreter, skipping syntax validation")
			return nil
		}
	}
	command := exec.Command(interpreter, flags...)
	cmdStdin, cmdStdout, cmdStderr, err := cmdGetPipes(command)
	if err != nil {
		return errors.Wrap(err, "unable to communicate with subprocess")
	}
	if err := command.Start(); err != nil {
		return errors.Wrap(err, "unable to start subprocess")
	}
	if _, err := cmdStdin.Write([]byte(script)); err != nil {
		return errors.Wrap(err, "unable to write to interpreter")
	}

	if err := cmdStdin.Close(); err != nil {
		return errors.Wrap(err, "failed to close stdin")
	}

	var buffer bytes.Buffer
	if data, err := ioutil.ReadAll(cmdStdout); err == nil {
		if len(data) > 0 {
			buffer.WriteString("Command Stdout:\n")
			buffer.Write(data)
		}
	}
	if data, err := ioutil.ReadAll(cmdStderr); err == nil {
		if len(data) > 0 {
			buffer.WriteString("Command Stderr:\n")
			buffer.Write(data)
		}
	}

	if err := command.Wait(); err != nil {
		return errors.Wrap(err, "syntax error: "+buffer.String())
	}

	return nil
}

func cmdGetPipes(command *exec.Cmd) (io.WriteCloser, io.ReadCloser, io.ReadCloser, error) {
	var err error
	cmdStdin, err := command.StdinPipe()
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "failed to get stdin pipe")
	}
	cmdStderr, err := command.StderrPipe()
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "failed to get stderr pipe")
	}
	cmdStdout, err := command.StdoutPipe()
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "failed to get stdout pipe")
	}
	return cmdStdin, cmdStdout, cmdStderr, nil
}
