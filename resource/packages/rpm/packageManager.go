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
	"syscall"

	"github.com/pkg/errors"
)

// PackageVersion is a type alias for a string, since various packages may use
// different naming conventions
type PackageVersion string

// PackageManager describes an interface for managing packages and helps make
// `Check` and `Apply` testable.
type PackageManager interface {
	// If the package is installed, returns the version and true, otherwise
	// returns an empty string and false.
	InstalledVersion(string) (PackageVersion, bool)

	// Installs a package, returning an error if something went wrong
	InstallPackage(string) error

	// Removes a package, returning an error if something went wrong
	RemovePackage(string) error
}

// SysCaller allows us to mock exec.Command
type SysCaller interface {
	Run(string) ([]byte, error)
}

// ExecCaller is a dummy struct to handle wrapping exec.Command in the SysCaller
// interface.
type ExecCaller struct{}

// Run executs `cmd` as a /bin/sh script and returns the output and error
func (e ExecCaller) Run(cmd string) ([]byte, error) {
	return exec.Command("sh", "-c", cmd).Output()
}

// YumManager provides a concrete implementation of PackageManager for yum
// packages.
type YumManager struct {
	Sys SysCaller
}

// InstalledVersion gets the installed version of package, if available
func (y *YumManager) InstalledVersion(pkg string) (PackageVersion, bool) {
	result, err := y.Sys.Run(fmt.Sprintf("rpm -q %s", pkg))
	exitCode, _ := getExitCode(err)
	if exitCode != 0 {
		return "", false
	}
	return (PackageVersion)(result), true
}

// InstallPackage installs a package, returning an error if something went wrong
func (y *YumManager) InstallPackage(pkg string) error {
	if _, isInstalled := y.InstalledVersion(pkg); isInstalled {
		return nil
	}
	_, err := y.Sys.Run(fmt.Sprintf("yum install -y %s", pkg))
	return err
}

// RemovePackage removes a package, returning an error if something went wrong
func (y *YumManager) RemovePackage(pkg string) error {
	_, err := y.Sys.Run(fmt.Sprintf("yum remove -y %s", pkg))
	return err
}

func getExitCode(err error) (uint32, error) {
	if err == nil {
		return 0, nil
	}
	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		return 255, errors.Wrap(err, "not a valid exitError")
	}
	status, ok := exitErr.Sys().(syscall.WaitStatus)
	if !ok {
		return 255, errors.Wrap(err, "failed to get sys waitstatus")
	}
	return uint32(status.ExitStatus()), nil
}
