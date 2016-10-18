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

package unlock

import (
	"github.com/asteris-llc/converge/load/registry"
	"github.com/asteris-llc/converge/parse/preprocessor/lock"
	"github.com/asteris-llc/converge/resource"
)

// Preparer doesn't do anything for lock resources
type Preparer struct{}

// Prepare a new lock
func (p *Preparer) Prepare(render resource.Renderer) (resource.Task, error) {
	return &Unlock{}, nil
}

func init() {
	registry.Register(lock.GetUnlockKeyword(), (*Preparer)(nil), (*Unlock)(nil))
}
