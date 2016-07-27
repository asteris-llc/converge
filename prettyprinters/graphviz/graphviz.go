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

// Options specifies global graph options that can be configured for output.
// Arbitrary graphviz options are not supported.
type Options struct {
	// Specifies the way that connections between verices are drawn.  The default
	// is 'spline', other options include: 'line', 'orth', and 'none'.  See
	// http://www.graphviz.org/doc/info/attrs.html#d:splines
	Splines string

	// Specifies the direction of the graph.  See
	// http://www.graphviz.org/doc/info/attrs.html#d:rankdir for more information.
	Rankdir string
}

// DefaultOptions returns an Options struct with default values set.
func DefaultOptions() Options {
	return Options{}
}

// Printer is a DigraphPrettyPrinter implementation for drawing graphviz
// compatible DOT source code from a digraph.
type Printer struct {
	prettyprinters.DigraphPrettyPrinter
	options       Options
	printProvider GraphvizPrintProvider
	clusterIndex  int
}

// New will create a new graphviz.Printer with the options and print provider
// specified.
func New(opts Options, provider GraphvizPrintProvider) *Printer {
	return &Printer{
		options:       opts,
		printProvider: provider,
		clusterIndex:  0,
	}
}

// MarkNode will call SubgraphMarker() on the print provider to determine
// whether the current node is the beginning of a subgraph.
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

// DrawNode prints node data by calling VertexGetID(), VertexGetLabel() and
// VertexGetProperties() on the associated print provider.
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

// DrawEdge prints edge data in a fashion similar to DrawNode, bu.
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

// StartSubgraph returns a string with the beginning of the subgraph cluster
func (p *Printer) StartSubgraph(g *graph.Graph, startNode string, subgraphID prettyprinters.SubgraphID) (string, error) {
	clusterStart := fmt.Sprintf("subgraph cluster_%d {\n", subgraphID.(int))
	return clusterStart, nil
}

// FinishSubgraph provides the closing '}' for a subgraph
func (*Printer) FinishSubgraph(*graph.Graph, prettyprinters.SubgraphID) (string, error) {
	return "}\n", nil
}

// StartNodeSection would begin the node section of the output; DOT does not
// require any special formatting for a node section so we return "".
func (p *Printer) StartNodeSection(*graph.Graph) (string, error) {
	return "", nil
}

// FinishNodeSection would finish the section started by StartNodeSection.
// Since DOT has no special formatting for starting/ending node sections we
// return "".
func (p *Printer) FinishNodeSection(*graph.Graph) (string, error) {
	return "", nil
}

// StartEdgeSection returns "" because DOT doesn't require anything special for
// an edge section.
func (p *Printer) StartEdgeSection(*graph.Graph) (string, error) {
	return "", nil
}

// FinishEdgeSection returns "" because DOT doesnt' require anything special for
// an edge section.
func (p *Printer) FinishEdgeSection(*graph.Graph) (string, error) {
	return "", nil
}

// StartPP begins the DOT output as an unnamed digraph.
func (p *Printer) StartPP(*graph.Graph) (string, error) {
	attrs := p.GraphAttributes()
	return fmt.Sprintf("digraph {\n%s\n", attrs), nil
}

// GraphAttributes returns a string containing all of the global graph
// attributes specified in Options.
func (p *Printer) GraphAttributes() string {
	var buffer bytes.Buffer
	if p.options.Splines != "" {
		buffer.WriteString(fmt.Sprintf("splines = \"%s\";\n", p.options.Splines))
	}
	if p.options.Rankdir != "" {
		buffer.WriteString(fmt.Sprintf("rankdir = \"%s\";\n", p.options.Rankdir))
	}
	return buffer.String()
}

// FinishPP returns the closing '}' to end the DOT file.
func (*Printer) FinishPP(*graph.Graph) (string, error) {
	return "}", nil
}

// buildAttributeString takes a set of attribute keys and values and generates a
// DOT format attribute set in the form of [key1="value1",key2="value2"...]
func buildAttributeString(p PropertySet) string {
	var attrs []string
	for k, v := range p {
		attribute := fmt.Sprintf(` %s="%s" `, k, escapeNewline(v))
		attrs = append(attrs, attribute)
	}
	return fmt.Sprintf("[%s]", strings.Join(attrs, ","))
}

// DefaultProvider returns an empty BasicProvider as a convenience
func DefaultProvider() GraphvizPrintProvider {
	return BasicProvider{}
}

// A GraphIDProvider is a basic PrintProvider for Graphviz that uses the Node ID
// from the digraph as the Vertex ID and label.
type GraphIDProvider struct {
	BasicProvider
}

// IDProvider is a convenience function to generate a GraphIDProvider.
func IDProvider() GraphvizPrintProvider {
	return GraphIDProvider{}
}

// A BasicProvider is a basic PrintProvider for Graphviz that uses the value
// in the default format (%v) as the node label and id.
type BasicProvider struct{}

// VertexGetID provides a basic implementation that returns the %v quoted value
// of the node.
func (p BasicProvider) VertexGetID(e GraphEntity) (string, error) {
	return fmt.Sprintf("%v", e.Value), nil
}

// VertexGetLabel provides a basic implementation that returns the %v quoted
// value of the node (as with VertexGetID).
func (p BasicProvider) VertexGetLabel(e GraphEntity) (string, error) {
	return fmt.Sprintf("%v", e.Value), nil
}

// VertexGetProperties provides a basic implementation that returns an empty
// property set.
func (p BasicProvider) VertexGetProperties(GraphEntity) PropertySet {
	return PropertySet{}
}

// EdgeGetLabel provides a basic implementation leaves the edge unlabeled.
func (p BasicProvider) EdgeGetLabel(GraphEntity, GraphEntity) (string, error) {
	return "", nil
}

// EdgeGetProperties provides a basic implementation that returns an empty
// property set.
func (p BasicProvider) EdgeGetProperties(GraphEntity, GraphEntity) PropertySet {
	return PropertySet{}
}

// SubgraphMarker provides a basic implementation that returns a NOP.
func (p BasicProvider) SubgraphMarker(GraphEntity) SubgraphMarkerKey {
	return SubgraphMarkerNOP
}

// VertexGetID provides a basic implementation that uses the ID from the graph
// to generate the VertexID.
func (p GraphIDProvider) VertexGetID(e GraphEntity) (string, error) {
	return e.Name, nil
}

// VertexGetLabel provides a basic implementation that uses the ID from the
// graph to generate the Vertex Label.
func (p GraphIDProvider) VertexGetLabel(e GraphEntity) (string, error) {
	return e.Name, nil
}

// Replace embedded newlines with their escaped form.
func escapeNewline(s string) string {
	return strings.Replace(s, "\n", "\\n", -1)
}
