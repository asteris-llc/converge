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
	"path"

	"os"

	"github.com/asteris-llc/converge/resource"
	"github.com/pkg/errors"
)

// Directory makes sure a directory is present on disk
type Directory struct {
	Destination string
	CreateAll   bool
}

// Check if the directory exists
func (d *Directory) Check(resource.Renderer) (resource.TaskStatus, error) {
	status := resource.NewStatus()

	dest := d.Destination
	for dest != "/" {
		stat, err := os.Stat(dest)

		switch {
		case err != nil && !os.IsNotExist(err):
			return status, errors.Wrapf(err, "could not stat %q")

		case os.IsNotExist(err):
			status.WillChange = true

			// if we aren't told to create everything, we should fail early
			// since we can't create the parent directories
			if !d.CreateAll && dest != d.Destination {
				status.RaiseLevel(resource.StatusFatal)
				status.AddMessage(fmt.Sprintf("%q does not exist and will not be created (enable create_all to do this)", dest))
				status.Differences = nil
				return status, nil
			}

			status.RaiseLevel(resource.StatusWillChange)
			status.AddDifference(dest, "<absent>", "<present>", "<absent>")

		case !stat.IsDir():
			status.RaiseLevel(resource.StatusFatal)
			status.AddMessage(fmt.Sprintf("%q already exists and is not a directory", dest))
			return status, nil

		default:
			status.RaiseLevel(resource.StatusNoChange)
			if !status.HasChanges() {
				status.AddMessage(fmt.Sprintf("%q already exists", dest))
			}
			return status, nil
		}

		// pop a bit off the end
		dest = path.Dir(dest)
	}

	return status, nil
}

// Apply creates the directory
func (d *Directory) Apply(resource.Renderer) (resource.TaskStatus, error) {
	var err error

	if d.CreateAll {
		err = os.MkdirAll(d.Destination, 0700)
	} else {
		err = os.Mkdir(d.Destination, 0700)
	}

	if err != nil {
		return nil, err
	}

	status := resource.NewStatus()
	status.RaiseLevel(resource.StatusWillChange)
	status.AddMessage(fmt.Sprintf("%q exists", d.Destination))

	return status, err
}
