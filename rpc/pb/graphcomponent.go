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

package pb

// NewGraphComponent wraps the given component in a GraphComponent wrapper. If
// the argument is not something that can be wrapped, GraphComponent will not
// contain anything.
func NewGraphComponent(component interface{}) *GraphComponent {
	container := new(GraphComponent)
	switch c := component.(type) {
	case *GraphComponent_Edge:
		container.Component = &GraphComponent_Edge_{Edge: c}

	case *GraphComponent_Vertex:
		container.Component = &GraphComponent_Vertex_{Vertex: c}
	}

	return container
}
