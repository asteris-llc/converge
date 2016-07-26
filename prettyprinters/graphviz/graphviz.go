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

// graphviz provides a concrete prettyprinters.DigraphPrettyPrinter
// implementation for rendering directed graphs as Graphviz-compatible dot
// source files.  It exports an interface, graphviz.GraphvizPrintProvider, that
// allows users to provide general methods for rendering graphed data types into
// graphviz documents.
package graphviz

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/prettyprinters"
)

// SubgraphMarkerKey is a type alias for an integer and represents the state
// values for subgraph tracking.
type SubgraphMarkerKey int

const (
	// Specifies that the current node is the beginning of a subgraph.
	SubgraphMarkerStart SubgraphMarkerKey = iota

	// Specifies that the current node is the end of a subgraph.
	SubgraphMarkerEnd

	// Specifies that the current node should not change the subgraph state.
	SubgraphMarkerNOP
)

// PropertySet is a map of graphviz compatible node or edge options and their
// values.  See http://www.graphviz.org/doc/info/attrs.html for a list of
// supported attributes.  Node that graph and subgraph attributes are not
// currently supported.
type PropertySet map[string]string

// GraphEntity defines the value at a given vertex, providing both it's
// canonical name and the untyped value associated with the vertex.
type GraphEntity struct {
	Name  string      // The canonical node ID
	Value interface{} // The value at the node
}

// The GraphvizPrinterProvider interface allows specific serializable types to
// be rendered as a Graphviz document.
type GraphvizPrintProvider interface {

	// Given a graph entity, this function shall return a unique ID that will be
	// used for the vertex.  Note that this is not used for display, see
	// VertexGetLabel() for information on vertex displays.
	VertexGetID(GraphEntity) (string, error)

	// Defines the label associated with the vertex.  This will be the name
	// applied to the vertex when it is drawn.  Labels are stored in the vertex
	// attribute list automatically.  The 'label' parameter should therefore not
	// be returned as part of VertexGetProperties().
	VertexGetLabel(GraphEntity) (string, error)

	// VertexGetProperties allows the GraphvizPrintProvider to provide additional
	// attributes to a given vertex.  Note that the 'label' attribute is special
	// and should not be returned as part of this call.
	VertexGetProperties(GraphEntity) PropertySet

	// Defines the label associated with the edge formed between two graph
	// entities.  This will be the name applied to the edge when it is drawn.
	EdgeGetLabel(GraphEntity, GraphEntity) (string, error)

	// EdgeGetProperties allows the GraphvizPrintProvider to provide additional
	// attributes to a given edge.  Note that the 'label' attribute is special
	// and should not be returned as part of this call.
	EdgeGetProperties(GraphEntity, GraphEntity) PropertySet

	// Allows the underlying provider to identify vertices that are the beginning
	// of a subgraph.  The function should return one of:
	//   SubgraphMarkerStart
	//   SubgraphMarkerEnd
	//   SubgraphMarkerNOP
	// in order to specify the vertexs' relation to subgraphs.
	SubgraphMarker(GraphEntity) SubgraphMarkerKey
}

type Options struct {
	Splines string
	Rankdir string
}

