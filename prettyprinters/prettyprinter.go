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

//Package prettyprinters provides a general interface and concrete
//implementations for implementing prettyprinters.  This package was originally
//created to facilitate the development of graphviz visualizations for resource
//graphs, however it is intended to be useful for creating arbitrary output
//generators so that resource graph data can be used in other applications.
package prettyprinters

import (
	"bytes"
	"errors"

	"github.com/asteris-llc/converge/graph"
)

//DigraphPrettyPrinter interface defines the minimal set of required functions
//for defining a pretty printer.
type DigraphPrettyPrinter interface {

	//StartPP will be given as it's argument a pointer to the root node of the
	//graph structure.  It should do any necessary work to create the beginning of
	//the document and do any first-pass walks of the graph that may be necessary
	//for rendering output.
	StartPP(*graph.Graph) (string, error)

	//FinishPP will be given as it's argument a pointer to the root node of the
	//graph structure.  It should do any necessary work to finish the generation
	//of the prettyprinted output.
	FinishPP(*graph.Graph) (string, error)

	//When DrawNode() calls the provided node mark function,
	StartSubgraph(*graph.Graph, string, SubgraphID) (string, error)
	FinishSubgraph(*graph.Graph, SubgraphID) (string, error)

	StartNodeSection(*graph.Graph) (string, error)
	FinishNodeSection(*graph.Graph) (string, error)
	DrawNode(*graph.Graph, string) (string, error)

	MarkNode(*graph.Graph, string) *SubgraphMarker

	StartEdgeSection(*graph.Graph) (string, error)
	FinishEdgeSection(*graph.Graph) (string, error)
	DrawEdge(*graph.Graph, string, string) (string, error)
}

type Printer struct {
	pp DigraphPrettyPrinter
}

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
	subgraphBottomID SubgraphID = "bottom"
)

// makeSubgraphMap creates a new map to hold subgraph data and includes a ⊥
// element.  Returns a tuple of the subgraph map and the identifier for the join
// element.
func makeSubgraphMap() (map[SubgraphID]Subgraph, SubgraphID) {
	subgraph := make(map[SubgraphID]Subgraph)
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
	p.loadSubgraphs(g, subgraphBottomID, subgraphs)
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
		if str, err := p.pp.DrawEdge(g, edges[idx].Source, edges[idx].Dest); err == nil {
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

func (p Printer) loadSubgraphs(g *graph.Graph, bottom SubgraphID, subgraphs map[SubgraphID]Subgraph) {
	g.RootFirstWalk(func(id string, val interface{}) error {
		if sgMarker := p.pp.MarkNode(g, id); sgMarker != nil {
			if sgMarker.Start {
				addNodeToSubgraph(subgraphs, sgMarker.SubgraphID, id)
			} else {
				thisSubgraph := getSubgraphByID(g, subgraphs, graph.ParentID(id))
				setSubgraphEndNode(subgraphs, thisSubgraph, id)
			}
		} else {
			defer addNodeToSubgraph(subgraphs, getSubgraphByID(g, subgraphs, graph.ParentID(id)), id)
		}
		return nil
	})
}

func getParentSubgraph(g *graph.Graph, subgraphs map[SubgraphID]Subgraph, thisSubgraph SubgraphID, id string) SubgraphID {
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
					return getParentSubgraph(g, subgraphs, thisSubgraph, graph.ParentID(id))
				} else {
					return subgraphID
				}
			}
		}
	}
	return getParentSubgraph(g, subgraphs, thisSubgraph, graph.ParentID(id))
}

func setSubgraphEndNode(subgraphs map[SubgraphID]Subgraph, subgraphID SubgraphID, node string) {
	if subgraphID == subgraphBottomID {
		return
	}

	sg := subgraphs[subgraphID]
	sg.EndNodes = append(sg.EndNodes, node)
	sg.Nodes = append(sg.Nodes, node)
	subgraphs[subgraphID] = sg
}

func isSubgraphEnd(subgraphs map[SubgraphID]Subgraph, id SubgraphID, node string) bool {
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

func getSubgraphByID(g *graph.Graph, subgraphs map[SubgraphID]Subgraph, id string) SubgraphID {
	if id == "." {
		return subgraphBottomID
	}
	for subgraphID, subgraph := range subgraphs {
		if isSubgraphEnd(subgraphs, subgraphID, id) {
			return getParentSubgraph(g, subgraphs, subgraphID, id)
		}
		for nodeIdx := range subgraph.Nodes {
			if id == subgraph.Nodes[nodeIdx] {
				return subgraphID
			}
		}
	}
	return getSubgraphByID(g, subgraphs, graph.ParentID(id))
}

func addNodeToSubgraph(subgraphs map[SubgraphID]Subgraph, subgraphID SubgraphID, vertexID string) {
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
