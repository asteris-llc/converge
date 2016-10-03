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
	"os/exec"

	log "github.com/Sirupsen/logrus"
	"github.com/asteris-llc/converge/resource"
)

type Package struct {
	Name  string
	State string
}

func (p *Package) Check() (resource.TaskStatus, error) {
	status := &resource.Status{
		Status:     "whoop",
		WillChange: false,
	}

	//current_pkg_status, err := exec.Command("/bin/bash", "-c", "yum info "+p.Name+" | grep Repo | awk '{print $3}'").Output()
	current_pkg_status, err := exec.Command("bash", "-c", "apt-cache policy "+p.Name+" | grep Installed | awk '{print $2}'").Output()
	if err != nil {
		log.WithField("module", "rpm").Warn(err)
	}

	log.WithField("module", "rpm").WithField("exec", "sh").Info(string(current_pkg_status))
	log.Info(p)
	if p.State == "installed" {
		if string(current_pkg_status) == "(none)" {
			status.AddDifference("package", p.Name, "will be installed", "")
			status.WillChange = true
		}
		status.Status = string(current_pkg_status)
	} else if p.State == "absent" {
		if string(current_pkg_status) != "(none)" {
			status.AddDifference("package", p.Name, "will be uninstalled", "")
			status.WillChange = true
		}
	}

	return status, nil
}

func (t *Package) Apply() error {
	return nil
}
