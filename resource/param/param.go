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

// Param controls parameter flow inside execution
type Param struct {
	Value string
}

// Check just returns the current value of the parameter. It should never have to change.
func (p *Param) Check() (string, bool, error) {
	return p.String(), false, nil
}

// Apply doesn't do anything since params are final values
func (*Param) Apply() error {
	return nil
}

// String is the final value of thie Param
func (p *Param) String() string {
	return p.Value
}
