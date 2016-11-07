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

	"github.com/asteris-llc/converge/resource/package"
)

// YumManager provides a concrete implementation of PackageManager for yum
// packages.
type YumManager struct {
	Sys pkg.SysCaller
}

// InstalledVersion gets the installed version of package, if available
func (y *YumManager) InstalledVersion(p string) (pkg.PackageVersion, bool) {
	result, err := y.Sys.Run(fmt.Sprintf("rpm -q %s", p))
	exitCode, _ := pkg.GetExitCode(err)
	if exitCode != 0 {
		return "", false
	}
	return (pkg.PackageVersion)(result), true
}

// InstallPackage installs a package, returning an error if something went wrong
func (y *YumManager) InstallPackage(pkg string) (string, error) {
	if _, isInstalled := y.InstalledVersion(pkg); isInstalled {
		return "already installed", nil
	}
	res, err := y.Sys.Run(fmt.Sprintf("yum install -y %s", pkg))
	return string(res), err
}

// RemovePackage removes a package, returning an error if something went wrong
func (y *YumManager) RemovePackage(pkg string) (string, error) {
	res, err := y.Sys.Run(fmt.Sprintf("yum remove -y %s", pkg))
	return string(res), err
}
