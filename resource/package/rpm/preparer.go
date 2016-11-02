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
	"github.com/asteris-llc/converge/load/registry"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/package"
)

// Preparer for RPM Package
//
// RPM Package manages system packages with `rpm` and `yum`. It assumes that
// both `rpm` and `yum` are installed on the system, and that the user has
// permissions to install, remove, and query packages.
type Preparer struct {
	// Name of the package or package group.
	Name string `hcl:"name" required:"true" `

	// State of the package. Present means the package will be installed if
	// missing; Absent means the package will be uninstalled if present.
	State State `hcl:"state" valid_values:"present,absent"`
}

// Prepare a new package
func (p *Preparer) Prepare(render resource.Renderer) (resource.Task, error) {
	if p.State == "" {
		p.State = "present"
	}

	return &Package{
		Name:   p.Name,
		State:  p.State,
		PkgMgr: &YumManager{Sys: pkg.ExecCaller{}},
	}, nil
}

func init() {
	registry.Register("package.rpm", (*Preparer)(nil), (*pkg.Package)(nil))
}
