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

package param

import "github.com/asteris-llc/converge/resource"

// Preparer for params
type Preparer struct {
	Default string `hcl:"default"`
}

// Prepare a new task
func (p *Preparer) Prepare(render resource.RenderFunc) (resource.Task, error) {
	// TODO: this isn't quite right. We need to potentially pass in another set of
	// values instead of just taking the default? This should probably happen in
	// the call to this function
	def, err := render("default", p.Default)
	if err != nil {
		return nil, err
	}

	return &Param{Value: def}, nil
}
