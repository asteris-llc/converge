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
	"strings"

	"github.com/asteris-llc/converge/resource/package"
)

// Outputs from dpkg-query
const (
	// PkgHold means package is installed and held to a version
	PkgHold = "hold ok installed"

	// PkgInstalled means package is installed normally
	PkgInstalled = "install ok installed"

	// PkgRemoved means package has been removed, but not purged (config files still present)
	PkgRemoved = "deinstall ok config-files"

	// PkgUninstalled means package has been uninstalled
	PkgUninstalled = "unknown ok not-installed"
)

// Manager provides a concrete implementation of PackageManager for debian
// packages.
type Manager struct {
	Sys pkg.SysCaller
}

// InstalledVersion gets the installed version of package, if available
func (a *Manager) InstalledVersion(p string) (pkg.PackageVersion, bool) {
	var version string
	var installed bool

	result, err := a.Sys.Run(fmt.Sprintf("dpkg-query -W -f'${Package},${Status},${Version}\n' %s", p))
	exitCode, _ := pkg.GetExitCode(err)
	if exitCode != 0 {
		return "", false
	}
	// Deal with situations where dpkg exits 0 with packages that have been uninstalled
	for _, line := range strings.Split(strings.TrimSpace(string(result)), "\n") {
		l := strings.Split(line, ",")
		if len(l) == 3 {
			name, status, ver := l[0], l[1], l[2]

			if strings.Contains(status, PkgRemoved) || strings.Contains(status, PkgUninstalled) {
				return "", false
			}

			if strings.Contains(status, PkgInstalled) || strings.Contains(status, PkgHold) {
				version = fmt.Sprintf("%s-%s", name, ver)
				installed = true
			}
		}
	}
	return (pkg.PackageVersion)(version), installed
}

// InstallPackage installs a package, returning an error if something went wrong
func (a *Manager) InstallPackage(p string) (string, error) {
	if _, isInstalled := a.InstalledVersion(p); isInstalled {
		return "already installed", nil
	}
	res, err := a.Sys.Run(fmt.Sprintf("apt-get install -y %s", p))
	return string(res), err
}

// RemovePackage removes a package, returning an error if something went wrong
func (a *Manager) RemovePackage(p string) (string, error) {
	switch _, isInstalled := a.InstalledVersion(p); isInstalled {
	case true:
		res, err := a.Sys.Run(fmt.Sprintf("apt-get purge -y %s", p))
		return string(res), err
	default:
		return "package is not installed ", nil
	}

}
