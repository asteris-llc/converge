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

	log "github.com/Sirupsen/logrus"
	"github.com/asteris-llc/converge/resource"
)

// Package is an API for package state
type Package struct {
	resource.TaskStatus

	Name  string
	State string
}

// Check if the package has to be 'installed', or 'absent'
func (p *Package) Check(resource.Renderer) (resource.TaskStatus, error) {
	status := resource.NewStatus()

	currentPkgStatusRaw, err := exec.Command("bash", "-c", fmt.Sprintf("yum info %q | grep Repo | awk '{print $3}'", p.Name)).Output()
	if err != nil {
		log.WithField("module", "rpm").Warn(err)
	}
	currentPkgStatus := strings.TrimSpace(string(currentPkgStatusRaw))

	log.WithField("module", "rpm").WithField("exec", "bash").Info(string(currentPkgStatus))
	log.Info(p)
	if p.State == "installed" {
		if string(currentPkgStatus) != "installed" {
			status.AddDifference(p.Name, string(currentPkgStatus), "installed", "")
		} else {
			status.AddMessage("Package is installed")
		}
	} else if p.State == "absent" {
		if string(currentPkgStatus) == "installed" {
			status.AddDifference(p.Name, string(currentPkgStatus), "uninstalled", "")
		}
	}

	p.TaskStatus = status
	return p, nil
}

// Apply desired package state
func (p *Package) Apply() (resource.TaskStatus, error) {
	status := resource.NewStatus()

	if p.State == "installed" {
		currentPkgStatus, err := exec.Command("bash", "-c", "yum install -y "+p.Name).Output()
		if err != nil {
			log.WithField("module", "rpm").Warn(err)
			log.WithField("module", "rpm").Warn(string(currentPkgStatus))
		}
		status.AddMessage(fmt.Sprintf("Package %q installed", p.Name))
		status.AddDifference(p.Name, "absent", "installed", "")

		log.WithField("module", "rpm").Info(fmt.Sprintf("Installed %q: \n %q", p.Name, string(currentPkgStatus)))
	} else if p.State == "absent" {
		currentPkgStatus, err := exec.Command("bash", "-c", "yum remove -y "+p.Name).Output()
		if err != nil {
			log.WithField("module", "rpm").Warn(err)
			log.WithField("module", "rpm").Warn(string(currentPkgStatus))
		}

		status.AddMessage(fmt.Sprintf("Package %q removed", p.Name))
		status.AddDifference(p.Name, "installed", "absent", "")

		log.WithField("module", "rpm").Info(fmt.Sprintf("Removed %q: \n %q", p.Name, string(currentPkgStatus)))
	}

	status.RaiseLevel(resource.StatusNoChange)
	p.TaskStatus = status

	return p, nil
}
