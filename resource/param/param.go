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

import (
	"fmt"

	"github.com/asteris-llc/converge/resource"
	"golang.org/x/net/context"
)

// Param controls parameter flow inside execution
type Param struct {
	resource.Status

	Val interface{}
}

// Check just returns the current value of the parameter. It should never have to change.
func (p *Param) Check(context.Context, resource.Renderer) (resource.TaskStatus, error) {
	p.Status = resource.Status{Output: []string{p.String()}}

	return p, nil
}

// Apply doesn't do anything since params are final values
func (p *Param) Apply(context.Context) (resource.TaskStatus, error) {
	return p, nil
}

// String is the final value of this Param
func (p *Param) String() string {
	return fmt.Sprintf("%v", p.Val)
}
