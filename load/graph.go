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
	"github.com/awalterschulze/gographviz"
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

	return g.graph.Validate()
}

func (g *Graph) String() string {
	return strings.TrimSpace(g.graph.String())
}

// GraphString returns the loaded graph as a GraphViz string
func (g *Graph) GraphString() string {

	//Create Graph
	graphViz := gographviz.NewGraph()
	graphName := escape(g.root.String())
	graphViz.SetName(graphName)
	graphViz.SetDir(true)

	//Add all Nodes and Edges
	for _, node := range g.graph.Vertices() {
		attrs := map[string]string{"label": escape(g.resources[node.(string)].String())}
		graphViz.AddNode(graphName, escape(node.(string)), attrs)
	}
	for _, edge := range g.graph.Edges() {
		graphViz.AddEdge(escape(edge.Source().(string)), escape(edge.Target().(string)), true, nil)
	}

	//Create Subgraphs
	var subGraphFromModule func(rootID string, mod *resource.Module)
	subGraphFromModule = func(rootID string, mod *resource.Module) {
		rootID = genID(rootID, mod.String()) //ID of this resource
		var (
			name     = "cluster_module_" + mod.String() //name of this cluster. "cluster_" prefix required
			children = mod.Children()                   //resources in this module
		)
		//Make sure to include this node in the graph
		graphViz.AddNode(name, escape(rootID), nil)
		for i := 0; i < len(children)-1; i++ {
			current := escape(genID(rootID, children[i].String()))
			nxtChild := escape(genID(rootID, children[i+1].String()))

			//Using a switch case in case different resources get different attributes
			switch v := children[i].(type) {
			case *resource.Module:
				subGraphFromModule(rootID, v)
				graphViz.AddSubGraph(name, "cluster_module_"+v.String(), nil)
			}
			graphViz.AddNode(name, current, nil)
			graphViz.AddNode(name, nxtChild, nil)
			graphViz.AddEdge(current, nxtChild, true, nil)
		}
	}
	//Add a cluster
	for _, res := range g.root.Children() {
		if mod, ok := res.(*resource.Module); ok {
			subGraphFromModule(g.root.String(), mod)
			graphViz.AddSubGraph(graphName, "cluster_module_"+mod.String(), nil)
		}
	}

	//Set Attributes for Subgraphs
	for name, sub := range graphViz.SubGraphs.SubGraphs {
		sub.Attrs.Add("style", "filled")
		sub.Attrs.Add("color", "lightgrey")
		sub.Attrs.Add("label", name)
	}
	return graphViz.String()
}
func escape(str string) string {
	return fmt.Sprintf("%q", str)
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
