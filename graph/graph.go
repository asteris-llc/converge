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
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/asteris-llc/converge/graph/node"
	"github.com/asteris-llc/converge/helpers/logging"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform/dag"
	"github.com/pkg/errors"
	cmap "github.com/streamrail/concurrent-map"
)

// WalkFunc is taken by the walking functions
type WalkFunc func(string, *node.Node) error

// TransformFunc is taken by the transformation functions
type TransformFunc func(string, *Graph) error

type walkerFunc func(context.Context, *Graph, WalkFunc) error

// An Edge is a generic pair of IDs indicating a directed edge in the graph
type Edge struct {
	Source     string   `json:"source"`
	Dest       string   `json:"dest"`
	Attributes []string `json:"attributes"`
}

// Graph is a generic graph structure that uses IDs to connect the graph
type Graph struct {
	inner  *dag.AcyclicGraph
	values cmap.ConcurrentMap

	innerLock *sync.RWMutex
}

// New constructs and returns a new Graph
func New() *Graph {
	return &Graph{
		inner:     new(dag.AcyclicGraph),
		values:    cmap.New(),
		innerLock: new(sync.RWMutex),
	}
}

// Add a new value by ID
func (g *Graph) Add(node *node.Node) {
	g.innerLock.Lock()
	defer g.innerLock.Unlock()

	g.inner.Add(node.ID)
	g.values.Set(node.ID, node)
}

// Remove an existing value by ID
func (g *Graph) Remove(id string) {
	g.innerLock.Lock()
	defer g.innerLock.Unlock()

	g.inner.Remove(id)
	g.values.Remove(id)
}

// Get a value by ID
func (g *Graph) Get(id string) *node.Node {
	val, _ := g.values.Get(id)

	if val == nil {
		return nil
	}

	return val.(*node.Node)
}

// GetParent returns the direct parent vertex of the current node.
func (g *Graph) GetParent(id string) *node.Node {
	var parentID string
	for _, edge := range g.UpEdges(id) {
		switch edge.(type) {
		case *ParentEdge:
			parentID = edge.Source().(string)
			break
		}
	}

	return g.Get(parentID)
}

// ConnectParent connects a parent node to a child node
func (g *Graph) ConnectParent(from, to string) {
	g.innerLock.Lock()
	defer g.innerLock.Unlock()

	g.inner.Connect(NewParentEdge(from, to))
}

// Connect two vertices together by ID
func (g *Graph) Connect(from, to string) {
	g.innerLock.Lock()
	defer g.innerLock.Unlock()

	g.inner.Connect(dag.BasicEdge(from, to))
}

// Disconnect two vertices by IDs
func (g *Graph) Disconnect(from, to string) {
	g.innerLock.Lock()
	defer g.innerLock.Unlock()

	g.inner.RemoveEdge(dag.BasicEdge(from, to))
}

// UpEdges returns inward-facing edges of the specified vertex
func (g *Graph) UpEdges(id string) (out []dag.Edge) {
	g.innerLock.RLock()
	defer g.innerLock.RUnlock()

	for _, edge := range g.inner.Edges() {
		if edge.Target().(string) == id {
			out = append(out, edge)
		}
	}

	return out
}

// DownEdges returns outward-facing edges of the specified vertex
func (g *Graph) DownEdges(id string) (out []dag.Edge) {
	g.innerLock.RLock()
	defer g.innerLock.RUnlock()

	for _, edge := range g.inner.Edges() {
		if edge.Source().(string) == id {
			out = append(out, edge)
		}
	}

	return out
}

// Descendents gets a list of all descendents (not just children, everything)
// This only works if you're using the hierarchical ID functions from this
// module.
func (g *Graph) Descendents(id string) (out []string) {
	g.innerLock.RLock()
	defer g.innerLock.RUnlock()

	for _, node := range g.inner.Vertices() {
		if IsDescendentID(id, node.(string)) {
			out = append(out, node.(string))
		}
	}

	return out
}

// Dependencies gets a list of all dependencies without relying on the ID
// functions and will work for dependencies that have been added during
// load.ResolveDependencies
func (g *Graph) Dependencies(id string) []string {
	var uniq []string
	for key := range g.dependencies(id, make(map[string]struct{})) {
		uniq = append(uniq, key)
	}
	return uniq
}

// internal version of dependencies with a carry map
func (g *Graph) dependencies(id string, carry map[string]struct{}) map[string]struct{} {
	for _, edge := range g.DownEdges(id) {
		elem := edge.Target().(string)
		carry[elem] = struct{}{}
		carry = g.dependencies(elem, carry)
	}
	return carry
}

