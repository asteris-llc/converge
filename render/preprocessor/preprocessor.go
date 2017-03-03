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

package preprocessor

import (
	"fmt"
	"os"
	"strings"

	"github.com/asteris-llc/converge/parse"

	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/resource/module"
)

// VertexSplit takes a graph with a set of vertexes and a string, and returns
// the longest vertex id from the graph and the remainder of the string.  If no
// matching vertex is found 'false' is returned.
func VertexSplit(g *graph.Graph, s string) (string, string, bool) {
	prefix, found := Find(Prefixes(s), g.Contains)
	if !found {
		return "", s, false
	}
	if prefix == s {
		return prefix, "", true
	}
	return prefix, s[len(prefix)+1:], true
}

// VertexSplitTraverse will act like vertex split, looking for a prefix matching
// the current set of graph nodes, however unlike `VertexSplit`, if a node is
// not found at the current level it will look at the parent level to the
// provided starting node, unless stop(parent) returns true.
func VertexSplitTraverse(g *graph.Graph, toFind string, startingNode string, stop func(*graph.Graph, string) bool, history map[string]struct{}) (string, string, bool) {
	history[startingNode] = struct{}{}

	for _, child := range g.Children(startingNode) {
		if _, ok := history[child]; ok {
			continue
		}
		if stop(g, child) {
			continue
		}
		vertex, middle, found := VertexSplitTraverse(g, toFind, child, stop, history)
		if found {
			return vertex, middle, found
		}
	}
	if stop(g, startingNode) {
		return "", toFind, false
	}

	fqgn := graph.SiblingID(startingNode, toFind)
	vertex, middle, found := VertexSplit(g, fqgn)
	if found {
		return vertex, middle, found
	}
	parentID := graph.ParentID(startingNode)
	return VertexSplitTraverse(g, toFind, parentID, stop, history)
}

// TraverseUntilModule is a function intended to be used with
// VertexSplitTraverse and will cause vertex splitting to propogate upwards
// until it encounters a module
func TraverseUntilModule(g *graph.Graph, id string) bool {
	if graph.IsRoot(id) {
		fmt.Fprintf(os.Stderr, "TraverseUnitModule: encountered root, aborting\n") // DEBUG
		return true
	}
	elemMeta, ok := g.Get(id)
	if !ok {
		fmt.Fprintf(os.Stderr, "TraverseUntilModule: %s isn't in graph, aborting\n", id) // DEBUG
		return true
	}
	elem := elemMeta.Value()
	if _, ok := elem.(*module.Module); ok {
		return true
	}
	if _, ok := elem.(*module.Preparer); ok {
		return true
	}
	if node, ok := elem.(*parse.Node); ok {
		return node.Kind() == "module"
	}
	return false
}

// Find returns the first element of the string slice for which f returns true
func Find(slice []string, f func(string) bool) (string, bool) {
	for _, elem := range slice {
		if f(elem) {
			return elem, true
		}
	}
	return "", false
}

// SplitTerms takes a string and splits it on '.'
func SplitTerms(in string) []string {
	return strings.Split(in, ".")
}

// JoinTerms takes a list of terms and joins them with '.'
func JoinTerms(s []string) string {
	return strings.Join(s, ".")
}

// Inits returns a list of heads of the string,
// e.g. [1,2,3] -> [[1,2,3],[1,2],[1]]
func Inits(in []string) [][]string {
	var results [][]string
	for i := 0; i < len(in); i++ {
		results = append([][]string{in[0 : i+1]}, results...)
	}
	return results
}

// Prefixes returns a set of prefixes for a string, e.g. "a.b.c.d" will yield
// []string{"a.b.c.d","a.b.c","a.b.","a"}
func Prefixes(in string) (out []string) {
	for _, termSet := range Inits(SplitTerms(in)) {
		out = append(out, JoinTerms(termSet))
	}
	return out
}
