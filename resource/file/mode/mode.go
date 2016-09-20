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

package mode

import (
	"fmt"
	"os"

	"github.com/asteris-llc/converge/resource"
)

// Mode monitors the mode of a file
type Mode struct {
	resource.Status

	Destination string
	Mode        os.FileMode
}

// Check whether the Destination has the right Mode
func (t *Mode) Check(resource.Renderer) (resource.TaskStatus, error) {
	diffs := make(map[string]resource.Diff)
	stat, err := os.Stat(t.Destination)
	if os.IsNotExist(err) {
		diffs[t.Destination] = &FileModeDiff{Expected: t.Mode}
		status := fmt.Sprintf("%q does not exist", t.Destination)
		return &resource.Status{
			Level:       resource.StatusFatal,
			Differences: diffs,
			Output:      []string{status},
		}, nil
	} else if err != nil {
		return nil, err
	}
	mode := stat.Mode().Perm()
	modeDiff := &FileModeDiff{Actual: mode, Expected: t.Mode}
	diffs[t.Destination] = modeDiff
	status := fmt.Sprintf("%q's mode is %q expected %q", t.Destination, mode, t.Mode)
	warningLevel := resource.StatusWontChange
	if modeDiff.Changes() {
		warningLevel = resource.StatusWillChange
	}

	t.Status = resource.Status{
		Level:       warningLevel,
		Differences: diffs,
		Output: []string{
			fmt.Sprintf("%q exist", t.Destination),
			status,
		},
	}
	return t, nil
}

// Apply the changes the Mode
func (t *Mode) Apply() (resource.TaskStatus, error) {
	err := os.Chmod(t.Destination, t.Mode.Perm())

	if err != nil {
		return &resource.Status{
			Level:  resource.StatusFatal,
			Output: []string{fmt.Sprintf("failed to set mode on %s: %s", t.Destination, err)},
		}, err
	}

	return t, nil
}

// Validate Mode
func (t *Mode) Validate() error {
	if t.Destination == "" {
		return fmt.Errorf("task requires a %q parameter", "destination")
	}
	if !(t.Mode.IsDir() || t.Mode.IsRegular()) {
		return fmt.Errorf("invalid %q parameter: %q", "mode", t.Mode)
	}
	return nil
}

// FileModeDiff shows a diff of the file modes
type FileModeDiff struct {
	Actual   os.FileMode
	Expected os.FileMode
}

// Original shows the original file mode
func (diff *FileModeDiff) Original() string {
	return fmt.Sprint(diff.Actual)
}

// Current shows the current file mode
func (diff *FileModeDiff) Current() string {
	return fmt.Sprint(diff.Expected)
}

// Changes returns true if the expected file mode differs from the current mode
func (diff *FileModeDiff) Changes() bool {
	return diff.Actual != diff.Expected
}
