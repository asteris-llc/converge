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

// Package prettyprinters provides a general interface and concrete
// implementations for implementing prettyprinters.  This package was originally
// created to facilitate the development of graphviz visualizations for resource
// graphs, however it is intended to be useful for creating arbitrary output
// generators so that resource graph data can be used in other applications.
//
// See the 'examples' directory for examples of using the prettyprinter, and see
// the 'graphviz' package for an example of a concrete implementation of
// DigraphPrettyPrinter.
package prettyprinters

import (
	"bytes"

	"github.com/asteris-llc/converge/graph"
)

func New(p DigraphPrettyPrinter) Printer {
	return Printer{pp: p}
}

func (p Printer) Show(g *graph.Graph) (string, error) {
	var outputBuffer bytes.Buffer
	edges := g.Edges()
	subgraphs, subgraphJoinID := makeSubgraphMap()
	p.loadSubgraphs(g, subgraphs)
	rootNodes := subgraphs[subgraphJoinID].Nodes
	if str, err := p.pp.StartPP(g); err == nil {
		outputBuffer.WriteString(str)
	} else {
		return "", err
	}
	for idx := range rootNodes {
		if str, err := p.pp.DrawNode(g, rootNodes[idx]); err == nil {
			outputBuffer.WriteString(str)
		} else {
			return "", err
		}
	}

	for graphID, graph := range subgraphs {
		if graphID == subgraphJoinID {
			continue
		}

		if str, err := p.drawSubgraph(g, graphID, graph); err == nil {
			outputBuffer.WriteString(str)
		} else {
			return "", err
		}
	}

	if str, err := p.pp.StartEdgeSection(g); err == nil {
		outputBuffer.WriteString(str)
	} else {
		return "", err
	}

	for idx := range edges {
		src := edges[idx].Source
		dst := edges[idx].Dest
		if str, err := p.pp.DrawEdge(g, src, dst); err == nil {
			outputBuffer.WriteString(str)
		} else {
			return "", err
		}
	}

	if str, err := p.pp.StartEdgeSection(g); err == nil {
		outputBuffer.WriteString(str)
	} else {
		return "", err
	}

	if str, err := p.pp.FinishPP(g); err == nil {
		outputBuffer.WriteString(str)
	} else {
		return "", err
	}

	return outputBuffer.String(), nil
}

type Printer struct {
	pp DigraphPrettyPrinter
}
