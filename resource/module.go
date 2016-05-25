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

package resource

// Module is the container for tasks and is the basic compositional unit of the
// system.
type Module struct {
	ModuleTask
	Params    map[string]*Param `hcl:"param"`
	Resources []Resource
}

// Name returns name for metadata
func (m *Module) Name() string {
	return m.ModuleName
}

// Validate checks shell tasks validity
func (m *Module) Validate() ParamError {
	return ParamError{Field: "", Error: nil}
}

// Children returns the managed resources under this module
func (m *Module) Children() []Resource {
	return m.Resources
}
