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
	"strconv"
	"strings"

	"github.com/asteris-llc/converge/resource"
	"github.com/awalterschulze/gographviz"
	"github.com/hashicorp/terraform/dag"
)

// New returns a new Executor from a module
func New(mod *resource.Module) (*Executor, error) {
	e := &Executor{
		module:    mod,
		resources: map[string]resource.Resource{},
		graph:     &dag.AcyclicGraph{},
	}
	return e, e.dagify()
}

// Executor executes resources
// Should Executer be unexported since it has a constructor.
type Executor struct {
	module    *resource.Module
	resources map[string]resource.Resource
	graph     *dag.AcyclicGraph
	graphViz  *gographviz.Graph
}

type ident struct {
	ID       string
	Resource resource.Resource
}

// Now Using a GraphViz Library for prettier graphs
func (e *Executor) dagify() error {
	graphViz := gographviz.NewGraph()
	graphName := e.module.Name()
	graphViz.SetName(graphName)
	graphViz.SetDir(true)
	ids := []ident{{e.module.Name(), e.module}}
	for len(ids) > 0 {
		var id ident
		id, ids = ids[0], ids[1:]

		e.resources[id.ID] = id.Resource
		e.graph.Add(id.ID)

		if parent, ok := id.Resource.(resource.Parent); ok {
			for _, child := range parent.Children() {
				childID := ident{id.ID + "_" + child.Name(), child}
				e.graph.Add(childID.ID)
				e.graph.Connect(dag.BasicEdge(id.ID, childID.ID))
				ids = append(ids, childID)
			}
		}

	}

	walkDeptFunc := func(vert dag.Vertex, dept int) error {
		attrs := make(map[string]string)
		if dept == 0 {
			attrs["shape"] = "Msquare"
		} else {
			maxDept := dept
			if maxDept == 5 {
				maxDept = 5
			}
			attrs["peripheries"] = strconv.Itoa(maxDept)
		}
		graphViz.AddNode(graphName, vert.(string), attrs)
		return nil
	}

	e.graph.DepthFirstWalk(e.graph.Vertices()[:1], walkDeptFunc)

	for _, edge := range e.graph.Edges() {
		graphViz.AddEdge(edge.Source().(string), edge.Target().(string), true, nil)
	}
	e.graphViz = graphViz

	return nil
}

func (e *Executor) String() string {
	e.dagify()
	return strings.TrimSpace(e.graph.String())
}

// GraphString returns the loaded graph as a GraphViz string
func (e *Executor) GraphString() string {
	e.dagify()

	return e.graphViz.String()
}
