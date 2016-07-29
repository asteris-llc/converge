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

package prettyprinters

import (
	"bytes"
	"errors"

	"golang.org/x/net/context"

	"github.com/asteris-llc/converge/graph"
)

// Subgraph are treated as a semi-lattice with the root graph as the ⊥ (join)
// element.  Subgraphs are partially ordered based on the rank of their root
// element in the graph.

// Subgraph is a type that represents a subgraph containing it's ID, it's root
// node, a list of it's terminal nodes, and any other included nodes.
type Subgraph struct {
	StartNode *string
	EndNodes  []string
	ID        SubgraphID
	Nodes     []string
}

// SubgraphMap is a simple type alias for keeping track of SubgraphID ->
// Subgraph mappings
type SubgraphMap map[SubgraphID]Subgraph

// A SubgraphID is an opaque type that is handled by the DigraphPrettyPrinter.
// SubgraphID must adhere to the following conditions:
//
//   * Unique per graph
//   * Comparable
type SubgraphID interface{}

// A SubgraphMarker is a structure used to identify graph nodes that mark the
// beginning or end of a Subgraph.  SubgraphMakers are returned by MarkNode().
// The SubgraphID field should be set to the name of a unique SubgraphID if
// start is true.  When Start is false the SubgraphID field is ignored.
type SubgraphMarker struct {
	SubgraphID SubgraphID // The ID for this subgraph (if Start is true)
	Start      bool       // True if this is the start of a new subgraph
}

// SubgraphBottomID defines the SubgraphID for the bottom subgraph.  "⊥" is a
// reserved SubgraphID and shouldn't be returned as the SubgraphID by any calls
// to MakeNode.
const SubgraphBottomID string = "⊥"

var (
	// This is the join element that contains all other subgraphs.  subgraphBottom
	// is the parent of all top-level subgraphs.
	subgraphBottom = Subgraph{
		StartNode: nil,
		Nodes:     make([]string, 0),
	}
)

// makeSubgraphMap creates a new map to hold subgraph data and includes a ⊥
// element.  Returns a tuple of the subgraph map and the identifier for the join
// element.
func makeSubgraphMap() SubgraphMap {
	subgraph := make(SubgraphMap)
	subgraph[SubgraphBottomID] = subgraphBottom
	return subgraph
}

// loadSubgraphs takes a graph and a subgraph map and traverses the graph,
// calling MarkNode() on each node, creating and updating subgraphs as
// necessary.
func (p Printer) loadSubgraphs(ctx context.Context, g *graph.Graph, subgraphs SubgraphMap) {
	g.RootFirstWalk(ctx, func(id string, val interface{}) error {
		if sgMarker := p.pp.MarkNode(g, id); sgMarker != nil {
			if sgMarker.Start {
				addNodeToSubgraph(subgraphs, sgMarker.SubgraphID, id)
			} else {
				thisSubgraph := getSubgraphByID(subgraphs, graph.ParentID(id))
				setSubgraphEndNode(subgraphs, thisSubgraph, id)
			}
		} else {
			subgraphID := getSubgraphByID(subgraphs, graph.ParentID(id))
			addNodeToSubgraph(subgraphs, subgraphID, id)
		}
		return nil
	})
}

// getParentSubgraph will, given a SubgraphMap a SubgraphID, and a position in
// the graph, locate the parent subgraph, or return ⊥ if no other subgraph is
// found.
func getParentSubgraph(subgraphs SubgraphMap, thisSubgraph SubgraphID, id string) SubgraphID {
	parent := graph.ParentID(id)
	if thisSubgraph == SubgraphBottomID {
		return SubgraphBottomID
	}
	if id == "." || parent == "." {
		return SubgraphBottomID
	}
	for subgraphID, subgraph := range subgraphs {
		for nodeIdx := range subgraph.Nodes {
			if id == subgraph.Nodes[nodeIdx] {
				if subgraphID == thisSubgraph {
					return getParentSubgraph(subgraphs, thisSubgraph, graph.ParentID(id))
				}
				return subgraphID
			}
		}
	}
	return getParentSubgraph(subgraphs, thisSubgraph, graph.ParentID(id))
}

// setSubgraphEndNode appends the given node to the nodes and end nodes lists
// for the specified subgraph.
func setSubgraphEndNode(subgraphs SubgraphMap, subgraphID SubgraphID, node string) {
	if subgraphID == SubgraphBottomID {
		return
	}

	if sg, found := subgraphs[subgraphID]; found {
		sg.EndNodes = append(sg.EndNodes, node)
		sg.Nodes = append(sg.Nodes, node)
		subgraphs[subgraphID] = sg
	}
}

// isSubgraphEnd returns true if the subgraph exists and the node string is a
// member it's EndNodes list, and false otherwise.
func isSubgraphEnd(subgraphs SubgraphMap, id SubgraphID, node string) bool {
	sg, found := subgraphs[id]
	if !found {
		return false
	}
	for endNodeIdx := range sg.EndNodes {
		if node == sg.EndNodes[endNodeIdx] {
			return true
		}
	}
	return false
}

// getSubgraphByID returns the first subgraph that has the given node as an
// element, and ⊥ otherwise.
func getSubgraphByID(subgraphs SubgraphMap, id string) SubgraphID {
	if id == "." {
		return SubgraphBottomID
	}
	for subgraphID, subgraph := range subgraphs {
		if isSubgraphEnd(subgraphs, subgraphID, id) {
			return getParentSubgraph(subgraphs, subgraphID, id)
		}
		for nodeIdx := range subgraph.Nodes {
			if id == subgraph.Nodes[nodeIdx] {
				return subgraphID
			}
		}
	}
	return getSubgraphByID(subgraphs, graph.ParentID(id))
}

// addNodeToSubgraph appnds the given node to a subgraph, creating a new
// subgraph if necessary.
func addNodeToSubgraph(subgraphs SubgraphMap, subgraphID SubgraphID, vertexID string) {
	oldGraph, found := subgraphs[subgraphID]
	if !found {
		oldGraph.StartNode = &vertexID
		oldGraph.ID = subgraphID
	}
	oldGraph.StartNode = &vertexID
	oldGraph.Nodes = append(oldGraph.Nodes, vertexID)
	subgraphs[subgraphID] = oldGraph
}

// drawSubgraph call DrawNode for each element within the specified subgraph,
// calling StartSubgraph() and FinishSubgraph before and after to ensure that
// the returned string contains any prefix/postfix additions defined by the
// printer.
func (p Printer) drawSubgraph(g *graph.Graph, id SubgraphID, subgraph Subgraph) (*StringRenderable, error) {
	var buffer bytes.Buffer
	subgraphNodes := subgraph.Nodes
	if nil == subgraph.StartNode {
		return nil, errors.New("Cannot draw subgraph starting at nil vertex")
	}
	if str, err := p.pp.StartSubgraph(g, *subgraph.StartNode, subgraph.ID); err == nil {
		writeRenderable(buffer, str)
	} else {
		return nil, err
	}

	for idx := range subgraphNodes {
		if str, err := p.pp.DrawNode(g, subgraphNodes[idx]); err == nil {
			writeRenderable(buffer, str)
		} else {
			return nil, err
		}
	}

	if str, err := p.pp.FinishSubgraph(g, id); err == nil {
		writeRenderable(buffer, str)
	}

	return VisibleString(buffer.String()), nil
}
