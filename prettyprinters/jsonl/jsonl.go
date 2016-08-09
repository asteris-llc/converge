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

type Node struct {
	ID    string      `json:"id"`
	Value interface{} `json:"value"`
}

type Edge struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
}

// JSONPrinter prints a graph in JSONL format
type JSONPrinter struct{}

func (j *JSONPrinter) DrawNode(graph *graph.Graph, nodeID string) (pp.Renderable, error) {
	out, err := json.Marshal(&Node{ID: nodeID, Value: graph.Get(nodeID)})
	return pp.VisibleString(string(out) + "\n"), err
}

func (j *JSONPrinter) DrawEdge(graph *graph.Graph, srcNodeID string, dstNodeID string) (pp.Renderable, error) {
	out, err := json.Marshal(&Edge{Source: srcNodeID, Destination: dstNodeID})
	return pp.VisibleString(string(out) + "\n"), err
}
