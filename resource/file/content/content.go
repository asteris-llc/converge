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

	"github.com/asteris-llc/converge/resource"
)

// Content renders a content to disk
type Content struct {
	Content     string
	Destination string
}

// Check if the content needs to be rendered
func (t *Content) Check() (resource.TaskStatus, error) {
	diffs := make(map[string]resource.Diff)
	contentDiff := resource.TextDiff{Values: [2]string{"", t.Content}}
	stat, err := os.Stat(t.Destination)
	if os.IsNotExist(err) {
		contentDiff.Values[0] = "<file-missing>"
		diffs[t.Destination] = contentDiff
		return &resource.Status{
			WarningLevel: resource.StatusWillChange,
			WillChange:   true,
			Differences:  diffs,
		}, nil
	} else if err != nil {
		return &resource.Status{
			WarningLevel: resource.StatusFatal,
		}, err
	} else if stat.IsDir() {
		return &resource.Status{
			WarningLevel: resource.StatusFatal,
			WillChange:   true,
		}, fmt.Errorf("cannot update contents of %q, it is a directory", t.Destination)
	}

	actual, err := ioutil.ReadFile(t.Destination)
	if err != nil {
		return &resource.Status{}, err
	}

	diffs[t.Destination] = resource.TextDiff{Values: [2]string{string(actual), t.Content}}
	return &resource.Status{
		Status:      t.Destination,
		Differences: diffs,
		WillChange:  resource.AnyChanges(diffs),
	}, nil
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

func getWarningLevel(original *string, content string) int {
	if original == nil || *original == "" {
		return resource.StatusWillChange
	}
	if *original != content {
		return resource.StatusWillChange
	}
	return resource.StatusNoChange
}
