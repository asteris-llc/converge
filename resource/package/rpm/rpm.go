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
	"github.com/asteris-llc/converge/resource"
	"golang.org/x/net/context"
)

// Package is an API for package state
type Package struct {
	Name   string
	State  State
	PkgMgr PackageManager
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
