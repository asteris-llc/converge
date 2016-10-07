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
	"github.com/asteris-llc/converge/load/registry"
	"github.com/asteris-llc/converge/resource"
)

// Preparer for APT Package
//
// APT Package manages system packages
type Preparer struct {
	// Name of the system package to be managed.
	Name string `hcl:"name" required:"true" `

	// State defines desired system package state.
	State State `hcl:"state" valid_values:"installed,absent" default:"installed"`
}

// Prepare a new package
func (p *Preparer) Prepare(render resource.Renderer) (resource.Task, error) {
	return &Package{
		Name:  p.Name,
		State: p.State,
	}, nil
}

func init() {
	registry.Register("apt.package", (*Preparer)(nil), (*Package)(nil))
}
