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

package resource

import (
	"bytes"
	"fmt"
	"os/exec"
	"syscall"
)

// ShellTask is a task defined as two shell scripts
type ShellTask struct {
	DependencyTracker `hcl:",squash"`

	Name           string
	RawCheckSource string `hcl:"check"`
	RawApplySource string `hcl:"apply"`

	renderer    *Renderer
	checkSource string
	applySource string
}

// String returns name for metadata
func (st *ShellTask) String() string {
	return "task." + st.Name
}

func (st *ShellTask) validateScriptSyntax(script string) error {
	command := exec.Command("sh", "-n")

	in, err := command.StdinPipe()
	if err != nil {
		return err
	}

	if err := command.Start(); err != nil {
		return err
	}

	if _, err := in.Write([]byte(script)); err != nil {
		return err
	}

	if err := in.Close(); err != nil {
		return err
	}

	if err := command.Wait(); err != nil {
		return err
	}

	return nil
}

// Check satisfies the Monitor interface
func (st *ShellTask) Check() (string, bool, error) {
	out, code, err := st.exec(st.checkSource)
	return out, code != 0, err
}

// Apply (plus Check) satisfies the Task interface
func (st *ShellTask) Apply() error {
	out, code, err := st.exec(st.applySource)
	if code != 0 {
		return fmt.Errorf("exit code %d, output: %q", code, out)
	}
	return err
}

func (st *ShellTask) exec(script string) (out string, code uint32, err error) {
	command := exec.Command("sh")
	stdin, err := command.StdinPipe()
	if err != nil {
		return "", 0, err
	}

	// TODO: does this create a race condition?
	var sink bytes.Buffer
	command.Stdout = &sink
	command.Stderr = &sink

	if err := command.Start(); err != nil {
		return "", 0, err
	}

	if _, err := stdin.Write([]byte(script)); err != nil {
		return "", 0, err
	}

	if err := stdin.Close(); err != nil {
		return "", 0, err
	}

	err = command.Wait()
	if _, ok := err.(*exec.ExitError); !ok && err != nil {
		return "", 0, err
	}

	switch result := command.ProcessState.Sys().(type) {
	case syscall.WaitStatus:
		code = uint32(result)
	default:
		panic(fmt.Sprintf("unknown type %+v", result))
	}

	return sink.String(), code, nil
}

// Prepare this module for use
func (st *ShellTask) Prepare(parent *Module) (err error) {
	st.renderer, err = NewRenderer(parent)
	if err != nil {
		return err
	}

	// validate and cache check
	st.checkSource, err = st.renderer.Render(st.String()+".check", st.RawCheckSource)
	if err != nil {
		return ValidationError{Location: st.String() + ".check", Err: err}
	}

	if err := st.validateScriptSyntax(st.checkSource); err != nil {
		return ValidationError{Location: st.String() + ".check", Err: err}
	}

	// validate and cache apply
	st.applySource, err = st.renderer.Render(st.String()+".apply", st.RawApplySource)
	if err != nil {
		return ValidationError{Location: st.String() + ".apply", Err: err}
	}

	if err := st.validateScriptSyntax(st.applySource); err != nil {
		return ValidationError{Location: st.String() + ".apply", Err: err}
	}

	err = st.DependencyTracker.ComputeDependencies(
		st.String()+".dependencies",
		st.renderer,
		st.RawCheckSource,
		st.RawApplySource,
	)
	if err != nil {
		return ValidationError{Location: st.String() + ".dependencies", Err: err}
	}

	return nil
}

// SetName modifies the name of this ShellTask
func (st *ShellTask) SetName(name string) {
	st.Name = name
}
