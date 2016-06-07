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
	TaskName       string
	RawCheckSource string   `hcl:"check"`
	RawApplySource string   `hcl:"apply"`
	Dependencies   []string `hcl:"depends"`

	renderer *Renderer
}

// Name returns name for metadata
func (st *ShellTask) Name() string {
	return st.TaskName
}

// Validate checks shell tasks validity
func (st *ShellTask) Validate() error {
	script, err := st.CheckSource()
	if err != nil {
		return ValidationError{Location: "check", Err: err}
	}
	if err := st.validateScriptSyntax(script); err != nil {
		return ValidationError{Location: "check", Err: err}
	}

	script, err = st.ApplySource()
	if err != nil {
		return ValidationError{Location: "apply", Err: err}
	}
	if err := st.validateScriptSyntax(script); err != nil {
		return ValidationError{Location: "apply", Err: err}
	}

	return nil
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

//AddDep Addes a dependency to this task
func (st *ShellTask) AddDep(dep string) {
	for _, same := range st.Dependencies {
		if same == dep {
			return
		}
	}
	st.Dependencies = append(st.Dependencies, dep)
}

//RemoveDep Removes a depencency from this task
func (st *ShellTask) RemoveDep(dep string) {
	for i, same := range st.Dependencies {
		if same == dep {
			st.Dependencies = append(st.Dependencies[:i], st.Dependencies[i+1:]...)
		}
	}
}

//Depends list dependencies for this task
func (st *ShellTask) Depends() []string {

	return st.Dependencies
}

// Check satisfies the Monitor interface
func (st *ShellTask) Check() (string, bool, error) {
	check, err := st.CheckSource()
	if err != nil {
		return "", false, err
	}

	out, code, err := st.exec(check)
	return out, code != 0, err
}

// Apply (plus Check) satisfies the Task interface
func (st *ShellTask) Apply() (string, bool, error) {
	apply, err := st.ApplySource()
	if err != nil {
		return "", false, err
	}

	out, code, err := st.exec(apply)
	return out, code == 0, err
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
	return err
}

// CheckSource renders RawCheckSource for execution
func (st *ShellTask) CheckSource() (string, error) {
	return st.renderer.Render("check", st.RawCheckSource)
}

// ApplySource renders RawApplySource for execution
func (st *ShellTask) ApplySource() (string, error) {
	return st.renderer.Render("apply", st.RawApplySource)
}
