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

// Content renders content to disk
type Content struct {
	Content     string
	Destination string
	*resource.Status
}

// Check if the content needs to be rendered
func (t *Content) Check(resource.Renderer) (resource.TaskStatus, error) {
	diffs := make(map[string]resource.Diff)
	contentDiff := resource.TextDiff{Values: [2]string{"", t.Content}}
	stat, err := os.Stat(t.Destination)
	if os.IsNotExist(err) {
		contentDiff.Values[0] = "<file-missing>"
		diffs[t.Destination] = contentDiff
		t.Status = &resource.Status{
			WarningLevel: resource.StatusWillChange,
			WillChange:   true,
			Differences:  diffs,
			Status:       t.Destination + ": File is missing",
		}
		return t, nil
	} else if err != nil {
		t.Status = &resource.Status{
			WarningLevel: resource.StatusFatal,
			Status:       "Cannot read `" + t.Destination + "`",
		}
		return t, err
	} else if stat.IsDir() {
		t.Status = &resource.Status{
			WarningLevel: resource.StatusFatal,
			WillChange:   true,
			Status:       t.Destination + " is a directory",
		}
		return t, fmt.Errorf("cannot update contents of %q, it is a directory", t.Destination)
	}

	actual, err := ioutil.ReadFile(t.Destination)
	if err != nil {
		t.Status = &resource.Status{}
		return t, err
	}

	statusMessage := "OK"

	if string(actual) != t.Content {
		statusMessage = "contents differ"
		diffs[t.Destination] = resource.TextDiff{Values: [2]string{string(actual), t.Content}}
	}

	t.Status = &resource.Status{
		Status:      statusMessage,
		Differences: diffs,
		WillChange:  resource.AnyChanges(diffs),
	}
	return t, nil
}

// Apply writes the content to disk
func (t *Content) Apply() (resource.TaskStatus, error) {
	var perm os.FileMode
	var preChange string
	diffs := make(map[string]resource.Diff)

	stat, err := os.Stat(t.Destination)
	if os.IsNotExist(err) {
		diffs["mode"] = resource.TextDiff{Values: [2]string{"not set", "0600"}}
		perm = 0600
	} else if err != nil {
		return &resource.Status{
			WarningLevel: resource.StatusFatal,
			WillChange:   true,
			Output:       []string{err.Error()},
		}, err
	} else {
		perm = stat.Mode()
	}

	if rawData, readErr := ioutil.ReadFile(t.Destination); readErr != nil {
		preChange = "<file-missing>"
	} else {
		preChange = string(rawData)
	}

	diffs[t.Destination] = resource.TextDiff{Values: [2]string{preChange, t.Content}}

	if err = ioutil.WriteFile(t.Destination, []byte(t.Content), perm); err != nil {
		t.Status = &resource.Status{
			Status:       fmt.Sprintf("%s", err),
			WarningLevel: resource.StatusFatal,
			Differences:  diffs,
		}
		return t, err
	}

	t.Status = &resource.Status{Differences: diffs}
	return t, nil
}
