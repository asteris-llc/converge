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
	"io/ioutil"
	"os"
	"os/exec"
)

// ShellTask is a task defined as two shell scripts
type ShellTask struct {
	TaskName    string
	CheckSource string `hcl:"check"`
	ApplySource string `hcl:"apply"`

	parent *Module
}

// Name returns name for metadata
func (st *ShellTask) Name() string {
	return st.TaskName
}

// Validate checks shell tasks validity
func (st *ShellTask) Validate() error {
	for _, script := range []string{st.CheckSource, st.ApplySource} {
		file, err := ioutil.TempFile("", "")
		if err != nil {
			return err
		}
		if err := ioutil.WriteFile(file.Name(), []byte(script), 0664); err != nil {
			return err
		}
		// validate the script using sh's built-in validation
		if err := exec.Command("sh", "-n", file.Name()).Run(); err != nil {
			switch script {
			case st.CheckSource:
				return ValidationError{Location: "check", Err: err}
			case st.ApplySource:
				return ValidationError{Location: "apply", Err: err}
			default:
				return err
			}
		}
		if err := os.Remove(file.Name()); err != nil {
			return err
		}
	}
	return nil
}

// Check satisfies the Monitor interface
func (st *ShellTask) Check() (string, error) {
	return "", nil
}

// Apply (plus Check) satisfies the Task interface
func (st *ShellTask) Apply() error {
	return nil
}

// Prepare this module for use
func (st *ShellTask) Prepare(parent *Module) error {
	st.parent = parent
	return nil
}