// Walk the graph leaf-to-root
func (g *Graph) Walk(ctx context.Context, cb WalkFunc) error {
	return dependencyWalk(ctx, g, cb)
}

// dependencyWalk walks a graph leaf-to-root respecting dependencies
func dependencyWalk(rctx context.Context, g *Graph, cb WalkFunc) error {
	// the basic idea of this implementation is that we want to defer schedule
	// children of any given node until after that node's non-child dependencies
	// are satisfied. We're going to have a couple major components of this.
	// First, a scheduler/latch to make sure we don't schedule work more than
	// once. We also need the workers themselves, which take care of waiting for
	// their own dependencies and executing the callback for their node once the
	// dependencies are satisfied.
	root, err := g.Root()
	if err != nil {
		return err
	}

	logger := logging.GetLogger(rctx).WithField("function", "dependencyWalk")

	logger.Debug("started")

	// errors
	var (
		errLock      = new(sync.RWMutex)
		errs         = map[string]error{}
		errDepFailed = errors.New("dependency failed")
	)
	getErr := func(id string) error {
		errLock.RLock()
		defer errLock.RUnlock()
		return errs[id]
	}
	setErr := func(id string, err error) {
		errLock.Lock()
		defer errLock.Unlock()
		errs[id] = err
	}

	// tracking which dependencies have finished
	done := map[string]chan struct{}{}
	for _, id := range g.Vertices() {
		done[id] = make(chan struct{}, 0)
	}

	// create a child context out of the parent we receive. We'll use this to
	// make everything cancellable.
	ctx, cancel := context.WithCancel(rctx)
	defer cancel()

	wait := new(sync.WaitGroup)

	// keep track of what we've scheduled so we don't schedule the same work
	// twice
	var worker func(id string)
	scheduler := make(chan string)
	go func() {
		logger.Debug("starting scheduler")
		// it's OK to leave this unguarded by a lock, since we're only accessing
		// it in a single thread. If this algorithm ever changes to schedule
		// work in parallel, this should be protected by a lock (and the lock
		// should be held until the work is completely scheduled)
		scheduled := map[string]struct{}{}

		for {
			select {
			case <-ctx.Done():
				logger.Debug("stopping scheduler")
				return

			case id := <-scheduler:
				if _, ok := scheduled[id]; !ok {
					logger.WithField("id", id).Debug("scheduling")
					scheduled[id] = struct{}{}
					go worker(id)
				} else {
					logger.WithField("id", id).Debug("already scheduled")
				}
			}
		}
	}()

	// utility function to wait for a list of IDs
	waitFor := func(ids []string) error {
		for _, id := range ids {
			depChan, ok := done[id]
			if !ok {
				return fmt.Errorf("%q did not have done channel", id)
			}

			logger.WithField("id", id).Debug("waiting for id")
			select {
			case <-ctx.Done():
				return nil

			case <-depChan:
				if err := getErr(id); err != nil {
					return err
				}
			}
		}

		return nil
	}

	worker = func(id string) {
		wait.Add(1)
		defer wait.Done()

		logger.WithField("id", id).Debug("starting worker")

		var deps, children []string
		for _, edge := range g.DownEdges(id) {
			switch edge.(type) {
			case *ParentEdge:
				children = append(children, edge.Target().(string))
			default:
				deps = append(deps, edge.Target().(string))
			}
		}

		myDone, ok := done[id]
		if !ok {
			setErr(id, errors.New("could not get done channel"))
			return
		}
		defer close(myDone)

		// schedule deps - this prevents against the case where only Connect has
		// been used and there is no lineage information in the graph. If this
		// isn't here we'll be waiting for dependencies that never got scheduled
		// below.
		for _, dep := range deps {
			select {
			case <-ctx.Done():
				return
			case scheduler <- dep:
			}
		}

		if err := waitFor(deps); err != nil {
			setErr(id, errDepFailed)
			return
		}

		for _, child := range children {
			select {
			case <-ctx.Done():
				return
			case scheduler <- child:
			}
		}

		if err := waitFor(children); err != nil {
			setErr(id, errDepFailed)
			return
		}

		logger.WithField("id", id).Debug("executing")
		if err := cb(id, g.Get(id)); err != nil {
			setErr(id, err)
		}
	}

	worker(root)

	wait.Wait()

	// construct error
	if len(errs) > 0 {
		var err error
		for k, v := range errs {
			if v == errDepFailed {
				continue
			}
			err = multierror.Append(err, errors.Wrap(v, k))
		}
		return err
	}
	return nil
}

