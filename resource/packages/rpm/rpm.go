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

package rpm

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/asteris-llc/converge/resource"
	"github.com/pkg/errors"
)

// Package is an API for package state
type Package struct {
	Name  string
	State State
	*resource.Status
}

// State type for Package
type State string

const (
	// StatePresent indicates the package should be present
	StatePresent State = "present"

	// StateAbsent indicates the package should be absent
	StateAbsent State = "absent"
)

// Check if the package has to be 'installed', or 'absent'
func (p *Package) Check(resource.Renderer) (resource.TaskStatus, error) {
	status := resource.NewStatus()

	currentPkgStatusRaw, err := exec.Command("sh", "-c", fmt.Sprintf("yum info %q | grep Repo | awk '{print $3}'", p.Name)).Output()
	if err != nil {
		return status, errors.Wrapf(err, "checking package %s", p.Name)
	}
	currentPkgStatus := strings.TrimSpace(string(currentPkgStatusRaw))

	switch p.State {
	case StatePresent:
		if currentPkgStatus != "installed" {
			status.AddDifference(p.Name, currentPkgStatus, "installed", "")
			status.RaiseLevel(resource.StatusWillChange)
		} else {
			status.AddMessage("Package is installed")
		}
	case StateAbsent:
		if currentPkgStatus == "installed" {
			status.AddDifference(p.Name, currentPkgStatus, "uninstalled", "")
			status.RaiseLevel(resource.StatusWillChange)
		} else {
			status.AddMessage("Package is absent")
		}
	}

	p.Status = status
	return p, nil
}

// Apply desired package state
func (p *Package) Apply() (resource.TaskStatus, error) {
	status := resource.NewStatus()

	switch p.State {
	case StatePresent:
		currentPkgStatus, err := exec.Command("sh", "-c", "yum install -y "+p.Name).Output()

		if err != nil {
			return status, errors.Wrapf(err, "installing package %s, what happened: %s", p.Name, currentPkgStatus)
		}
		status.AddMessage(fmt.Sprintf("Package %q installed", p.Name))
		status.AddDifference(p.Name, "absent", "installed", "")

	case StateAbsent:
		currentPkgStatus, err := exec.Command("sh", "-c", "yum remove -y "+p.Name).Output()

		if err != nil {
			return status, errors.Wrapf(err, "installing package %s, what happened: %s", p.Name, currentPkgStatus)
		}
		status.AddMessage(fmt.Sprintf("Package %q removed", p.Name))
		status.AddDifference(p.Name, "installed", "absent", "")
	}

	p.Status = status
	return p, nil
}
