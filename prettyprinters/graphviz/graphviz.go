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
	"fmt"

	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/prettyprinters"
)

// SubgraphMarkerKey is a type alias for an integer and represents the state
// values for subgraph tracking.
//
//   SubgraphMarkerStart
//   SubgraphMarkerEnd
//   SubgraphMarkerNOP
//
type SubgraphMarkerKey int

const (
	SubgraphMarkerStart SubgraphMarkerKey = iota
	SubgraphMarkerEnd
	SubgraphMarkerNOP
)

type PropertySet map[string]string

// The GraphvizPrinterProvider interface allows specific serializable types to
// be rendered as a Graphviz document.
type GraphvizPrintProvider interface {
	VertexGetID(interface{}) (string, error)
	VertexGetLabel(interface{}) (string, error)
	VertexGetProperties(interface{}) PropertySet
	EdgeGetLabel(interface{}, interface{}) (string, error)
	EdgeGetProperties(interface{}, interface{}) PropertySet
	SubgraphMarker(interface{}) SubgraphMarkerKey
}

type Options struct {
	Splines string
	Rankdir string
}

func DefaultOptions() Options {
	return Options{
		Splines: "curved",
		Rankdir: "LR",
	}
}

type Printer struct {
	prettyprinters.DigraphPrettyPrinter
	options       Options
	printProvider GraphvizPrintProvider
	clusterIndex  int
}

func New(opts Options, provider GraphvizPrintProvider) *Printer {
	return &Printer{
		options:       opts,
		printProvider: provider,
		clusterIndex:  0,
	}
}

func (p *Printer) DrawNode(g *graph.Graph, id string, sgMarker func(string, bool)) (string, error) {
	graphValue := g.Get(id)
	vertexID, err := p.printProvider.VertexGetID(graphValue)
	if err != nil {
		return "", err
	}
	vertexLabel, err := p.printProvider.VertexGetLabel(graphValue)
	if err != nil {
		return "", err
	}
	attributes := p.printProvider.VertexGetProperties(g.Get(id))
	switch p.printProvider.SubgraphMarker(g.Get(id)) {
	case SubgraphMarkerStart:
		sgMarker(id, true)
	case SubgraphMarkerEnd:
		sgMarker(id, false)
	}

	attributes["label"] = vertexLabel

	dotCode := fmt.Sprintf("\"%s\" %s;\n", vertexID, buildAttributeString(attributes))
	return dotCode, nil
}

func (p *Printer) DrawEdge(g *graph.Graph, id1, id2 string) (string, error) {
	sourceVertex, err := p.printProvider.VertexGetID(g.Get(id1))
	if err != nil {
		return "", err
	}
	destVertex, err := p.printProvider.VertexGetID(g.Get(id2))
	if err != nil {
		return "", err
	}
	label, err := p.printProvider.EdgeGetLabel(id1, id2)
	if err != nil {
		return "", err
	}
	attributes := p.printProvider.EdgeGetProperties(g.Get(id1), g.Get(id2))
	attributes["label"] = label
	return fmt.Sprintf("\"%s\" -> \"%s\" %s;\n", sourceVertex, destVertex, buildAttributeString(attributes)), nil
}

func (p *Printer) StartSubgraph(*graph.Graph, string) (string, error) {
	clusterStart := fmt.Sprintf("subgraph cluster_%d {\n", p.clusterIndex)
	p.clusterIndex++
	return clusterStart, nil
}

func (*Printer) FinishSubgraph(*graph.Graph, string) (string, error) {
	return "}\n", nil
}

func (p *Printer) StartNodeSection(*graph.Graph) (string, error) {
	return "", nil
}

func (p *Printer) FinishNodeSection(*graph.Graph) (string, error) {
	return "", nil
}

func (p *Printer) StartEdgeSection(*graph.Graph) (string, error) {
	return "", nil
}

func (p *Printer) FinishEdgeSection(*graph.Graph) (string, error) {
	return "", nil
}

func (*Printer) StartPP(*graph.Graph) (string, error) {
	return "digraph {\n", nil
}

func (*Printer) FinishPP(*graph.Graph) (string, error) {
	return "}", nil
}

func buildAttributeString(p PropertySet) string {
	accumulator := "["
	for k, v := range p {
		accumulator = fmt.Sprintf("%s %s=\"%s\",", accumulator, k, v)
	}
	accumulator = accumulator[0 : len(accumulator)-1]
	return fmt.Sprintf("%s]", accumulator)
}

func DefaultProvider() GraphvizPrintProvider {
	return BasicProvider{}
}

type BasicProvider struct{}

func (p BasicProvider) VertexGetID(i interface{}) (string, error) {
	return fmt.Sprintf("%p", &i), nil
}

func (p BasicProvider) VertexGetLabel(i interface{}) (string, error) {
	return fmt.Sprintf("%v", i), nil
}

func (p BasicProvider) VertexGetProperties(interface{}) PropertySet {
	return make(map[string]string)
}

func (p BasicProvider) EdgeGetLabel(interface{}, interface{}) (string, error) {
	return "", nil
}

func (p BasicProvider) EdgeGetProperties(interface{}, interface{}) PropertySet {
	return make(map[string]string)
}

func (p BasicProvider) SubgraphMarker(interface{}) SubgraphMarkerKey {
	return SubgraphMarkerNOP
}
