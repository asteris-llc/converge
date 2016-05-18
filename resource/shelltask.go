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

// ShellTask is a task defined as two shell scripts
type ShellTask struct {
	Name        string
	CheckSource string `hcl:"check"`
	ApplySource string `hcl:"apply"`
}

// Check satisfies the Monitor interface
func (st *ShellTask) Check() (string, error) {
	return "", nil
}

// Apply (plus Check) satisfies the Task interface
func (st *ShellTask) Apply() error {
	return nil
}
