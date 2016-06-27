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
	"bytes"
	"fmt"
	"strings"
	"sync"

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

func genID(previousID, name string) string {
	return previousID + graphIDSeparator + name
}

func (g *Graph) load() error {
	ids := []ident{{g.root.String(), g.root}}

	for len(ids) > 0 {
		var id ident
		id, ids = ids[0], ids[1:]

		g.resources[id.ID] = id.Resource
		g.graph.Add(id.ID)

		if parent, ok := id.Resource.(resource.Parent); ok {
			for _, child := range parent.Children() {
				childID := ident{genID(id.ID, child.String()), child}
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

	return g.Validate()
}

func (g *Graph) String() string {
	return strings.TrimSpace(g.graph.String())
}

// GraphString returns the loaded graph as a GraphViz string
func (g *Graph) GraphString() (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)

	// create graph
	fmt.Fprintf(buf, "digraph %q {\n", g.root)
	defer buf.WriteString("}")

	lock := new(sync.Mutex)

	err := g.Walk(func(path string, res resource.Resource) error {
		lock.Lock()
		defer lock.Unlock()

		parent, ok := res.(resource.Parent)
		if !ok {
			return nil
		}

		label := path
		if path == g.root.String() {
			label += " (root)"
		}

		fmt.Fprintf(buf, "\tsubgraph \"cluster_%s\" {\n", path)
		fmt.Fprintf(buf, "\t\t%q[label=%q,shape=box];\n", path, label)
		defer fmt.Fprintf(buf, "\t}\n\n")

		for _, child := range parent.Children() {
			var color string
			switch child.(type) {
			case *resource.Param:
				color = "blue"
			default:
				color = "black"
			}

			fmt.Fprintf(
				buf,
				"\t\t%q[label=%q,color=%q];\n",
				genID(path, child.String()),
				child,
				color,
			)
		}

		return nil
	})

	if err != nil {
		return buf, err // this shouldn't ever happen. But just so we know if it *does*...
	}

	// add edges
	for _, edge := range g.graph.Edges() {
		fmt.Fprintf(buf, "\t%q -> %q;\n", edge.Source(), edge.Target())
	}

	return buf, nil
}

func escape(str string) string {
	return fmt.Sprintf("%q", str)
}

// Walk the graph, calling the specified function at each vertex
func (g *Graph) Walk(f func(path string, res resource.Resource) error) error {
	return g.graph.Walk(
		func(path dag.Vertex) error {
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

// Depends lists the dependencies of a vertex in the graph. It requires a fully
// qualified path to function properly.
func (g *Graph) Depends(path string) (paths []string) {
	list := g.graph.DownEdges(path).List()
	for _, item := range list {
		if path, ok := item.(string); ok {
			paths = append(paths, path)
		}
	}
	return paths
}

// Validate this graph
func (g *Graph) Validate() error {
	err := g.graph.Validate()
	if err != nil {
		return err
	}

	root, err := g.graph.Root()
	if err != nil {
		return err
	}

	return g.graph.DepthFirstWalk(
		[]dag.Vertex{root},
		func(path dag.Vertex, _ int) error {
			for _, dep := range g.graph.DownEdges(path).List() {
				if !g.graph.HasVertex(dep) {
					return fmt.Errorf(
						"Resource %q depends on resource %q, which does not exist",
						path,
						dep,
					)
				}
			}

			return nil
		},
	)
}
