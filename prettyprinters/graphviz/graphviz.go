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
	VertexGetID(GraphEntity) (pp.Renderable, error)

	// Defines the label associated with the vertex.  This will be the name
	// applied to the vertex when it is drawn.  Labels are stored in the vertex
	// attribute list automatically.  The 'label' parameter should therefore not
	// be returned as part of VertexGetProperties().
	VertexGetLabel(GraphEntity) (pp.Renderable, error)

	// VertexGetProperties allows the PrintProvider to provide additional
	// attributes to a given vertex.  Note that the 'label' attribute is special
	// and should not be returned as part of this call.
	VertexGetProperties(GraphEntity) PropertySet

	// Defines the label associated with the edge formed between two graph
	// entities.  This will be the name applied to the edge when it is drawn.
	EdgeGetLabel(GraphEntity, GraphEntity) (pp.Renderable, error)

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
	pp.DigraphPrettyPrinter
	options       Options
	printProvider PrintProvider
	clusterIndex  int
	nodeClusters  map[string]int
	clusterEdges  map[int][]int
}

// New will create a new graphviz.Printer with the options and print provider
// specified.
func New(opts Options, provider PrintProvider) *Printer {
	return &Printer{
		options:       opts,
		printProvider: provider,
		clusterIndex:  0,
		nodeClusters:  make(map[string]int),
		clusterEdges:  make(map[int][]int),
	}
}

