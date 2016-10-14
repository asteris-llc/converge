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
	pp "github.com/asteris-llc/converge/prettyprinters"
)

// SubgraphMarkerKey is a type alias for an integer and represents the state
// values for subgraph tracking.
type SubgraphMarkerKey int

const (
	// SubgraphMarkerStart specifies that the current node is the beginning of a
	// subgraph.
	SubgraphMarkerStart SubgraphMarkerKey = iota

	// SubgraphMarkerEnd pecifies that the current node is the end of a subgraph.
	SubgraphMarkerEnd

	// SubgraphMarkerNOP pecifies that the current node should not change the
	// subgraph state.
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

// PrintProvider allows specific serializable types to be rendered as a
// Graphviz document.
type PrintProvider interface {

	// Given a graph entity, this function shall return a unique ID that will be
	// used for the vertex.  Note that this is not used for display, see
	// VertexGetLabel() for information on vertex displays.
	VertexGetID(GraphEntity) (pp.VisibleRenderable, error)

	// Defines the label associated with the vertex.  This will be the name
	// applied to the vertex when it is drawn.  Labels are stored in the vertex
	// attribute list automatically.  The 'label' parameter should therefore not
	// be returned as part of VertexGetProperties().
	VertexGetLabel(GraphEntity) (pp.VisibleRenderable, error)

	// VertexGetProperties allows the PrintProvider to provide additional
	// attributes to a given vertex.  Note that the 'label' attribute is special
	// and should not be returned as part of this call.
	VertexGetProperties(GraphEntity) PropertySet

	// Defines the label associated with the edge formed between two graph
	// entities.  This will be the name applied to the edge when it is drawn.
	EdgeGetLabel(GraphEntity, GraphEntity) (pp.VisibleRenderable, error)

	// EdgeGetProperties allows the PrintProvider to provide additional
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
	options       Options
	printProvider PrintProvider
	clusterIndex  int
}

// New will create a new graphviz.Printer with the options and print provider
// specified.
func New(opts Options, provider PrintProvider) *Printer {
	return &Printer{
		options:       opts,
		printProvider: provider,
		clusterIndex:  0,
	}
}

// MarkNode will call SubgraphMarker() on the print provider to determine
// whether the current node is the beginning of a subgraph.
func (p *Printer) MarkNode(g *graph.Graph, id string) *pp.SubgraphMarker {
	var val interface{}
	if meta, ok := g.Get(id); ok {
		val = meta.Value()
	}

	entity := GraphEntity{Name: id, Value: val}

	sgState := p.printProvider.SubgraphMarker(entity)
	subgraphID := p.clusterIndex
	switch sgState {
	case SubgraphMarkerStart:
		p.clusterIndex++
		return &pp.SubgraphMarker{
			SubgraphID: subgraphID,
			Start:      true,
		}
	case SubgraphMarkerEnd:
		return &pp.SubgraphMarker{
			SubgraphID: subgraphID,
			Start:      false,
		}
	case SubgraphMarkerNOP:
		return nil
	}
	return nil
}

// DrawNode prints node data by calling VertexGetID(), VertexGetLabel() and
// VertexGetProperties() on the associated print provider.  It will return a
// visible Renderable IFF the underlying VertexID is renderable.
func (p *Printer) DrawNode(g *graph.Graph, id string) (pp.Renderable, error) {
	var val interface{}
	if meta, ok := g.Get(id); ok {
		val = meta.Value()
	}

	graphEntity := GraphEntity{id, val}
	vertexID, err := p.printProvider.VertexGetID(graphEntity)
	if err != nil || !vertexID.Visible() {
		return pp.HiddenString(), err
	}

	vertexLabel, err := p.printProvider.VertexGetLabel(graphEntity)
	if err != nil {
		return pp.HiddenString(), err
	}

	attributes := p.printProvider.VertexGetProperties(graphEntity)
	attributes = maybeSetProperty(attributes, "label", escapeNewline(vertexLabel))
	attributeStr := buildAttributeString(attributes)
	vertexID = escapeNewline(vertexID)

	return pp.SprintfRenderable(
		true,
		"\"%s\" %s;\n",
		vertexID,
		attributeStr,
	), nil
}

// DrawEdge prints edge data in a fashion similar to DrawNode.  It will return a
// visible Renderable IFF the source and destination vertices are visible.
func (p *Printer) DrawEdge(g *graph.Graph, id1, id2 string) (pp.Renderable, error) {
	var srcVal, destVal interface{}

	if src, ok := g.Get(id1); ok {
		srcVal = src.Value()
	}

	if dest, ok := g.Get(id2); ok {
		destVal = dest.Value()
	}

	sourceEntity := GraphEntity{id1, srcVal}
	destEntity := GraphEntity{id2, destVal}

	sourceVertex, err := p.printProvider.VertexGetID(sourceEntity)
	if err != nil {
		return pp.HiddenString(), err
	}
	destVertex, err := p.printProvider.VertexGetID(destEntity)
	if err != nil {
		return pp.HiddenString(), err
	}
	label, err := p.printProvider.EdgeGetLabel(sourceEntity, destEntity)
	if err != nil {
		return pp.HiddenString(), err
	}
	attributes := p.printProvider.EdgeGetProperties(sourceEntity, destEntity)
	maybeSetProperty(attributes, "label", escapeNewline(label))

	edgeStr := fmt.Sprintf("\"%s\" -> \"%s\" %s;\n",
		escapeNewline(sourceVertex),
		escapeNewline(destVertex),
		buildAttributeString(attributes),
	)
	visible := sourceVertex.Visible() && destVertex.Visible()

	return pp.RenderableString(edgeStr, visible), nil
}

// StartSubgraph returns a string with the beginning of the subgraph cluster
func (p *Printer) StartSubgraph(g *graph.Graph, startNode string, subgraphID pp.SubgraphID) (pp.Renderable, error) {
	clusterStart := fmt.Sprintf("subgraph cluster_%d {\n", subgraphID.(int))
	return pp.VisibleString(clusterStart), nil
}

// FinishSubgraph provides the closing '}' for a subgraph
func (*Printer) FinishSubgraph(*graph.Graph, pp.SubgraphID) (pp.Renderable, error) {
	return pp.VisibleString("}\n"), nil
}

// StartPP begins the DOT output as an unnamed digraph.
func (p *Printer) StartPP(*graph.Graph) (pp.Renderable, error) {
	attrs := p.GraphAttributes()
	return pp.VisibleString(fmt.Sprintf("digraph {\n%s\n", attrs)), nil
}

// GraphAttributes returns a string containing all of the global graph
// attributes specified in Options.
func (p *Printer) GraphAttributes() string {
	var buffer bytes.Buffer
	if p.options.Splines != "" {
		_, _ = buffer.WriteString(fmt.Sprintf("splines = \"%s\";\n", p.options.Splines))
	}
	if p.options.Rankdir != "" {
		_, _ = buffer.WriteString(fmt.Sprintf("rankdir = \"%s\";\n", p.options.Rankdir))
	}
	return buffer.String()
}

// FinishPP returns the closing '}' to end the DOT file.
func (*Printer) FinishPP(*graph.Graph) (pp.Renderable, error) {
	return pp.VisibleString("}"), nil
}

// buildAttributeString takes a set of attribute keys and values and generates a
// DOT format attribute set in the form of [key1="value1",key2="value2"...]
func buildAttributeString(p PropertySet) string {
	var attrs []string
	if len(p) == 0 {
		return ""
	}
	for k, v := range p {
		attribute := fmt.Sprintf(` %s="%s" `, k, escapeNewline(pp.VisibleString(v)))
		attrs = append(attrs, attribute)
	}
	return fmt.Sprintf("[%s]", strings.Join(attrs, ","))
}

// DefaultProvider returns an empty BasicProvider as a convenience
func DefaultProvider() PrintProvider {
	return BasicProvider{}
}

// A GraphIDProvider is a basic PrintProvider for Graphviz that uses the Node ID
// from the digraph as the Vertex ID and label.
type GraphIDProvider struct {
	BasicProvider
}

// IDProvider is a convenience function to generate a GraphIDProvider.
func IDProvider() PrintProvider {
	return GraphIDProvider{}
}

// A BasicProvider is a basic PrintProvider for Graphviz that uses the value
// in the default format (%v) as the node label and id.
type BasicProvider struct{}

// VertexGetID provides a basic implementation that returns the %v quoted value
// of the node.
func (p BasicProvider) VertexGetID(e GraphEntity) (pp.VisibleRenderable, error) {
	return pp.VisibleString(fmt.Sprintf("%v", e.Value)), nil
}

// VertexGetLabel provides a basic implementation that returns the %v quoted
// value of the node (as with VertexGetID).
func (p BasicProvider) VertexGetLabel(e GraphEntity) (pp.VisibleRenderable, error) {
	return pp.VisibleString(fmt.Sprintf("%v", e.Value)), nil
}

// VertexGetProperties provides a basic implementation that returns an empty
// property set.
func (p BasicProvider) VertexGetProperties(GraphEntity) PropertySet {
	return PropertySet{}
}

// EdgeGetLabel provides a basic implementation leaves the edge unlabeled.
func (p BasicProvider) EdgeGetLabel(GraphEntity, GraphEntity) (pp.VisibleRenderable, error) {
	return pp.HiddenString(), nil
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
func (p GraphIDProvider) VertexGetID(e GraphEntity) (pp.VisibleRenderable, error) {
	return pp.VisibleString(e.Name), nil
}

// VertexGetLabel provides a basic implementation that uses the ID from the
// graph to generate the Vertex Label.
func (p GraphIDProvider) VertexGetLabel(e GraphEntity) (pp.VisibleRenderable, error) {
	return pp.VisibleString(e.Name), nil
}

// Replace embedded newlines with their escaped form.
func escapeNewline(r pp.Renderable) pp.VisibleRenderable {
	return pp.VisibleString(strings.Replace(r.String(), "\n", "\\n", -1))
}

// maybeSetProperty will add an attribute to the property set iff the value is
// renderable.
func maybeSetProperty(p PropertySet, key string, r pp.VisibleRenderable) PropertySet {
	if r.Visible() {
		p[key] = r.String()
	}
	return p
}
