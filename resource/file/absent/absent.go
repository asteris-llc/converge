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

package absent

import (
	"fmt"
	"os"

	"github.com/asteris-llc/converge/resource"
)

// Absent monitors the file Absent of a file
type Absent struct {
	Destination string
}

// Check whether the file at the Destination exist
// 1. If file or directory exist, remove it
// 2. Else do nothing
func (t *Absent) Check() (resource.TaskStatus, error) {
	diffs := make(map[string]resource.Diff)
	_, err := os.Stat(t.Destination)
	expected := fmt.Sprintf("%q does not exist", t.Destination)

	// If something exsist remove it
	if os.IsNotExist(err) {
		diffs[t.Destination] = &FileAbsentDiff{Expected: expected, Actual: expected}
		return &resource.Status{
			Status:       expected,
			WarningLevel: resource.StatusWontChange,
			WillChange:   false,
			Differences:  diffs,
			Output:       []string{expected},
		}, nil
	} else if err != nil { // Actual error
		return nil, err
	} else {
		// If it exist set change to true
		status := fmt.Sprintf("%q exist", t.Destination)
		diffs[t.Destination] = &FileAbsentDiff{Expected: expected, Actual: status}
		return &resource.Status{
			Status:       status,
			WarningLevel: resource.StatusWillChange,
			WillChange:   true,
			Differences:  diffs,
			Output:       []string{status},
		}, nil
	}
}

// Remove files at the destination
func (t *Absent) Apply() error {
	return os.Remove(t.Destination)
}

// Validate Mode
func (t *Absent) Validate() error {
	if t.Destination == "" {
		return fmt.Errorf("task requires a %q parameter", "destination")
	}
	return nil
}

// FileAbsentDiff is a basic string diff
type FileAbsentDiff struct {
	Actual   string
	Expected string
}

// Original shows the original file mode
func (diff *FileAbsentDiff) Original() string {
	return diff.Actual
}

// Current shows the current file mode
func (diff *FileAbsentDiff) Current() string {
	return diff.Expected
}

// Changes returns true if the expected file mode differs from the current mode
func (diff *FileAbsentDiff) Changes() bool {
	return diff.Actual != diff.Expected
}
