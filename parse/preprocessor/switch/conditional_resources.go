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

package control

import (
	"github.com/asteris-llc/converge/resource"
	"golang.org/x/net/context"
)

// NopTask does nothing, verbosely
type NopTask struct {
	resource.Status
	Predicate string
}

// Check does nothing, verbosely
func (n *NopTask) Check(context.Context, resource.Renderer) (resource.TaskStatus, error) {
	n.AddMessage("Skipiping check; short-circuited in conditional")
	n.AddMessage("predicate: " + n.Predicate)
	return n, nil
}

// Apply does nothing, verbosely
func (n *NopTask) Apply(context.Context) (resource.TaskStatus, error) {
	n.AddMessage("Skipiping application; short-circuited in conditional")
	n.AddMessage("predicate: " + n.Predicate)
	return n, nil
}
