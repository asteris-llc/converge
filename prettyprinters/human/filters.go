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

package human

import (
	"strings"

	"github.com/asteris-llc/converge/graph"
)

// FilterFunc is used to filter out nodes which shouldn't be printed. The func
// should return `true` if the node is to be printed.
type FilterFunc func(id string, value Printable) bool

// ShowOnlyChanged filters nodes to show only changed nodes
func ShowOnlyChanged(id string, value Printable) bool {
	return value.HasChanges()
}

// ShowEverything shows all nodes
func ShowEverything(string, Printable) bool {
	return true
}

// HideByKind hides certain ID types in the graph. So for example if you want
// to hide all params, specify `param` as a type to this func. This uses the ID
// functions in the graph module.
func HideByKind(types ...string) FilterFunc {
	return func(id string, value Printable) bool {
		for _, t := range types {
			if strings.HasPrefix(graph.BaseID(id), t) {
				return false
			}
		}

		return true
	}
}

// AndFilter chains FilterFuncs together
func AndFilter(a, b FilterFunc) FilterFunc {
	return func(id string, value Printable) bool {
		return a(id, value) && b(id, value)
	}
}
