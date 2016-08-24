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

package directory

import (
	"fmt"
	"os"

	"github.com/asteris-llc/converge/resource"
)

// Mode monitors the file Mode of a file
type Directory struct {
	Destination string
	Force       bool
}

// Check whether the Directory at the destination should exist
// 1. If directory doesn't exist create it
// 2. If directory doesn't exist don't create it
// 3. If a file is present at the destination and force is true, create it.
func (t *Directory) Check() (resource.TaskStatus, error) {
	diffs := make(map[string]resource.Diff)
	stat, err := os.Stat(t.Destination)
	status := fmt.Sprintf("directory %q exists", t.Destination)
	if os.IsNotExist(err) {
		newStatus := fmt.Sprintf("directory %q does not exist", t.Destination)
		diffs[t.Destination] = &FileDirectoryDiff{
			Expected: status,
			Actual:   newStatus,
		}
		status := newStatus
		return &resource.Status{
			Status:       status,
			WarningLevel: resource.StatusWillChange,
			WillChange:   true,
			Differences:  diffs,
			Output: []string{
				status,
				fmt.Sprintf("directory %q will be created", t.Destination),
			},
		}, nil
	} else if err != nil {
		return nil, err
	}
	directoryOrFile := "file"
	if stat.IsDir() {
		directoryOrFile = "directory"
	}
	// Assuming Force is false
	status = fmt.Sprintf("%s %q exists", directoryOrFile, t.Destination)
	directoryDiff := &FileDirectoryDiff{Actual: status, Expected: status}
	diffs[t.Destination] = directoryDiff
	warningLevel := resource.StatusWontChange
	forceMsg := fmt.Sprintf("files will not be overriden unless the %q param is true", t.Destination, "force")
	// If force change it anyway
	if t.Force && !stat.IsDir() {
		directoryDiff.Expected = fmt.Sprintf("directory %q exists", t.Destination)
		warningLevel = resource.StatusWillChange
		forceMsg = fmt.Sprintf("%q %q will be overriden because the %q param is true", directoryOrFile, t.Destination, "force")
	}

	return &resource.Status{
		Status:       status,
		WarningLevel: warningLevel,
		WillChange:   directoryDiff.Changes() && t.Force && !stat.IsDir(),
		Differences:  diffs,
		Output: []string{
			status,
			forceMsg,
		},
	}, nil

}

// Apply creates the directory based on the above checks
func (t *Directory) Apply() error {
	stat, err := os.Stat(t.Destination)
	if os.IsNotExist(err) {
		return os.MkdirAll(t.Destination, 700)
	} else if err != nil {
		return err
	}
	if t.Force && !stat.IsDir() {
		err = os.Remove(t.Destination)
		if err != nil {
			return err
		}
		return os.MkdirAll(t.Destination, 700)
	}
	return nil
}

// Validate Directory
func (t *Directory) Validate() error {
	if t.Destination == "" {
		return fmt.Errorf("task requires a %q parameter", "destination")
	}
	return nil
}

// String diff between mesages
type FileDirectoryDiff struct {
	Actual   string
	Expected string
}

func (diff *FileDirectoryDiff) Original() string {
	return diff.Actual
}

func (diff *FileDirectoryDiff) Current() string {
	return diff.Expected
}

func (diff *FileDirectoryDiff) Changes() bool {
	return diff.Actual != diff.Expected
}