// RootFirstWalk walks the graph root-to-leaf, checking sibling dependencies
// before descending.
func (g *Graph) RootFirstWalk(ctx context.Context, cb WalkFunc) error {
	return rootFirstWalk(ctx, g, cb)
}

// rootFirstWalk is separate for internal use in the transformations
func rootFirstWalk(ctx context.Context, g *Graph, cb WalkFunc) error {
	root, err := g.inner.Root()
	if err != nil {
		return err
	}

	logger := logging.GetLogger(ctx).WithField("function", "rootFirstWalk")

	var (
		todo = []string{root.(string)}
		done = map[string]struct{}{}
	)

	for len(todo) > 0 {
		id := todo[0]
		todo = todo[1:]

		select {
		case <-ctx.Done():
			return fmt.Errorf("interrupted at %q", id)
		default:
		}

		// first check if we've already done this ID. We check multiple times as a
		// signal to re-check after finding a dependency needs waiting for.
		if _, ok := done[id]; ok {
			continue
		}

		// make sure all sibling dependencies are finished first
		var skip bool
		for _, edge := range g.DownEdges(id) {
			if _, ok := done[edge.Target().(string)]; AreSiblingIDs(id, edge.Target().(string)) && !ok {
				logger.WithField("id", id).WithField("target", edge).Debug("still waiting for sibling")
				todo = append(todo, id)
				skip = true
			}
		}
		if skip {
			continue
		}

		logger.WithField("id", id).Debug("walking")

		if err := cb(id, g.Get(id)); err != nil {
			return err
		}

		// mark this ID as done and do the children
		done[id] = struct{}{}
		for _, edge := range g.DownEdges(id) {
			todo = append(todo, edge.Target().(string))
		}
	}

	return nil
}

// Transform a graph of type A to a graph of type B. A and B can be the same.
func (g *Graph) Transform(ctx context.Context, cb TransformFunc) (*Graph, error) {
	return transform(ctx, g, dependencyWalk, cb)
}

// RootFirstTransform does Transform, but starting at the root
func (g *Graph) RootFirstTransform(ctx context.Context, cb TransformFunc) (*Graph, error) {
	return transform(ctx, g, rootFirstWalk, cb)
}

// Copy the graph for further modification
func (g *Graph) Copy() *Graph {
	out := New()

	for _, v := range g.Vertices() {
		out.Add(g.Get(v))
	}

	for _, e := range g.inner.Edges() {
		out.inner.Connect(e)
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

// Vertices will det a list of the IDs for every vertex in the graph, cast to a
// string.
func (g *Graph) Vertices() []string {
	graphVertices := g.inner.Vertices()
	vertices := make([]string, len(graphVertices))
	for v := range graphVertices {
		vertices[v] = graphVertices[v].(string)
	}
	return vertices
}

// MaybeGet returns the value of the element and a bool indicating if it was
// found, if it was not found the value of the returned element is nil.
func (g *Graph) MaybeGet(id string) (*node.Node, bool) {
	raw, ok := g.values.Get(id)
	if !ok {
		return nil, ok
	}

	return raw.(*node.Node), true
}

// Contains returns true if the id exists in the map
func (g *Graph) Contains(id string) bool {
	_, found := g.MaybeGet(id)
	return found
}

// Edges will get a list of all of the edges in the graph.
func (g *Graph) Edges() []Edge {
	graphEdges := g.inner.Edges()
	edges := make([]Edge, len(graphEdges))
	for idx, srcEdge := range graphEdges {
		edge := Edge{
			Source: srcEdge.Source().(string),
			Dest:   srcEdge.Target().(string),
		}

		if _, ok := srcEdge.(*ParentEdge); ok {
			edge.Attributes = append(edge.Attributes, "parent")
		}

		edges[idx] = edge
	}
	return edges
}

// Root will get the root element of the graph
func (g *Graph) Root() (string, error) {
	r, err := g.inner.Root()
	if err != nil {
		return "", err
	}
	return r.(string), nil
}

func (g *Graph) String() string {
	return strings.Trim(g.inner.String(), "\n")
}

func transform(ctx context.Context, source *Graph, walker walkerFunc, cb TransformFunc) (*Graph, error) {
	dest := source.Copy()

	err := walker(ctx, dest, func(id string, _ *node.Node) error {
		return cb(id, dest)
	})
	if err != nil {
		return dest, err
	}

	return dest, dest.Validate()
}
