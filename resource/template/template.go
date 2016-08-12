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

package template

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/asteris-llc/converge/resource"
)

// Template renders a template to disk
type Template struct {
	Content     string
	Destination string
}

// Check if the template needs to be rendered
func (t *Template) Check() (resource.TaskStatus, error) {
	stat, err := os.Stat(t.Destination)
	if os.IsNotExist(err) {
		return resource.NewStatus("", true, nil)
	} else if err != nil {
		return resource.NewStatus("", false, err)
	} else if stat.IsDir() {
		return resource.NewStatus("", true, fmt.Errorf("cannot template %q, is a directory", t.Destination))
	}

	actual, err := ioutil.ReadFile(t.Destination)
	if err != nil {
		return resource.NewStatus("", false, err)
	}

	return resource.NewStatus(string(actual), t.Content != string(actual), nil)
}

// Apply writes the content to disk
func (t *Template) Apply() error {
	var perm os.FileMode

	stat, err := os.Stat(t.Destination)
	if os.IsNotExist(err) {
		perm = 0600
	} else if err != nil {
		return err
	} else {
		perm = stat.Mode()
	}

	return ioutil.WriteFile(t.Destination, []byte(t.Content), perm)
}
