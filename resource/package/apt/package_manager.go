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

package apt

import (
	"fmt"

	"github.com/asteris-llc/converge/resource/package"
)

// Manager provides a concrete implementation of PackageManager for debian
// packages.
type Manager struct {
	Sys pkg.SysCaller
}

// InstalledVersion gets the installed version of package, if available
func (y *Manager) InstalledVersion(p string) (pkg.PackageVersion, bool) {
	result, err := y.Sys.Run(fmt.Sprintf("dpkg -s %s", p))
	exitCode, _ := pkg.GetExitCode(err)
	if exitCode != 0 {
		return "", false
	}
	return (pkg.PackageVersion)(result), true
}

// InstallPackage installs a package, returning an error if something went wrong
func (y *Manager) InstallPackage(p string) (string, error) {
	if _, isInstalled := y.InstalledVersion(p); isInstalled {
		return "already installed", nil
	}
	res, err := y.Sys.Run(fmt.Sprintf("apt-get install -y %s", p))
	return string(res), err
}

// RemovePackage removes a package, returning an error if something went wrong
func (y *Manager) RemovePackage(pkg string) (string, error) {
	res, err := y.Sys.Run(fmt.Sprintf("apt-get remove -y %s", pkg))
	return string(res), err
}
