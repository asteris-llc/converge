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

package graph

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"

	"github.com/hashicorp/terraform/dag"
)

// Graph is a generic graph structure that uses IDs to connect the graph
type Graph struct {
	inner  *dag.AcyclicGraph
	values map[string]interface{}
}

// New constructs and returns a new Graph
func New() *Graph {
	return &Graph{
		inner:  new(dag.AcyclicGraph),
		values: map[string]interface{}{},
	}
}

// Add a new value by ID
func (g *Graph) Add(id string, value interface{}) {
	g.inner.Add(id)
	g.values[id] = value
}

// Get a value by ID
func (g *Graph) Get(id string) interface{} {
	return g.values[id]
}

// GetParent returns the direct parent vertex of the current node. This only
// works if you're using the hierarchical ID functions from this module.
func (g *Graph) GetParent(id string) interface{} {
	return g.Get(ParentID(id))
}

// Connect two vertices together by ID
func (g *Graph) Connect(from, to string) {
	g.inner.Connect(dag.BasicEdge(from, to))
}

// DownEdges returns outward-facing edges of the specified vertex
func (g *Graph) DownEdges(id string) (out []string) {
	for _, edge := range g.inner.DownEdges(id).List() {
		out = append(out, edge.(string))
	}

	return out
}

// Walk the graph leaf-to-root
func (g *Graph) Walk(cb func(string, interface{}) error) error {
	return g.inner.Walk(func(v dag.Vertex) error {
		id, ok := v.(string)
		if !ok {
			// something has gone horribly wrong
			return fmt.Errorf(`ID "%v" was not a string`, v)
		}

		return cb(id, g.values[id])
	})
}

// Transform a graph of type A to a graph of type B. A and B can be the same.
func (g *Graph) Transform(cb func(string, interface{}, []string) (interface{}, []string, error)) (transformed *Graph, err error) {
	transformed = New()
	lock := new(sync.Mutex)

	err = g.Walk(func(id string, value interface{}) error {
		edges := g.DownEdges(id)

		newValue, newEdges, err := cb(id, value, edges)
		if err != nil {
			return err
		}

		lock.Lock()
		defer lock.Unlock()
		transformed.Add(id, newValue)

		for _, edge := range newEdges {
			transformed.Connect(id, edge)
		}

		return nil
	})
	if err != nil {
		return transformed, err
	}

	return transformed, transformed.Validate()
}

// Validate that the graph...
//
// 1. has a root
// 2. has no cycles
// 3. has no dangling edges
// 4. has values of a single type (excepting nil)
func (g *Graph) Validate() error {
	err := g.inner.Validate()
	if err != nil {
		return err
	}

	// check for dangling dependencies
	var bad []string
	for _, edge := range g.inner.Edges() {
		if !g.inner.HasVertex(edge.Source()) {
			bad = append(bad, edge.Source().(string))
		}

		if !g.inner.HasVertex(edge.Target()) {
			bad = append(bad, edge.Target().(string))
		}
	}

	if bad != nil {
		return fmt.Errorf(
			"nonexistent vertices in edges: %s",
			strings.Join(bad, ", "),
		)
	}

	// check for differing types
	types := map[reflect.Type]struct{}{}
	for _, val := range g.values {
		if val == nil {
			continue
		}

		types[reflect.TypeOf(val)] = struct{}{}
	}

	if len(types) > 1 {
		var names []string
		for t := range types {
			names = append(names, fmt.Sprint(t))
		}
		sort.Strings(names)

		return fmt.Errorf(
			"differing types in graph vertices: %s",
			strings.Join(names, ", "),
		)
	}

	return nil
}

func (g *Graph) String() string {
	return strings.Trim(g.inner.String(), "\n")
}
