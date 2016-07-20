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

package graphviz

import (
	"errors"
	"fmt"

	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/prettyprinters"
)

var (
	//BadOptionsError is returned when a GraphvizOptionMap contains invalid keys
	//or values, or is missing required values.
	BadOptionsError = errors.New("Invalid configuration option")

	//BadGraphError is returned when a provided graph.Graph is not valid
	BadGraphError = errors.New("Provided graph was invalid, or missing ID")
)

type SubgraphMarkerKey int

const (
	SubgraphMarkerStart SubgraphMarkerKey = iota
	SubgraphMarkerEnd
	SubgraphMarkerNop
)

type GraphvizOptions map[string]string

type VertexGetId func(interface{}) (string, error)
type VertexGetLabel func(interface{}) (string, error)
type SubgraphMarker func(interface{}) SubgraphMarkerKey

type GraphvizPrinter struct {
	prettyprinters.DigraphPrettyPrinter
	optsMap  GraphvizOptions
	vPrinter VertexGetLabel
	vLabeler VertexGetId
	vMarker  SubgraphMarker
}

func NewPrinter(opts GraphvizOptions, idF VertexGetId, labelF VertexGetLabel, marker SubgraphMarker) *GraphvizPrinter {
	opts = mergeDefaultOptions(opts)

	if idF == nil {
		idF = DefaultVertexId
	}

	if labelF == nil {
		labelF = func(val interface{}) (string, error) { return idF(val) }
	}

	if marker == nil {
		marker = DefaultMarker
	}

	return &GraphvizPrinter{
		optsMap:  opts,
		vPrinter: labelF,
		vLabeler: idF,
		vMarker:  marker,
	}
}

func (p *GraphvizPrinter) DrawNode(g *graph.Graph, id string, sgMarker func(string, bool)) (string, error) {
	switch p.vMarker(g.Get(id)) {
	case SubgraphMarkerStart:
		sgMarker(id, true)
	case SubgraphMarkerEnd:
		sgMarker(id, false)
	}

	if nodeStr, err := p.vPrinter(g.Get(id)); err != nil {
		return "", err
	} else {
		return fmt.Sprintf("\"%s\";\n", nodeStr), nil
	}
}

/* FIXME: Stubs*/
func (*GraphvizPrinter) StartPP(*graph.Graph) (string, error) {
	return "", nil
}

func (*GraphvizPrinter) FinishPP(*graph.Graph) (string, error) {
	return "", nil
}

func (*GraphvizPrinter) StartSubgraph(*graph.Graph) (string, error) {
	return "", nil
}

func (*GraphvizPrinter) FinishSubgraph(*graph.Graph) (string, error) {
	return "", nil
}

func (p *GraphvizPrinter) DrawEdge(g *graph.Graph, id1, id2 string) (string, error) {
	return "", nil
}

// Utility Functions

//mergeDefaultOptions iterates over an option map and copies the defaults for
//any missing entries.
func mergeDefaultOptions(opts GraphvizOptions) GraphvizOptions {
	for k, v := range DefaultOptions() {
		if _, ok := opts[k]; !ok {
			opts[k] = v
		}
	}
	return opts
}

func (g *GraphvizPrinter) Options() GraphvizOptions {
	return g.optsMap
}

var defaultOptions GraphvizOptions = map[string]string{
	"splines": "curved",
	"rankdir": "LR",
}

func DefaultOptions() GraphvizOptions {
	return defaultOptions
}

func DefaultVertexId(val interface{}) (string, error) {
	nodeStr := fmt.Sprintf("%x", val)
	return nodeStr, nil
}

func DefaultMarker(val interface{}) SubgraphMarkerKey {
	return SubgraphMarkerNop
}