// MarkNode will call SubgraphMarker() on the print provider to determine
// whether the current node is the beginning of a subgraph.
func (p *Printer) MarkNode(g *graph.Graph, id string) *pp.SubgraphMarker {
	entity := GraphEntity{Name: id, Value: g.Get(id)}
	sgState := p.printProvider.SubgraphMarker(entity)
	subgraphID := p.clusterIndex
	switch sgState {
	case SubgraphMarkerStart:
		p.nodeClusters[id] = subgraphID
		p.clusterIndex++
		fmt.Printf("starting subgraph for %s\n", id)
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
	graphValue := g.Get(id)
	graphEntity := GraphEntity{id, graphValue}
	vertexID, err := p.printProvider.VertexGetID(graphEntity)
	if err != nil {
		return pp.HiddenString(""), err
	}
	vertexLabel, err := p.printProvider.VertexGetLabel(graphEntity)
	if err != nil {
		return pp.HiddenString(""), err
	}
	attributes := p.printProvider.VertexGetProperties(graphEntity)
	attributes = maybeSetProperty(attributes, "label", escapeNewline(vertexLabel))
	attributeStr := buildAttributeString(attributes)
	vertexID = escapeNewline(vertexID)
	dotCode := pp.ApplyRenderable(vertexID, func(s string) string {
		return fmt.Sprintf("\"%s\" %s;\n", s, attributeStr)
	})
	return dotCode, nil
}

// DrawEdge prints edge data in a fashion similar to DrawNode.  It will return a
// visible Renderable IFF the source and destination vertices are visible.
func (p *Printer) DrawEdge(g *graph.Graph, id1, id2 string) (pp.Renderable, error) {
	sourceEntity := GraphEntity{id1, g.Get(id1)}
	destEntity := GraphEntity{id2, g.Get(id2)}
	attributes := p.printProvider.EdgeGetProperties(sourceEntity, destEntity)

	sourceVertex, err := p.printProvider.VertexGetID(sourceEntity)

	if err != nil {
		return pp.HiddenString(""), err
	}

	destVertex, err := p.printProvider.VertexGetID(destEntity)
	if err != nil {
		return pp.HiddenString(""), err
	}

	label, err := p.printProvider.EdgeGetLabel(sourceEntity, destEntity)
	if err != nil {
		return pp.HiddenString(""), err
	}

	maybeSetProperty(attributes, "label", escapeNewline(label))

	if sourceVertex.Visible() && destVertex.Visible() {
		return pp.VisibleString(fmt.Sprintf("\"%s\" -> \"%s\" %s;\n",
			escapeNewline(sourceVertex),
			escapeNewline(destVertex),
			buildAttributeString(attributes))), nil
	}

	if isSame, err := p.sameCluster(g, id1, id2); err != nil || isSame {
		return pp.HiddenString(""), err
	}

	if sourceVertex, err = p.maybeClusterizeVertex(g, true, &attributes, id1, sourceVertex); err != nil {
		return pp.HiddenString(""), err
	}
	if destVertex, err = p.maybeClusterizeVertex(g, false, &attributes, id2, destVertex); err != nil {
		return pp.HiddenString(""), err
	}

	srcID, fndsrc, _ := p.getCluster(g, id1)
	dstID, fnddst, _ := p.getCluster(g, id2)

	res := fmt.Sprintf("# %s (%v:%d) -> %s (%v:%d) \n", id1, fndsrc, srcID, id2, fnddst, dstID)

	return pp.VisibleString(fmt.Sprintf("%s\"%s\" -> \"%s\" %s;\n",
		res,
		escapeNewline(sourceVertex),
		escapeNewline(destVertex),
		buildAttributeString(attributes))), nil
}

func (p *Printer) sameCluster(g *graph.Graph, id1, id2 string) (bool, error) {
	same := false
	cluster1, found1, err := p.getCluster(g, id1)
	if err != nil {
		return false, err
	}
	cluster2, found2, err := p.getCluster(g, id2)
	if err != nil {
		return false, err
	}

	if cluster1 == cluster2 {
		same = true
	}

	if !found1 && !found2 {
		same = false
	}
	return same, nil
}

func (p *Printer) getCluster(g *graph.Graph, id string) (int, bool, error) {
	clusterID, found := p.nodeClusters[id]
	if found {
		return clusterID, true, nil
	}
	root, err := g.Root()
	if err != nil {
		return 0, false, err
	}
	if id == root {
		return 0, true, nil
	}
	parent := graph.ParentID(id)
	return p.getCluster(g, parent)
}

// StartSubgraph returns a string with the beginning of the subgraph cluster
func (p *Printer) StartSubgraph(g *graph.Graph, startNode string, subgraphID pp.SubgraphID) (pp.Renderable, error) {
	fmt.Printf("startSubgraph on: %s (%d)\n", startNode, subgraphID)
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("subgraph cluster_%d {\n", subgraphID.(int)))
	for attr, val := range p.getSubgraphAttributes(g, startNode, subgraphID) {
		buffer.WriteString(fmt.Sprintf("%s = \"%s\";\n", attr, val))
	}
	buffer.WriteString(fmt.Sprintf("\"%s\" [style=\"invis\",fontsize=0.01,shape=none];\n", clusterMarkerNode(subgraphID)))
	return pp.VisibleString(buffer.String()), nil
}

func (p *Printer) getSubgraphAttributes(g *graph.Graph, node string, subgraphID pp.SubgraphID) map[string]string {
	fmt.Printf("# returning subgraph attributes for node %s\n", node)
	attributes := make(map[string]string)
	markerNode, err := p.printProvider.VertexGetID(GraphEntity{node, g.Get(node)})
	if err != nil || !(markerNode.Visible()) {
		attributes["label"] = node
	}
	return attributes
}

// FinishSubgraph provides the closing '}' for a subgraph
func (*Printer) FinishSubgraph(*graph.Graph, pp.SubgraphID) (pp.Renderable, error) {
	return pp.VisibleString("}\n"), nil
}

// StartNodeSection would begin the node section of the output; DOT does not
// require any special formatting for a node section so we return "".
func (p *Printer) StartNodeSection(*graph.Graph) (pp.Renderable, error) {
	return pp.HiddenString(""), nil
}

