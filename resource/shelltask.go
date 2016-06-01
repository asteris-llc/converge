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

import "os/exec"

// ShellTask is a task defined as two shell scripts
type ShellTask struct {
	TaskName       string
	RawCheckSource string `hcl:"check"`
	RawApplySource string `hcl:"apply"`

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

// Check satisfies the Monitor interface
func (st *ShellTask) Check() (string, bool, error) {
	return "", false, nil
}

// Apply (plus Check) satisfies the Task interface
func (st *ShellTask) Apply() error {
	return nil
}

// Prepare this module for use
func (st *ShellTask) Prepare(parent *Module) (err error) {
	st.renderer, err = NewRenderer(parent)
	return err
}

// CheckSource renders RawCheckSource for execution
func (st *ShellTask) CheckSource() (string, error) {
	return st.renderer.Render(st.RawCheckSource)
}

// ApplySource renders RawApplySource for execution
func (st *ShellTask) ApplySource() (string, error) {
	return st.renderer.Render(st.RawApplySource)
}
