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

import "text/template"

// Template is a task defined by content and a destination
type Template struct {
	TemplateName string
	Content      string `hcl:"content"`
	Destination  string `hcl:"destination"`
}

// Name returns the name of this template
func (t *Template) Name() string {
	return t.TemplateName
}

// Validate validates the template config
func (t *Template) Validate() error {
	_, err := template.New("validating").Parse(t.Content)
	if err != nil {
		return err
	}
	return err
}

// Check satisfies the Monitor interface
func (t *Template) Check() (string, error) {
	return "", nil
}

// Apply (plus Check) satisfies the Task interface
func (t *Template) Apply() error {
	return nil
}
