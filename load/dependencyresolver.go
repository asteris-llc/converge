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
	"context"
	"fmt"
	"io/ioutil"
	"sort"
	"sync"
	"text/template"

	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/graph/node"
	"github.com/asteris-llc/converge/helpers/logging"
	"github.com/asteris-llc/converge/parse"
	"github.com/asteris-llc/converge/render/extensions"
	"github.com/asteris-llc/converge/render/preprocessor"
	"github.com/pkg/errors"
)

type dependencyGenerator func(g *graph.Graph, id string, node *parse.Node) ([]string, error)

// ResolveDependencies examines the strings and depdendencies at each vertex of
// the graph and creates edges to fit them
func ResolveDependencies(ctx context.Context, g *graph.Graph) (*graph.Graph, error) {
	logger := logging.GetLogger(ctx).WithField("function", "ResolveDependencies")
	logger.Debug("resolving dependencies")

	groupLock := new(sync.Mutex)
	groupMap := make(map[string]struct{})
	g, err := g.Transform(ctx, func(meta *node.Node, out *graph.Graph) error {
		if graph.IsRoot(meta.ID) { // skip root
			return nil
		}

		node, ok := meta.Value().(*parse.Node)
		if !ok {
			return fmt.Errorf("ResolveDependencies can only be used on Graphs of *parse.Node. I got %T", meta.Value())
		}

		depGenerators := []dependencyGenerator{getDepends, getParams, getXrefs}

		// we have dependencies from various sources, but they're always IDs, so we
		// can connect them pretty easily
		for _, source := range depGenerators {
			deps, err := source(g, meta.ID, node)
			if err != nil {
				return err
			}
			for _, dep := range deps {
				if err := out.SafeConnect(meta.ID, dep); err != nil {
					logger.Error(err)
					return err
				}
			}
		}

		// collect group information
		if meta.Group != "" {
			groupLock.Lock()
			groupMap[meta.Group] = struct{}{}
			groupLock.Unlock()
		}

		return nil
	})

	for group := range groupMap {
		groupDeps(ctx, g, group)
	}
	return g, err
}

func getDepends(g *graph.Graph, id string, node *parse.Node) ([]string, error) {
	deps, err := node.GetStringSlice("depends")
	switch err {
	case parse.ErrNotFound:
		return []string{}, nil
	case nil:
		for idx, dep := range deps {
			if ancestor, ok := getNearestAncestor(g, id, dep); ok {
				deps[idx] = ancestor
			} else {
				return nil, fmt.Errorf("nonexistent vertices in edges: %s", dep)
			}
		}
		return deps, nil
	default:
		return nil, err
	}
}

func getParams(g *graph.Graph, id string, node *parse.Node) (out []string, err error) {
	var nodeStrings []string
	nodeStrings, err = node.GetStrings()
	if err != nil {
		return nil, err
	}

	type stub struct{}
	language := extensions.MinimalLanguage()
	language.On("param", extensions.RememberCalls(&out, ""))
	language.On("paramList", extensions.RememberCalls(&out, []interface{}(nil)))
	language.On("paramMap", extensions.RememberCalls(&out, map[string]interface{}(nil)))

	for _, s := range nodeStrings {
		useless := stub{}
		tmpl, tmplErr := template.New("DependencyTemplate").Funcs(language.Funcs).Parse(s)
		if tmplErr != nil {
			return out, tmplErr
		}
		tmpl.Execute(ioutil.Discard, &useless)
	}
	for idx, val := range out {
		ancestor, found := getNearestAncestor(g, id, "param."+val)
		if !found {
			return out, fmt.Errorf("unknown parameter: param.%s", val)
		}
		out[idx] = ancestor
	}
	return out, err
}

func getXrefs(g *graph.Graph, id string, node *parse.Node) (out []string, err error) {
	var nodeStrings []string
	var calls []string
	nodeRefs := make(map[string]struct{})
	nodeStrings, err = node.GetStrings()
	if err != nil {
		return nil, err
	}
	language := extensions.MinimalLanguage()
	language.On(extensions.RefFuncName, extensions.RememberCalls(&calls, 0))
	for _, s := range nodeStrings {
		tmpl, tmplErr := template.New("DependencyTemplate").Funcs(language.Funcs).Parse(s)
		if tmplErr != nil {
			return out, tmplErr
		}
		tmpl.Execute(ioutil.Discard, &struct{}{})
	}
	for _, call := range calls {
		vertex, _, found := preprocessor.VertexSplitTraverse(g, call, id, preprocessor.TraverseUntilModule, make(map[string]struct{}))
		if !found {
			return []string{}, fmt.Errorf("dependency generator: unresolvable call to %s", call)
		}
		if _, ok := nodeRefs[vertex]; !ok {
			nodeRefs[vertex] = struct{}{}
			out = append(out, vertex)
			if peerVertex, ok := getPeerVertex(id, vertex); ok {
				out = append(out, peerVertex)
			}
		}
	}
	return out, err
}

func getPeerVertex(src, dst string) (string, bool) {
	if dst == "." || graph.IsRoot(dst) {
		return "", false
	}
	if graph.AreSiblingIDs(src, dst) {
		return dst, true
	}
	return getPeerVertex(src, graph.ParentID(dst))
}

func getNearestAncestor(g *graph.Graph, id, node string) (string, bool) {
	if graph.IsRoot(id) || id == "" || id == "." {
		return "", false
	}

	siblingID := graph.SiblingID(id, node)

	valMeta, ok := g.Get(siblingID)
	if !ok {
		return getNearestAncestor(g, graph.ParentID(id), node)
	}
	_, ok = valMeta.Value().(*parse.Node)
	if !ok {
		return "", false
	}
	return siblingID, true
}

