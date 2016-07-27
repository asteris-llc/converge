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

package content

import (
	"fmt"
	"io/ioutil"
	"os"
)

// Content renders a content to disk
type Content struct {
	Content     string
	Destination string
}

// Check if the content needs to be rendered
func (t *Content) Check() (status string, willChange bool, err error) {
	stat, err := os.Stat(t.Destination)
	if os.IsNotExist(err) {
		return "", true, nil
	} else if err != nil {
		return "", false, err
	} else if stat.IsDir() {
		return "", true, fmt.Errorf("cannot content %q, is a directory", t.Destination)
	}

	actual, err := ioutil.ReadFile(t.Destination)
	if err != nil {
		return "", false, err
	}

	return string(actual), t.Content != string(actual), nil
}

// Apply writes the content to disk
func (t *Content) Apply() error {
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
