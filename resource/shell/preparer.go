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
	"io"
	"io/ioutil"
	"os/exec"
	"time"

	"github.com/pkg/errors"

	log "github.com/Sirupsen/logrus"
	"github.com/asteris-llc/converge/helpers/transform"
	"github.com/asteris-llc/converge/load/registry"
	"github.com/asteris-llc/converge/resource"
	"golang.org/x/net/context"
)

var (
	defaultInterpreter = "/bin/sh"
	defaultCheckFlags  = []string{"-n"}
	defaultExecFlags   = []string{}
)

// Preparer for shell tasks
//
// Task allows you to run arbitrary shell commands on your system, first
// checking if the command should be run.
type Preparer struct {
	// the shell interpreter that will be used for your scripts. `/bin/sh` is
	// used by default.
	Interpreter string `hcl:"interpreter"`

	// flags to pass to the `interpreter` binary to check validity. For
	// `/bin/sh` this is `-n`
	CheckFlags []string `hcl:"check_flags"`

	// flags to pass to the interpreter at execution time
	ExecFlags []string `hcl:"exec_flags"`

	// the script to run to check if a resource needs to be changed. It should
	// exit with exit code 0 if the resource does not need to be changed, and
	// 1 (or above) otherwise.
	Check string `hcl:"check"`

	// the script to run to apply the resource. Normal shell exit code
	// expectations apply (that is, exit code 0 for success, 1 or above for
	// failure.)
	Apply string `hcl:"apply"`

	// the amount of time the command will wait before halting forcefully. The
	// format is Go's duration string. A duration string is a possibly signed
	// sequence of decimal numbers, each with optional fraction and a unit
	// suffix, such as "300ms", "-1.5h" or "2h45m". Valid time units are "ns",
	// "us" (or "µs"), "ms", "s", "m", "h".
	Timeout time.Duration `hcl:"timeout"`

	// the working directory this command should be run in
	Dir string `hcl:"dir"`

	// any environment variables that should be passed to the command
	Env map[string]string `hcl:"env"`
}

// Prepare a new shell task
func (p *Preparer) Prepare(ctx context.Context, render resource.Renderer) (resource.Task, error) {
	env := transform.StringsMapToStringSlice(
		p.Env,
		func(k, v string) string {
			return fmt.Sprintf("%s=%s", k, v)
		},
	)

	generator := &CommandGenerator{
		Interpreter: p.Interpreter,
		Flags:       p.ExecFlags,
		Dir:         p.Dir,
		Env:         env,
		Timeout:     &p.Timeout,
	}

	shell := &Shell{
		CmdGenerator: generator,
		CheckStmt:    p.Check,
		ApplyStmt:    p.Apply,
		Dir:          p.Dir,
		Env:          env,
	}

	return shell, checkSyntax(p.Interpreter, p.CheckFlags, p.Check)
}

func checkSyntax(interpreter string, flags []string, script string) error {
	if interpreter == "" {
		interpreter = defaultInterpreter
		if len(flags) > 0 {
			return errors.New("custom syntax check_flags given without an interpreter")
		}
		flags = defaultCheckFlags
	} else {
		if len(flags) == 0 {
			// TODO: add ID in here somehow
			log.Debug("no check_flags specified for interpeter, skipping syntax validation")
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

func init() {
	registry.Register("task", (*Preparer)(nil), (*Shell)(nil))
	registry.Register("healthcheck.task", (*Preparer)(nil), (*Shell)(nil))
}
