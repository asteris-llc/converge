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

package exec

import (
	"fmt"
	"strings"

	"github.com/asteris-llc/converge/resource"
	"github.com/hashicorp/terraform/dag"
)

// New returns a new Executor from a module
func New(mod *resource.Module) (*Executor, error) {
	e := &Executor{
		module:    mod,
		resources: map[string]resource.Resource{},
		graph:     new(dag.AcyclicGraph),
	}
	return e, e.dagify()
}

// Executor executes resources
type Executor struct {
	module    *resource.Module
	resources map[string]resource.Resource
	graph     *dag.AcyclicGraph
}

type ident struct {
	ID       string
	Resource resource.Resource
}

func (e *Executor) dagify() error {
	ids := []ident{{e.module.Name(), e.module}}

	for len(ids) > 0 {
		var id ident
		id, ids = ids[0], ids[1:]

		e.resources[id.ID] = id.Resource
		e.graph.Add(id.ID)

		if parent, ok := id.Resource.(resource.Parent); ok {
			for _, child := range parent.Children() {
				childID := ident{id.ID + "." + child.Name(), child}
				e.graph.Add(childID.ID)
				e.graph.Connect(dag.BasicEdge(id.ID, childID.ID))
				ids = append(ids, childID)
			}
		}
	}

	return nil
}

func (e *Executor) String() string {
	return strings.TrimSpace(e.graph.String())
}

// GraphString returns the loaded graph as a GraphViz string
func (e *Executor) GraphString() string {
	s := "digraph {\n"

	for _, node := range e.graph.Vertices() {
		s += fmt.Sprintf(
			"  \"%s\"[label=\"%s\"];\n",
			node,
			e.resources[node.(string)].Name(),
		)
	}

	for _, edge := range e.graph.Edges() {
		s += fmt.Sprintf(
			"  \"%s\" -> \"%s\";\n",
			edge.Source(),
			edge.Target(),
		)
	}

	s += "}\n"

	return s
}
