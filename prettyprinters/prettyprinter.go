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
package prettyprinters

import (
	"bytes"
	"errors"

	"github.com/asteris-llc/converge/graph"
)

// DigraphPrettyPrinter interface defines the minimal set of required functions
// for defining a pretty printer.
type DigraphPrettyPrinter interface {

	// StartPP should take a graph and shall return a string used to start the
	// graph output, or an error which will be returned to the user.
	StartPP(*graph.Graph) (string, error)

	// FinishPP will be given a graph and shall return a string used to finalize
	// graph output, or an error which will be returned to the user.
	FinishPP(*graph.Graph) (string, error)

	// MarkNode() is a function used to identify the boundaries of subgraphs
	// within the larger graph.  MarkNode() is called with a graph and the id of a
	// node within the graph, and should return a *SubgraphMarker with the
	// SubgraphID and Start fields set to indicate the beginning or end of a
	// subgraph.  If nil is returned the node will be included in the subgraph of
	// it's parent (or ⊥ if it is not part of any subgraph)
	MarkNode(*graph.Graph, string) *SubgraphMarker

	// StartSubgraph will be given a graph, the ID of the root node of the
	// subgraph, and the SubgraphID returned as part of MarkNode().  It should
	// return the string used to start the subgraph or an error that will be
	// returned to the user.
	StartSubgraph(*graph.Graph, string, SubgraphID) (string, error)

	// FinishSubgraph will be given a graph and a SubgraphID returned by MarkNode
	// and should return a string used to end a subgraph, or an error that will be
	// returned to the user.
	FinishSubgraph(*graph.Graph, SubgraphID) (string, error)

	// StartNodeSection will be given a graph and should return a string used to
	// start the node section, or an error that will be returned to the user.
	StartNodeSection(*graph.Graph) (string, error)

	// FinishNodeSection will be given a graph and should return a string used to
	// finish the node section in the final output, or an error that will be
	// returned to the user.
	FinishNodeSection(*graph.Graph) (string, error)

	// DrawNode will be called once for each node in the graph.  The function will
	// be given a graph and a string ID for the current node in the graph, and
	// should return a string representing the node in the final output, or an
	// error that will be returned to the user.
	DrawNode(*graph.Graph, string) (string, error)

	// StartEdgeSection will be given a graph and should return a string used to
	// start the edge section, or an error that will be returned to the user.
	StartEdgeSection(*graph.Graph) (string, error)

	// FinishEdgeSection will be given a graph and should return a string used to
	// finish the edge section in the final output, or an error that will be
	// returned to the user.
	FinishEdgeSection(*graph.Graph) (string, error)

	// DrawEdge will be called once for each edge in the graph.  It is called with
	// a graph, the ID of the source vertex, and the ID of the target vertex.  It
	// should return a string representing the edge in the final output, or an
	// error that will be returned to the user.
	DrawEdge(*graph.Graph, string, string) (string, error)
}

type Printer struct {
	pp DigraphPrettyPrinter
}

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

func New(p DigraphPrettyPrinter) Printer {
	return Printer{pp: p}
}

func (p Printer) Show(g *graph.Graph) (string, error) {
	var outputBuffer bytes.Buffer
	edges := g.Edges()
	subgraphs, subgraphJoinID := makeSubgraphMap()
	p.loadSubgraphs(g, subgraphs)
	rootNodes := subgraphs[subgraphJoinID].Nodes
	if str, err := p.pp.StartPP(g); err == nil {
		outputBuffer.WriteString(str)
	} else {
		return "", err
	}
	for idx := range rootNodes {
		if str, err := p.pp.DrawNode(g, rootNodes[idx]); err == nil {
			outputBuffer.WriteString(str)
		} else {
			return "", err
		}
	}

	for graphID, graph := range subgraphs {
		if graphID == subgraphJoinID {
			continue
		}

		if str, err := p.drawSubgraph(g, graphID, graph); err == nil {
			outputBuffer.WriteString(str)
		} else {
			return "", err
		}
	}

	if str, err := p.pp.StartEdgeSection(g); err == nil {
		outputBuffer.WriteString(str)
	} else {
		return "", err
	}

	for idx := range edges {
		src := edges[idx].Source
		dst := edges[idx].Dest
		if str, err := p.pp.DrawEdge(g, src, dst); err == nil {
			outputBuffer.WriteString(str)
		} else {
			return "", err
		}
	}

	if str, err := p.pp.StartEdgeSection(g); err == nil {
		outputBuffer.WriteString(str)
	} else {
		return "", err
	}

	if str, err := p.pp.FinishPP(g); err == nil {
		outputBuffer.WriteString(str)
	} else {
		return "", err
	}

	return outputBuffer.String(), nil
}

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
