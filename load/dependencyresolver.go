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
	"github.com/asteris-llc/converge/helpers/logging"
	"github.com/asteris-llc/converge/parse"
	"github.com/asteris-llc/converge/render/extensions"
	"github.com/asteris-llc/converge/render/preprocessor"
)

type dependencyGenerator func(node *parse.Node) ([]string, error)

// ResolveDependencies examines the strings and depdendencies at each vertex of
// the graph and creates edges to fit them
func ResolveDependencies(ctx context.Context, g *graph.Graph) (*graph.Graph, error) {
	logger := logging.GetLogger(ctx).WithField("function", "ResolveDependencies")
	logger.Debug("resolving dependencies")

	return g.Transform(ctx, func(id string, out *graph.Graph) error {
		if id == "root" { // skip root
			return nil
		}

		node, ok := out.Get(id).(*parse.Node)
		if !ok {
			return fmt.Errorf("ResolveDependencies can only be used on Graphs of *parse.Node. I got %T", out.Get(id))
		}

		depGenerators := []dependencyGenerator{
			getDepends,
			func(node *parse.Node) ([]string, error) {
				return getParams(g, id, node)
			},
			func(node *parse.Node) ([]string, error) {
				return getXrefs(g, id, node)
			},
		}

		// we have dependencies from various sources, but they're always IDs, so we
		// can connect them pretty easily
		for _, source := range depGenerators {
			deps, err := source(node)
			if err != nil {
				return err
			}
			for _, dep := range deps {
				out.Connect(id, dep)
			}
		}
		return nil
	})
}

func getDepends(node *parse.Node) ([]string, error) {
	deps, err := node.GetStringSlice("depends")

	switch err {
	case parse.ErrNotFound:
		return []string{}, nil

	case nil:
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
	language.On("param", extensions.RememberCalls(&out, 0))
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
	if node == "root" || node == "" {
		return "", false
	}
	siblingID := graph.SiblingID(id, node)
	val := g.Get(siblingID)
	if val == nil {
		return getNearestAncestor(g, graph.ParentID(id), node)
	}
	elem, ok := val.(*parse.Node)
	if !ok {
		return "", false
	}
	if elem.Kind() == "module" {
		return "", false
	}
	return siblingID, true
}
