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
	"log"
	"strings"
	"sync"

	"github.com/hashicorp/terraform/dag"
)

// WalkFunc is taken by the walking functions
type WalkFunc func(string, interface{}) error

type walkerFunc func(WalkFunc) error

// Graph is a generic graph structure that uses IDs to connect the graph
type Graph struct {
	inner  *dag.AcyclicGraph
	values map[string]interface{}

	innerLock  *sync.RWMutex
	valuesLock *sync.RWMutex
}

// New constructs and returns a new Graph
func New() *Graph {
	return &Graph{
		inner:      new(dag.AcyclicGraph),
		values:     map[string]interface{}{},
		innerLock:  new(sync.RWMutex),
		valuesLock: new(sync.RWMutex),
	}
}

// Add a new value by ID
func (g *Graph) Add(id string, value interface{}) {
	g.innerLock.Lock()
	defer g.innerLock.Unlock()

	g.valuesLock.Lock()
	defer g.valuesLock.Unlock()

	g.inner.Add(id)
	g.values[id] = value
}

// Get a value by ID
func (g *Graph) Get(id string) interface{} {
	g.valuesLock.RLock()
	defer g.valuesLock.RUnlock()

	return g.values[id]
}

// GetParent returns the direct parent vertex of the current node. This only
// works if you're using the hierarchical ID functions from this module.
func (g *Graph) GetParent(id string) interface{} {
	return g.Get(ParentID(id))
}

// GetSibling returns the named sibling of the current node. This only works if
// you're using the hierarchical ID functions from this module.
func (g *Graph) GetSibling(id, sibling string) interface{} {
	return g.Get(SiblingID(id, sibling))
}

// Connect two vertices together by ID
func (g *Graph) Connect(from, to string) {
	g.innerLock.Lock()
	defer g.innerLock.Unlock()

	g.inner.Connect(dag.BasicEdge(from, to))
}

// DownEdges returns outward-facing edges of the specified vertex
func (g *Graph) DownEdges(id string) (out []string) {
	g.innerLock.RLock()
	defer g.innerLock.RUnlock()

	for _, edge := range g.inner.DownEdges(id).List() {
		out = append(out, edge.(string))
	}

	return out
}

// Walk the graph leaf-to-root
func (g *Graph) Walk(cb WalkFunc) error {
	return g.inner.Walk(func(v dag.Vertex) error {
		id, ok := v.(string)
		if !ok {
			// something has gone horribly wrong
			return fmt.Errorf(`ID "%v" was not a string`, v)
		}

		return cb(id, g.values[id])
	})
}

// RootFirstWalk walks the graph root-to-leaf, checking sibling dependencies
// before descending.
func (g *Graph) RootFirstWalk(cb WalkFunc) error {
	root, err := g.inner.Root()
	if err != nil {
		return err
	}

	var (
		todo = []string{root.(string)}
		done = map[string]struct{}{}
	)

	for len(todo) > 0 {
		id := todo[0]
		todo = todo[1:]

		// first check if we've already done this ID. We check multiple times as a
		// signal to re-check after finding a dependency needs waiting for.
		if _, ok := done[id]; ok {
			continue
		}

		// make sure all sibling dependencies are finished first
		var skip bool
		for _, edge := range g.DownEdges(id) {
			if _, ok := done[edge]; AreSiblingIDs(id, edge) && !ok {
				log.Printf("[DEBUG] walk(rootfirst): %q still waiting for sibling %q", id, edge)
				todo = append(todo, id)
				skip = true
			}
		}
		if skip {
			continue
		}

		log.Printf("[DEBUG] walk(rootfirst): walking %s\n", id)

		err := cb(id, g.Get(id))
		if err != nil {
			return err
		}

		// mark this ID as done and do the children
		done[id] = struct{}{}
		for _, edge := range g.DownEdges(id) {
			todo = append(todo, edge)
		}
	}

	return nil
}

// Transform a graph of type A to a graph of type B. A and B can be the same.
func (g *Graph) Transform(cb func(string, *Graph) error) (*Graph, error) {
	t := g.Copy()

	return transform(t, t.Walk, func(id string, _ interface{}) error { return cb(id, t) })
}

// RootFirstTransform does Transform, but starting at the root
func (g *Graph) RootFirstTransform(cb func(string, *Graph) error) (*Graph, error) {
	t := g.Copy()

	return transform(t, t.RootFirstWalk, func(id string, _ interface{}) error { return cb(id, t) })
}

// Copy the graph for further modification
func (g *Graph) Copy() *Graph {
	out := New()

	err := g.Walk(func(id string, val interface{}) error {
		out.Add(id, val)
		for _, dest := range g.DownEdges(id) {
			out.Connect(id, dest)
		}

		return nil
	})
	if err != nil {
		panic(err)
	}

	return out
}

// Validate that the graph...
//
// 1. has a root
// 2. has no cycles
// 3. has no dangling edges
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

	return nil
}

func (g *Graph) String() string {
	return strings.Trim(g.inner.String(), "\n")
}

func transform(target *Graph, walker walkerFunc, cb WalkFunc) (*Graph, error) {
	if err := walker(cb); err != nil {
		return target, err
	}

	return target, target.Validate()
}
