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

type GraphvizPrintProvider interface {
	VertexGetId(interface{}) (string, error)
	VertexGetLabel(interface{}) (string, error)
	SubgraphMarker(interface{}) SubgraphMarkerKey
}

type GraphvizOptions map[string]string

type GraphvizPrinter struct {
	prettyprinters.DigraphPrettyPrinter
	optsMap       GraphvizOptions
	printProvider GraphvizPrintProvider
}

func NewPrinter(opts GraphvizOptions, provider GraphvizPrintProvider) *GraphvizPrinter {
	opts = mergeDefaultOptions(opts)
	return &GraphvizPrinter{
		optsMap:       opts,
		printProvider: provider,
	}
}

func (p *GraphvizPrinter) DrawNode(g *graph.Graph, id string, sgMarker func(string, bool)) (string, error) {
	graphValue := g.Get(id)
	vertexId, err := p.printProvider.VertexGetId(graphValue)
	if err != nil {
		return "", err
	}
	vertexLabel, err := p.printProvider.VertexGetLabel(graphValue)
	if err != nil {
		return "", err
	}
	switch p.printProvider.SubgraphMarker(g.Get(id)) {
	case SubgraphMarkerStart:
		sgMarker(id, true)
	case SubgraphMarkerEnd:
		sgMarker(id, false)
	}

	dotCode := fmt.Sprintf(
		"\"%s\" [label=\"%s\"];\n",
		vertexId,
		vertexLabel,
	)
	return dotCode, nil
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

func (p *GraphvizPrinter) StartNodeSection(*graph.Graph) (string, error) {
	return "", nil
}

func (p *GraphvizPrinter) FinishNodeSection(*graph.Graph) (string, error) {
	return "", nil
}

func (p *GraphvizPrinter) StartEdgeSection(*graph.Graph) (string, error) {
	return "", nil
}

func (p *GraphvizPrinter) FinishEdgeSection(*graph.Graph) (string, error) {
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
