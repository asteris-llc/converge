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
	"fmt"
	"io/ioutil"
	"os"
)

// Template is a task defined by content and a destination
type Template struct {
	TemplateName   string
	RawContent     string   `hcl:"content"`
	RawDestination string   `hcl:"destination"`
	Dependencies   []string `hcl:"depends"`
	renderer       *Renderer
}

// Name returns the name of this template
func (t *Template) Name() string {
	return "template." + t.TemplateName
}

// Validate validates the template config
func (t *Template) Validate() error {
	_, err := t.renderer.Render("content", t.RawContent)
	if err != nil {
		return err
	}

	_, err = t.renderer.Render("destination", t.RawDestination)
	return err
}

//SetDepends overwrites the Dependencies of this resource
func (t *Template) SetDepends(deps []string) {
	t.Dependencies = deps
}

//Depends list dependencies for this task
func (t *Template) Depends() []string {

	return t.Dependencies
}

// Check satisfies the Monitor interface
func (t *Template) Check() (string, bool, error) {
	dest := t.Destination()

	stat, err := os.Stat(dest)
	if os.IsNotExist(err) {
		return "", true, nil
	} else if err != nil {
		return "", false, err
	} else if stat.IsDir() {
		return "", true, fmt.Errorf("cannot template %q, is a directory", dest)
	}

	actual, err := ioutil.ReadFile(dest)
	if err != nil {
		return "", false, err
	}

	return string(actual), t.Content() != string(actual), nil
}

// Apply (plus Check) satisfies the Task interface
func (t *Template) Apply() (string, bool, error) {
	dest := t.Destination()
	content := t.Content()
	err := ioutil.WriteFile(dest, []byte(content), 0600)
	if err != nil {
		return "", false, err
	}

	actual, err := ioutil.ReadFile(dest)
	if err != nil {
		return "", false, err
	}

	return content, content == string(actual), err
}

// Prepare this module for use
func (t *Template) Prepare(parent *Module) (err error) {
	t.renderer, err = NewRenderer(parent)

	return err
}

// Destination renders the destination
func (t *Template) Destination() string {
	// we're ignoring the error here because it's already been validated
	dest, _ := t.renderer.Render("destination", t.RawDestination)
	return dest
}

// Content renders the content
func (t *Template) Content() string {
	// we're ignoring the error here because it's already been validated
	content, _ := t.renderer.Render("content", t.RawContent)
	return content
}
