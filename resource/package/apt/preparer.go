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
	"errors"
	"strings"

	"github.com/asteris-llc/converge/load/registry"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/package"
	"golang.org/x/net/context"
)

// Preparer for Debian Package
//
// Apt Package manages system packages with `apt` and `dpkg`. It assumes that
// both `apt` and `dpkg` are installed on the system, and that the user has
// permissions to install, remove, and query packages.
type Preparer struct {
	// Name of the package or package group.
	Name string `hcl:"name" required:"true" nonempty:"true"`

	// State of the package. Present means the package will be installed if
	// missing; Absent means the package will be uninstalled if present.
	State pkg.State `hcl:"state" valid_values:"present,absent"`
}

// Prepare a new package
func (p *Preparer) Prepare(ctx context.Context, render resource.Renderer) (resource.Task, error) {

	if strings.TrimSpace(p.Name) == "" {
		return &pkg.Package{}, errors.New("package name cannot be empty")
	}

	if p.State == "" {
		p.State = "present"
	}

	return &pkg.Package{
		Name:   p.Name,
		State:  p.State,
		PkgMgr: &Manager{Sys: pkg.ExecCaller{}},
	}, nil
}

func init() {
	registry.Register("package.apt", (*Preparer)(nil), (*pkg.Package)(nil))
}
