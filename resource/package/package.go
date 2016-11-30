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

package pkg

import (
	"os/exec"
	"syscall"

	"github.com/asteris-llc/converge/resource"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

// PackageVersion is a type alias for a string, since various packages may use
// different naming conventions
type PackageVersion string

// State type for Package
type State string

const (
	// StatePresent indicates the package should be present
	StatePresent State = "present"

	// StateAbsent indicates the package should be absent
	StateAbsent State = "absent"
)

// PackageManager describes an interface for managing packages and helps make
// `Check` and `Apply` testable.
type PackageManager interface {
	// If the package is installed, returns the version and true, otherwise
	// returns an empty string and false.
	InstalledVersion(string) (PackageVersion, bool)

	// Installs a package, returning an error if something went wrong
	InstallPackage(string) (string, error)

	// Removes a package, returning an error if something went wrong
	RemovePackage(string) (string, error)
}

// Package is an API for package state
type Package struct {
	Name   string `export:"name"`
	State  State  `export:"state"`
	PkgMgr PackageManager
	*resource.Status
}

// SysCaller allows us to mock exec.Command
type SysCaller interface {
	Run(string) ([]byte, error)
}

// ExecCaller is a dummy struct to handle wrapping exec.Command in the SysCaller
// interface.
type ExecCaller struct{}

// Run executes `cmd` as a /bin/sh script and returns the output and error
func (e ExecCaller) Run(cmd string) ([]byte, error) {
	return exec.Command("sh", "-c", cmd).Output()
}

// GetExitCode returns the exit code of an error
func GetExitCode(err error) (uint32, error) {
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

// Check if the package has to be 'present', or 'absent'
func (p *Package) Check(context.Context, resource.Renderer) (resource.TaskStatus, error) {
	p.Status = resource.NewStatus()
	if p.State == p.PackageState() {
		return p, nil
	}
	p.Status.AddDifference(p.Name, string(p.PackageState()), string(p.State), "")
	p.RaiseLevel(resource.StatusWillChange)
	return p, nil
}

// Apply desired package state
func (p *Package) Apply(context.Context) (resource.TaskStatus, error) {
	var err error
	p.Status = resource.NewStatus()
	if p.State == p.PackageState() {
		return p, nil
	}

	var results string
	if p.State == StatePresent {
		results, err = p.PkgMgr.InstallPackage(p.Name)
		p.Status.AddMessage("installed " + p.Name)
	} else {
		results, err = p.PkgMgr.RemovePackage(p.Name)
		p.Status.AddMessage("removed  " + p.Name)
	}

	p.Status.AddMessage(results)
	if err != nil {
		return p, err
	}
	p.Status.AddDifference(p.Name, string(p.PackageState()), string(p.State), "")
	p.RaiseLevel(resource.StatusWillChange)
	return p, nil
}

// PackageState returns a State ("present","absent") based on whether a package
// is installed or not.
func (p *Package) PackageState() State {
	if _, installed := p.PkgMgr.InstalledVersion(p.Name); installed {
		return StatePresent
	}
	return StateAbsent
}