func withoutRoot(in []string) (out []string) {
	for _, id := range in {
		if !graph.IsRoot(id) {
			out = append(out, id)
		}
	}
	return out
}

func withoutModule(g *graph.Graph, in []string) (out []string) {
	for _, id := range in {
		if meta, ok := g.Get(id); ok {
			if node, ok := meta.Value().(*parse.Node); ok {
				if !node.IsModule() {
					out = append(out, id)
				}
			}
		}
	}
	return out
}

func withoutSelf(self string, in []string) (out []string) {
	for _, id := range in {
		if id != self {
			out = append(out, id)
		}
	}
	return out
}

func highestEdge(g *graph.Graph, id string) string {
	edges := graph.Sources(g.UpEdges(id))
	for _, edge := range edges {
		if !graph.IsRoot(edge) && edge != id {
			return highestEdge(g, edge)
		}
	}
	return id
}

type byDependencyCount struct {
	g     *graph.Graph
	nodes []*node.Node
}

func (b byDependencyCount) Len() int      { return len(b.nodes) }
func (b byDependencyCount) Swap(i, j int) { b.nodes[i], b.nodes[j] = b.nodes[j], b.nodes[i] }
func (b byDependencyCount) Less(i, j int) bool {
	return len(b.g.Dependencies(b.nodes[i].ID)) > len(b.g.Dependencies(b.nodes[j].ID))
}

func groupDeps(ctx context.Context, g *graph.Graph, group string) (*graph.Graph, error) {
	logger := logging.GetLogger(ctx).WithField("function", "groupDeps")

	nodes := g.GroupNodes(group)
	sort.Sort(byDependencyCount{g, nodes})

	for _, meta := range nodes {
		l := logger.WithField("id", meta.ID)
		// align all up edges in a single branch
		g, err := alignEdgesInGroup(ctx, g, meta.ID, group)
		if err != nil {
			l.Error(err)
			return g, errors.Wrap(err, "failed to align edges in branch")
		}
	}

	g, err := connectIsolatedGroupNodes(ctx, g, nodes, group)
	if err != nil {
		logger.Error(err)
		return g, errors.Wrap(err, "failed to connect group nodes")
	}

	g, err = connectIsolatedGroupBranches(ctx, g, nodes)
	if err != nil {
		logger.Error(err)
		return g, errors.Wrap(err, "failed to connect group branches")
	}

	return g, nil
}

func alignEdgesInGroup(ctx context.Context, g *graph.Graph, id, group string) (*graph.Graph, error) {
	upEdges := withoutRoot(graph.Sources(g.UpEdges(id)))
	for i, upEdge := range upEdges {
		if i > 0 {
			dest := highestEdge(g, upEdges[i-1])
			if err := g.SafeDisconnect(upEdge, id); err != nil {
				return g, err
			}

			if !willCycle(g, upEdge, dest) {
				if err := g.SafeConnect(upEdge, dest); err != nil {
					return g, err
				}
			}
		}

		// if the node has more than one down edge we want to keep the one with
		// the most dependencies
		downEdges := g.DownEdgesInGroup(upEdge, group)
		if len(downEdges) > 1 {
			var downNodes []*node.Node
			for _, downEdge := range downEdges {
				if meta, ok := g.Get(downEdge); ok {
					downNodes = append(downNodes, meta)
				}
			}
			sort.Sort(byDependencyCount{g, downNodes})
			for i := 1; i < len(downNodes); i++ {
				if err := g.SafeDisconnect(upEdge, downNodes[i].ID); err != nil {
					return g, err
				}
			}
		}
	}
	return g, nil
}

func connectIsolatedGroupNodes(ctx context.Context, g *graph.Graph, nodes []*node.Node, group string) (*graph.Graph, error) {
	// collect remaining group nodes that have no edges
	var unconnected []*node.Node
	for _, meta := range nodes {
		downEdges := withoutSelf(meta.ID, g.DownEdgesInGroup(meta.ID, group))
		upEdges := withoutModule(g, withoutRoot(graph.Sources(g.UpEdges(meta.ID))))
		if len(downEdges) == 0 && len(upEdges) == 0 {
			unconnected = append(unconnected, meta)
		}
	}

	groupDep := func(id string) string {
		pid := graph.ParentID(id)
		if !graph.IsRoot(pid) {
			id = pid
		}
		return id
	}

	// connect unconnected in single branch
	for i, meta := range unconnected {
		if i > 0 {
			from := meta.ID
			to := unconnected[i-1].ID
			if !graph.AreSiblingIDs(from, to) {
				from = groupDep(from)
				to = groupDep(to)
			}
			if err := g.SafeConnect(from, to); err != nil {
				return g, err
			}
		}
	}
	return g, nil
}

func connectIsolatedGroupBranches(ctx context.Context, g *graph.Graph, nodes []*node.Node) (*graph.Graph, error) {
	// collect all unconnected group branches
	var groupBranches []*node.Node
	for _, meta := range nodes {
		downEdges := graph.Targets(g.DownEdges(meta.ID))
		if len(downEdges) == 0 {
			groupBranches = append(groupBranches, meta)
		}
	}

	// connect branches into a single branch
	for i, treeRoot := range groupBranches {
		if i > 0 {
			from := treeRoot.ID
			dest := highestEdge(g, groupBranches[i-1].ID)

			if !willCycle(g, from, dest) {
				if err := g.SafeConnect(from, dest); err != nil {
					return g, err
				}
			}
		}
	}
	return g, nil
}

func willCycle(g *graph.Graph, from, to string) bool {
	var willCycle bool
	for _, dep := range g.Dependencies(to) {
		if dep == from {
			willCycle = true
			break
		}
	}
	return willCycle
}
