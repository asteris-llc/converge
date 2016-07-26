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

// Package prettyprinters provides a general interface and concrete
// implementations for implementing prettyprinters.  This package was originally
// created to facilitate the development of graphviz visualizations for resource
// graphs, however it is intended to be useful for creating arbitrary output
// generators so that resource graph data can be used in other applications.
//
// See the 'examples' directory for examples of using the prettyprinter, and see
// the 'graphviz' package for an example of a concrete implementation of
// DigraphPrettyPrinter.
package prettyprinters

import (
	"bytes"
	"errors"

	"github.com/asteris-llc/converge/graph"
)

type SubgraphMap map[SubgraphID]Subgraph

// Subgraphs are treated as a semi-lattice with the root graph as the ⊥ (join)
// element.  Subgraphs are partially ordered based on the rank of their root
// element in the graph.
type SubgraphID interface{}

type SubgraphMarker struct {
	SubgraphID SubgraphID
	Start      bool
}

type Subgraph struct {
	StartNode *string
	EndNodes  []string
	ID        SubgraphID
	Nodes     []string
}

var (
	subgraphBottom Subgraph = Subgraph{
		StartNode: nil,
		Nodes:     make([]string, 0),
	}
	subgraphBottomID SubgraphID = "⊥"
)

// makeSubgraphMap creates a new map to hold subgraph data and includes a ⊥
// element.  Returns a tuple of the subgraph map and the identifier for the join
// element.
func makeSubgraphMap() (SubgraphMap, SubgraphID) {
	subgraph := make(SubgraphMap)
	subgraph[subgraphBottomID] = subgraphBottom
	return subgraph, subgraphBottomID
}

// New() creates a new prettyprinter instance from the specified
// DigraphPrettyPrinter instance.

func (p Printer) loadSubgraphs(g *graph.Graph, subgraphs SubgraphMap) {
	g.RootFirstWalk(func(id string, val interface{}) error {
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

func getParentSubgraph(subgraphs SubgraphMap, thisSubgraph SubgraphID, id string) SubgraphID {
	parent := graph.ParentID(id)
	if thisSubgraph == subgraphBottomID {
		return subgraphBottomID
	}
	if id == "." || parent == "." {
		return subgraphBottomID
	}
	for subgraphID, subgraph := range subgraphs {
		for nodeIdx := range subgraph.Nodes {
			if id == subgraph.Nodes[nodeIdx] {
				if subgraphID == thisSubgraph {
					return getParentSubgraph(subgraphs, thisSubgraph, graph.ParentID(id))
				} else {
					return subgraphID
				}
			}
		}
	}
	return getParentSubgraph(subgraphs, thisSubgraph, graph.ParentID(id))
}

func setSubgraphEndNode(subgraphs SubgraphMap, subgraphID SubgraphID, node string) {
	if subgraphID == subgraphBottomID {
		return
	}

	sg := subgraphs[subgraphID]
	sg.EndNodes = append(sg.EndNodes, node)
	sg.Nodes = append(sg.Nodes, node)
	subgraphs[subgraphID] = sg
}

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

func getSubgraphByID(subgraphs SubgraphMap, id string) SubgraphID {
	if id == "." {
		return subgraphBottomID
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

func (p Printer) drawSubgraph(g *graph.Graph, id SubgraphID, subgraph Subgraph) (string, error) {
	var buffer bytes.Buffer
	subgraphNodes := subgraph.Nodes
	if nil == subgraph.StartNode {
		return "", errors.New("Cannot draw subgraph starting at nil vertex")
	}
	if str, err := p.pp.StartSubgraph(g, *subgraph.StartNode, subgraph.ID); err == nil {
		buffer.WriteString(str)
	} else {
		return "", err
	}

	for idx := range subgraphNodes {
		if str, err := p.pp.DrawNode(g, subgraphNodes[idx]); err == nil {
			buffer.WriteString(str)
		} else {
			return "", err
		}
	}

	if str, err := p.pp.FinishSubgraph(g, id); err == nil {
		buffer.WriteString(str)
	}

	return buffer.String(), nil
}
