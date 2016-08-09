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

import (
	"bytes"

	"golang.org/x/net/context"

	"github.com/asteris-llc/converge/graph"
)

// New creates a new prettyprinter instance from the specified
// DigraphPrettyPrinter instance.
func New(p BasePrinter) Printer {
	return Printer{pp: p}
}

// Show will take the given graph and return a string representing the text
// output of the graph according to the associated prettyprinter.  If an error
// is returned at any stage of the prettyprinting process it is returned
// unmodified to the user.
func (p Printer) Show(ctx context.Context, g *graph.Graph) (string, error) {
	outputBuffer := new(bytes.Buffer)

	subgraphs := makeSubgraphMap()
	p.loadSubgraphs(ctx, g, subgraphs)
	rootNodes := subgraphs[SubgraphBottomID].Nodes

	graphPrinter, gpOK := p.pp.(GraphPrinter)
	if gpOK {
		if str, err := graphPrinter.StartPP(g); err == nil {
			writeRenderable(outputBuffer, str)
		} else {
			return "", err
		}
	}

	nodePrinter, npOK := p.pp.(NodePrinter)
	if npOK {
		for _, node := range rootNodes {
			if str, err := nodePrinter.DrawNode(g, node); err == nil {
				writeRenderable(outputBuffer, str)
			} else {
				return "", err
			}
		}
	}

	for graphID, graph := range subgraphs {
		if graphID == SubgraphBottomID {
			continue
		}

		if str, err := p.drawSubgraph(g, graphID, graph); err == nil {
			writeRenderable(outputBuffer, str)
		} else {
			return "", err
		}
	}

	if str, err := p.drawEdges(g); err == nil {
		writeRenderable(outputBuffer, str)
	} else {
		return "", err
	}

	if gpOK {
		if str, err := graphPrinter.FinishPP(g); err == nil {
			writeRenderable(outputBuffer, str)
		} else {
			return "", err
		}
	}

	return outputBuffer.String(), nil
}

func (p Printer) drawEdges(g *graph.Graph) (*StringRenderable, error) {
	edgePrinter, epOK := p.pp.(EdgePrinter)
	edgeSectionPrinter, espOK := p.pp.(EdgeSectionPrinter)

	if !epOK && !espOK {
		return HiddenString(""), nil
	}

	outputBuffer := new(bytes.Buffer)
	edges := g.Edges()
	if espOK {
		if str, err := edgeSectionPrinter.StartEdgeSection(g); err == nil {
			writeRenderable(outputBuffer, str)
		} else {
			return nil, err
		}
	}

	if epOK {
		for _, edge := range edges {
			if str, err := edgePrinter.DrawEdge(g, edge.Source, edge.Dest); err == nil {
				writeRenderable(outputBuffer, str)
			} else {
				return nil, err
			}
		}
	}

	if espOK {
		if str, err := edgeSectionPrinter.FinishEdgeSection(g); err == nil {
			writeRenderable(outputBuffer, str)
		} else {
			return nil, err
		}
	}

	return VisibleString(outputBuffer.String()), nil
}

// Printer is the top-level structure for a pretty printer.
type Printer struct {
	pp BasePrinter
}
