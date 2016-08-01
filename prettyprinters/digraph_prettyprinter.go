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

import "github.com/asteris-llc/converge/graph"

// A DigraphPrettyPrinter is a concrete implementation of some output format
// that can be used to prettyprint a Directed Graph.
type DigraphPrettyPrinter interface {

	// StartPP should take a graph and shall return a string used to start the
	// graph output, or an error which will be returned to the user.
	StartPP(graph *graph.Graph) (Renderable, error)

	// FinishPP will be given a graph and shall return a string used to finalize
	// graph output, or an error which will be returned to the user.
	FinishPP(graph *graph.Graph) (Renderable, error)

	// MarkNode() is a function used to identify the boundaries of subgraphs
	// within the larger graph.  MarkNode() is called with a graph and the id of a
	// node within the graph, and should return a *SubgraphMarker with the
	// SubgraphID and Start fields set to indicate the beginning or end of a
	// subgraph.  If nil is returned the node will be included in the subgraph of
	// it's parent (or ⊥ if it is not part of any subgraph).
	// The string "⊥" is reserved for the bottom subgraph element and shouldn't be
	// used as a NodeID.
	MarkNode(graph *graph.Graph, nodeID string) *SubgraphMarker

	// StartSubgraph will be given a graph, the ID of the root node of the
	// subgraph, and the SubgraphID returned as part of MarkNode().  It should
	// return the string used to start the subgraph or an error that will be
	// returned to the user.
	StartSubgraph(graph *graph.Graph, nodeID string, subgraphID SubgraphID) (Renderable, error)

	// FinishSubgraph will be given a graph and a SubgraphID returned by MarkNode
	// and should return a string used to end a subgraph, or an error that will be
	// returned to the user.
	FinishSubgraph(graph *graph.Graph, subgraphID SubgraphID) (Renderable, error)

	// StartNodeSection will be given a graph and should return a string used to
	// start the node section, or an error that will be returned to the user.
	StartNodeSection(graph *graph.Graph) (Renderable, error)

	// FinishNodeSection will be given a graph and should return a string used to
	// finish the node section in the final output, or an error that will be
	// returned to the user.
	FinishNodeSection(graph *graph.Graph) (Renderable, error)

	// DrawNode will be called once for each node in the graph.  The function will
	// be given a graph and a string ID for the current node in the graph, and
	// should return a string representing the node in the final output, or an
	// error that will be returned to the user.
	DrawNode(graph *graph.Graph, nodeID string) (Renderable, error)

	// StartEdgeSection will be given a graph and should return a string used to
	// start the edge section, or an error that will be returned to the user.
	StartEdgeSection(graph *graph.Graph) (Renderable, error)

	// FinishEdgeSection will be given a graph and should return a string used to
	// finish the edge section in the final output, or an error that will be
	// returned to the user.
	FinishEdgeSection(graph *graph.Graph) (Renderable, error)

	// DrawEdge will be called once for each edge in the graph.  It is called with
	// a graph, the ID of the source vertex, and the ID of the target vertex.  It
	// should return a string representing the edge in the final output, or an
	// error that will be returned to the user.
	DrawEdge(graph *graph.Graph, srcNodeID string, dstNodeID string) (Renderable, error)
}
