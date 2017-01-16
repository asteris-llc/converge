// Copyright © 2017 Asteris, LLC
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

package unarchive

import (
	"fmt"
	"os"

	"github.com/asteris-llc/converge/resource"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

// Unarchive manages file unarchive
type Unarchive struct {

	// the source
	Source string `export:"source"`

	// the destination
	Destination string `export:"destination"`
}

// Check if changes are needed for unarchive
func (u *Unarchive) Check(context.Context, resource.Renderer) (resource.TaskStatus, error) {
	status := resource.NewStatus()

	err := u.Diff(status)
	if err != nil {
		return status, err
	}

	return status, nil
}

// Apply changes for unarchive
func (u *Unarchive) Apply(context.Context) (resource.TaskStatus, error) {
	status := resource.NewStatus()
	return status, nil
}

// Diff evaluates the differences for unarchive
func (u *Unarchive) Diff(status *resource.Status) error {
	_, err := os.Stat(u.Source)
	if os.IsNotExist(err) {
		status.RaiseLevel(resource.StatusCantChange)
		return errors.Wrap(err, "cannot unarchive")
	}

	stat, err := os.Stat(u.Destination)
	if err == nil {
		if !stat.IsDir() {
			status.RaiseLevel(resource.StatusCantChange)
			return fmt.Errorf("invalid destination %q, must be directory", u.Destination)
		}
	} else if os.IsNotExist(err) {
		status.RaiseLevel(resource.StatusCantChange)
		return fmt.Errorf("destination %q does not exist", u.Destination)
	}

	status.AddDifference("unarchive", u.Source, u.Destination, "")
	status.RaiseLevelForDiffs()

	return nil
}