func DefaultOptions() Options {
	return Options{
		Splines: "spline",
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

func (p *Printer) MarkNode(g *graph.Graph, id string) *prettyprinters.SubgraphMarker {
	entity := GraphEntity{Name: id, Value: g.Get(id)}
	sgState := p.printProvider.SubgraphMarker(entity)
	subgraphID := p.clusterIndex
	switch sgState {
	case SubgraphMarkerStart:
		p.clusterIndex++
		return &prettyprinters.SubgraphMarker{
			SubgraphID: subgraphID,
			Start:      true,
		}
	case SubgraphMarkerEnd:
		return &prettyprinters.SubgraphMarker{
			SubgraphID: subgraphID,
			Start:      false,
		}
	case SubgraphMarkerNOP:
		return nil
	}
	return nil
}

func (p *Printer) DrawNode(g *graph.Graph, id string) (string, error) {
	graphValue := g.Get(id)
	graphEntity := GraphEntity{id, graphValue}
	vertexID, err := p.printProvider.VertexGetID(graphEntity)
	if err != nil {
		return "", err
	}
	vertexLabel, err := p.printProvider.VertexGetLabel(graphEntity)
	if err != nil {
		return "", err
	}
	attributes := p.printProvider.VertexGetProperties(graphEntity)
	attributes["label"] = escapeNewline(vertexLabel)
	dotCode := fmt.Sprintf("\"%s\" %s;\n", escapeNewline(vertexID), buildAttributeString(attributes))
	return dotCode, nil
}

func (p *Printer) DrawEdge(g *graph.Graph, id1, id2 string) (string, error) {
	sourceEntity := GraphEntity{id1, g.Get(id1)}
	destEntity := GraphEntity{id2, g.Get(id2)}
	sourceVertex, err := p.printProvider.VertexGetID(sourceEntity)
	if err != nil {
		return "", err
	}
	destVertex, err := p.printProvider.VertexGetID(destEntity)
	if err != nil {
		return "", err
	}
	label, err := p.printProvider.EdgeGetLabel(sourceEntity, destEntity)
	if err != nil {
		return "", err
	}
	attributes := p.printProvider.EdgeGetProperties(sourceEntity, destEntity)
	attributes["label"] = escapeNewline(label)
	return fmt.Sprintf("\"%s\" -> \"%s\" %s;\n",
		escapeNewline(sourceVertex),
		escapeNewline(destVertex),
		buildAttributeString(attributes),
	), nil
}

func (p *Printer) StartSubgraph(g *graph.Graph, startNode string, subgraphID prettyprinters.SubgraphID) (string, error) {
	clusterStart := fmt.Sprintf("subgraph cluster_%d {\n", subgraphID.(int))
	return clusterStart, nil
}

func (*Printer) FinishSubgraph(*graph.Graph, prettyprinters.SubgraphID) (string, error) {
	return "}\n", nil
}

func (p *Printer) StartNodeSection(*graph.Graph) (string, error) {
	return "", nil
}

func (p *Printer) FinishNodeSection(*graph.Graph) (string, error) {
	return "", nil
}

func (p *Printer) StartEdgeSection(*graph.Graph) (string, error) {
	return "### Beginning of Edge Section\n", nil
}

func (p *Printer) FinishEdgeSection(*graph.Graph) (string, error) {
	return "", nil
}

func (p *Printer) StartPP(*graph.Graph) (string, error) {
	attrs := p.GraphAttributes()
	return fmt.Sprintf("digraph {\n%s\n", attrs), nil
}

func (p *Printer) GraphAttributes() string {
	opts := make(map[string]string)
	var buffer bytes.Buffer
	opts["splines"] = p.options.Splines
	opts["rankdir"] = p.options.Rankdir
	for k, v := range opts {
		buffer.WriteString(fmt.Sprintf("%s = \"%s\";\n", k, v))
	}
	return buffer.String()
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

type GraphIDProvider struct{}

func IDProvider() GraphvizPrintProvider {
	return GraphIDProvider{}
}

type BasicProvider struct{}

func (p BasicProvider) VertexGetID(e GraphEntity) (string, error) {
	return fmt.Sprintf("%v", e.Value), nil
}

func (p BasicProvider) VertexGetLabel(e GraphEntity) (string, error) {
	return fmt.Sprintf("%v", e.Value), nil
}

func (p BasicProvider) VertexGetProperties(GraphEntity) PropertySet {
	return make(map[string]string)
}

func (p BasicProvider) EdgeGetLabel(GraphEntity, GraphEntity) (string, error) {
	return "", nil
}

func (p BasicProvider) EdgeGetProperties(GraphEntity, GraphEntity) PropertySet {
	return make(map[string]string)
}

func (p BasicProvider) SubgraphMarker(GraphEntity) SubgraphMarkerKey {
	return SubgraphMarkerNOP
}

func escapeNewline(s string) string {
	return strings.Replace(s, "\n", "\\n", -1)
}

func (p GraphIDProvider) VertexGetID(e GraphEntity) (string, error) {
	return fmt.Sprintf("%v", e.Name), nil
}

func (p GraphIDProvider) VertexGetLabel(e GraphEntity) (string, error) {
	return fmt.Sprintf("%v", e.Name), nil
}

func (p GraphIDProvider) VertexGetProperties(GraphEntity) PropertySet {
	return make(map[string]string)
}

func (p GraphIDProvider) EdgeGetLabel(GraphEntity, GraphEntity) (string, error) {
	return "", nil
}

func (p GraphIDProvider) EdgeGetProperties(GraphEntity, GraphEntity) PropertySet {
	return make(map[string]string)
}

func (p GraphIDProvider) SubgraphMarker(GraphEntity) SubgraphMarkerKey {
	return SubgraphMarkerNOP
}
