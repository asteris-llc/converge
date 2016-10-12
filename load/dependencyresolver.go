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
	"text/template"

	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/graph/node"
	"github.com/asteris-llc/converge/helpers/logging"
	"github.com/asteris-llc/converge/parse"
	"github.com/asteris-llc/converge/parse/preprocessor/lock"
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

	return g.Transform(ctx, func(meta *node.Node, out *graph.Graph) error {
		if meta.ID == "root" { // skip root
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
				out.Connect(meta.ID, dep)
			}
		}

		// if the node has a named lock, connect it between the lock entry and exit
		// nodes
		lockName, err := getLockName(node)
		if err != nil {
			return err
		}
		if lockName != "" {
			lockNodeID := graph.SiblingID(meta.ID, lock.NewLockID(lockName))
			unlockNodeID := graph.SiblingID(meta.ID, lock.NewUnlockID(lockName))
			out.Connect(meta.ID, lockNodeID)
			out.Connect(unlockNodeID, meta.ID)
		}

		return nil
	})
}

// ResolveDependenciesInLocks ensures that all nodes in a lock depend on each
// other
func ResolveDependenciesInLocks(ctx context.Context, g *graph.Graph) (*graph.Graph, error) {
	logger := logging.GetLogger(ctx).WithField("function", "ResolveDependenciesInLocks")
	logger.Debug("resolving dependencies in locks")

	return g.Transform(ctx, func(meta *node.Node, out *graph.Graph) error {
		if meta.ID == "root" { // skip root
			return nil
		}

		node, ok := meta.Value().(*parse.Node)
		if !ok {
			return fmt.Errorf("ResolveDependenciesInLocks can only be used on Graphs of *parse.Node. I got %T", meta.Value())
		}

		if lock.IsUnlockNode(node) {
			var lockDeps []string
			// collect all of the dependencies with locks on them
			for _, dep := range out.Dependencies(meta.ID) {
				depnode, ok := out.Get(dep)
				if !ok {
					continue
				}

				parsedDepNode, ok := depnode.Value().(*parse.Node)
				if !ok {
					return fmt.Errorf("dependency node must be of type *parse.Node. I got %T", depnode.Value())
				}

				hasLock, err := lock.HasLock(parsedDepNode)
				if err != nil {
					return errors.Wrapf(err, "failed to check for lock on %s", dep)
				}
				if hasLock {
					lockDeps = append(lockDeps, dep)
				}
			}

			// ensure each dependency within the lock depends on another or on the lock
			// entry/exit
			if len(lockDeps) > 0 {
				lockName := lock.GetLockName(graph.BaseID(meta.ID))
				lockNodeID := graph.SiblingID(meta.ID, lock.NewLockID(lockName))

				var lastDep string
				lastDepIdx := len(lockDeps) - 1
				for i, dep := range lockDeps {
					if i > 0 {
						out.Disconnect(meta.ID, dep)
						out.Connect(lastDep, dep)
					}
					if i != lastDepIdx {
						out.Disconnect(dep, lockNodeID)
					}
					lastDep = dep
				}
			}
		}
		return nil
	})
}

func getLockName(node *parse.Node) (string, error) {
	lock, err := node.GetString("lock")
	switch err {
	case parse.ErrNotFound:
		return "", nil

	case nil:
		return lock, nil

	default:
		return "", err
	}
}

func getDepends(_ *graph.Graph, id string, node *parse.Node) ([]string, error) {
	deps, err := node.GetStringSlice("depends")
	switch err {
	case parse.ErrNotFound:
		return []string{}, nil
	case nil:
		for idx := range deps {
			deps[idx] = graph.SiblingID(id, deps[idx])
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
	if dst == "." || dst == "root" {
		return "", false
	}
	if graph.AreSiblingIDs(src, dst) {
		return dst, true
	}
	return getPeerVertex(src, graph.ParentID(dst))
}

func getNearestAncestor(g *graph.Graph, id, node string) (string, bool) {
	if id == "root" || id == "" || id == "." {
		return "", false
	}

	siblingID := graph.SiblingID(id, node)

	valMeta, ok := g.Get(siblingID)
	if !ok {
		return getNearestAncestor(g, graph.ParentID(id), node)
	}
	elem, ok := valMeta.Value().(*parse.Node)
	if !ok {
		return "", false
	}
	if elem.Kind() == "module" {
		return "", false
	}
	return siblingID, true
}
