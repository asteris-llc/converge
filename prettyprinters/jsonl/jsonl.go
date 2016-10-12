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

package jsonl

import (
	"encoding/json"

	"github.com/asteris-llc/converge/graph"
	pp "github.com/asteris-llc/converge/prettyprinters"
)

// Node is the serializable type for graph nodes
type Node struct {
	Kind  string      `json:"kind"`
	ID    string      `json:"id"`
	Value interface{} `json:"value"`
}

// Edge is the serializable type for graph edges
type Edge struct {
	Kind        string `json:"kind"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
}

// Printer prints a graph in JSONL format
type Printer struct{}

// DrawNode prints a node in JSONL format
func (j *Printer) DrawNode(graph *graph.Graph, nodeID string) (pp.Renderable, error) {
	meta, ok := graph.Get(nodeID)
	if !ok {
		return pp.HiddenString(), nil
	}

	// TODO: should this use meta instead of the value? Should we expose that?
	out, err := json.Marshal(&Node{Kind: "node", ID: nodeID, Value: meta.Value()})
	return pp.VisibleString(string(out) + "\n"), err
}

// DrawEdge returns an edge in JSONL format
func (j *Printer) DrawEdge(graph *graph.Graph, srcNodeID string, dstNodeID string) (pp.Renderable, error) {
	out, err := json.Marshal(&Edge{Kind: "edge", Source: srcNodeID, Destination: dstNodeID})
	return pp.VisibleString(string(out) + "\n"), err
}
