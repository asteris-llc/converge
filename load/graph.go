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

package load

import (
	"fmt"
	"strings"

	"github.com/asteris-llc/converge/resource"
	"github.com/hashicorp/terraform/dag"
)

var (
	graphIDSeparator = "/"
)

// NewGraph returns a DAG of the given Module
func NewGraph(root *resource.Module) (*Graph, error) {
	g := &Graph{
		root:      root,
		resources: map[string]resource.Resource{},
		graph:     new(dag.AcyclicGraph),
	}

	return g, g.load()
}

// Graph represents the graph of resources loaded
type Graph struct {
	root      *resource.Module
	resources map[string]resource.Resource
	graph     *dag.AcyclicGraph
}

type ident struct {
	ID       string
	Resource resource.Resource
}

func (g *Graph) load() error {
	ids := []ident{{g.root.Name(), g.root}}

	for len(ids) > 0 {
		var id ident
		id, ids = ids[0], ids[1:]

		g.resources[id.ID] = id.Resource
		g.graph.Add(id.ID)

		if parent, ok := id.Resource.(resource.Parent); ok {
			for _, child := range parent.Children() {
				childID := ident{id.ID + graphIDSeparator + resource.LongName(child), child}
				g.graph.Add(childID.ID)
				g.graph.Connect(dag.BasicEdge(id.ID, childID.ID))
				ids = append(ids, childID)
			}
		}

		segments := strings.Split(id.ID, graphIDSeparator)
		for _, dep := range id.Resource.Depends() {
			// connect dep to siblings
			depPath := strings.Join(append(segments[:len(segments)-1], dep), graphIDSeparator)
			g.graph.Connect(dag.BasicEdge(id.ID, depPath))
		}
	}

	g.graph.TransitiveReduction()
	return g.graph.Validate()

}

func (g *Graph) String() string {
	return strings.TrimSpace(g.graph.String())
}

// GraphString returns the loaded graph as a GraphViz string
func (g *Graph) GraphString() string {
	s := "digraph {\n"

	for _, node := range g.graph.Vertices() {
		s += fmt.Sprintf(
			"  \"%s\"[label=\"%s\"];\n",
			node,
			resource.LongName(g.resources[node.(string)]),
		)
	}

	for _, edge := range g.graph.Edges() {

		s += fmt.Sprintf(
			"  \"%s\" -> \"%s\";\n",
			edge.Source(),
			edge.Target(),
		)
	}

	s += "}\n"

	return s
}

// Walk the graph, calling the specified function at each vertex
func (g *Graph) Walk(f func(path string, res resource.Resource) error) error {
	root, err := g.graph.Root()
	if err != nil {
		return err
	}

	return g.graph.DepthFirstWalk(
		[]dag.Vertex{root},
		func(path dag.Vertex, depth int) error {
			res := g.resources[path.(string)]
			return f(path.(string), res)
		},
	)
}

// Parent retrieves the parent module of a given path
func (g *Graph) Parent(path string) (parent *resource.Module, err error) {
	parts := strings.Split(path, graphIDSeparator)
	parentPath := strings.Join(parts[:len(parts)-1], graphIDSeparator)

	above, ok := g.resources[parentPath]
	if !ok {
		// having no parent is alright, it could be the root of the graph
		return nil, nil
	}

	parent, ok = above.(*resource.Module)
	if !ok {
		return nil, fmt.Errorf("bad parent for %q", path)
	}

	return parent, nil
}