// FinishNodeSection would finish the section started by StartNodeSection.
// Since DOT has no special formatting for starting/ending node sections we
// return "".
func (p *Printer) FinishNodeSection(*graph.Graph) (pp.Renderable, error) {
	return pp.HiddenString(""), nil
}

// StartEdgeSection returns "" because DOT doesn't require anything special for
// an edge section.
func (p *Printer) StartEdgeSection(*graph.Graph) (pp.Renderable, error) {
	return pp.HiddenString(""), nil
}

// FinishEdgeSection returns "" because DOT doesnt' require anything special for
// an edge section.
func (p *Printer) FinishEdgeSection(*graph.Graph) (pp.Renderable, error) {
	return pp.HiddenString(""), nil
}

// StartPP begins the DOT output as an unnamed digraph.
func (p *Printer) StartPP(*graph.Graph) (pp.Renderable, error) {
	attrs := p.GraphAttributes()
	return pp.VisibleString(fmt.Sprintf("digraph {\ncompound=true;\n%s\n", attrs)), nil
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

func (p *Printer) alreadyHasClusterEdge(src, dst int) bool {
	dests, found := p.clusterEdges[src]
	if !found {
		return false
	}
	for _, knownDest := range dests {
		if dst == knownDest {
			return true
		}
	}
	return false
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
func (p BasicProvider) VertexGetID(e GraphEntity) (pp.Renderable, error) {
	return pp.VisibleString(fmt.Sprintf("%v", e.Value)), nil
}

// VertexGetLabel provides a basic implementation that returns the %v quoted
// value of the node (as with VertexGetID).
func (p BasicProvider) VertexGetLabel(e GraphEntity) (pp.Renderable, error) {
	return pp.VisibleString(fmt.Sprintf("%v", e.Value)), nil
}

// VertexGetProperties provides a basic implementation that returns an empty
// property set.
func (p BasicProvider) VertexGetProperties(GraphEntity) PropertySet {
	return PropertySet{}
}

// EdgeGetLabel provides a basic implementation leaves the edge unlabeled.
func (p BasicProvider) EdgeGetLabel(GraphEntity, GraphEntity) (pp.Renderable, error) {
	return pp.HiddenString(""), nil
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
func (p GraphIDProvider) VertexGetID(e GraphEntity) (pp.Renderable, error) {
	return pp.VisibleString(e.Name), nil
}

// VertexGetLabel provides a basic implementation that uses the ID from the
// graph to generate the Vertex Label.
func (p GraphIDProvider) VertexGetLabel(e GraphEntity) (pp.Renderable, error) {
	return pp.VisibleString(e.Name), nil
}

// Replace embedded newlines with their escaped form.
func escapeNewline(r pp.Renderable) pp.Renderable {
	return pp.ApplyRenderable(r, func(s string) string {
		return strings.Replace(s, "\n", "\\n", -1)
	})
}

// maybeSetProperty will add an attribute to the property set iff the value is
// renderable.
func maybeSetProperty(p PropertySet, key string, r pp.Renderable) PropertySet {
	if r.Visible() {
		p[key] = r.String()
	}
	return p
}

func clusterMarkerNode(id pp.SubgraphID) string {
	return fmt.Sprintf("__internal_cluser_%02d", id)
}

func (p *Printer) maybeClusterizeVertex(g *graph.Graph, src bool, attributes *PropertySet, id string, r pp.Renderable) (pp.Renderable, error) {
	if r.Visible() {
		return r, nil
	}
	clusterID, found, err := p.getCluster(g, id)
	if (err != nil) || (!found) {
		return r, err
	}

	var arrowEnd string
	if src {
		arrowEnd = "ltail"
	} else {
		arrowEnd = "lhead"
	}

	*attributes = maybeSetProperty(*attributes, arrowEnd, pp.VisibleString(fmt.Sprintf("cluster_%d", clusterID)))
	return pp.VisibleString(clusterMarkerNode(clusterID)), nil
}
