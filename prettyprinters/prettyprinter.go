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

//Package prettyprinters provides a general interface and concrete
//implementations for implementing prettyprinters.  This package was originally
//created to facilitate the development of graphviz visualizations for resource
//graphs, however it is intended to be useful for creating arbitrary output
//generators so that resource graph data can be used in other applications.
package prettyprinters

import (
	"bytes"
	"fmt"

	"github.com/asteris-llc/converge/graph"
)

//DigraphPrettyPrinter interface defines the minimal set of required functions
//for defining a pretty printer.
type DigraphPrettyPrinter interface {

	//StartPP will be given as it's argument a pointer to the root node of the
	//graph structure.  It should do any necessary work to create the beginning of
	//the document and do any first-pass walks of the graph that may be necessary
	//for rendering output.
	StartPP(*graph.Graph) (string, error)

	//FinishPP will be given as it's argument a pointer to the root node of the
	//graph structure.  It should do any necessary work to finish the generation
	//of the prettyprinted output.
	FinishPP(*graph.Graph) (string, error)

	//When DrawNode() calls the provided node mark function,
	StartSubgraph(*graph.Graph, string) (string, error)
	FinishSubgraph(*graph.Graph, string) (string, error)

	StartNodeSection(*graph.Graph) (string, error)
	FinishNodeSection(*graph.Graph) (string, error)
	DrawNode(*graph.Graph, string, func(string, bool)) (string, error)

	StartEdgeSection(*graph.Graph) (string, error)
	FinishEdgeSection(*graph.Graph) (string, error)
	DrawEdge(*graph.Graph, string, string) (string, error)
}

type Printer struct {
	pp DigraphPrettyPrinter
	gr *graph.Graph
}

func New(g *graph.Graph, p DigraphPrettyPrinter) Printer {
	return Printer{
		pp: p,
		gr: g,
	}
}

func (printer Printer) Show() (string, error) {
	var buffer bytes.Buffer
	subgraphs := make(map[string]bool)

	subgraphCB := func(id string, toggle bool) {
		subgraphs[id] = toggle
	}
	nodeStringGraph, graphTransformError := printer.gr.Transform(func(id string, gr *graph.Graph) error {
		printedNode, err := printer.pp.DrawNode(printer.gr, id, subgraphCB)
		if err == nil {
			gr.Add(id, printedNode)
		}
		return err
	})
	if graphTransformError != nil {
		return "", graphTransformError
	}

	if str, err := printer.pp.StartPP(printer.gr); err == nil {
		buffer.WriteString(str)
	} else {
		return "", err
	}

	if str, err := printer.pp.StartNodeSection(printer.gr); err == nil {
		buffer.WriteString(str)
	} else {
		return "", err
	}

	if graphTransformError != nil {
		fmt.Printf("graphTransformError: %s\n", graphTransformError)
		return "", graphTransformError
	}

	printer.gr.RootFirstWalk(func(id string, val interface{}) error {
		isSubgraphStart, ok := subgraphs[id]
		str := nodeStringGraph.Get(id).(string)
		var subgraphErr error = nil
		var subgraphStr string
		if !ok {
			buffer.WriteString(str)
			return nil
		}
		if isSubgraphStart {
			subgraphStr, subgraphErr = printer.pp.StartSubgraph(printer.gr, id)
		} else {
			subgraphStr, subgraphErr = printer.pp.FinishSubgraph(printer.gr, id)
		}
		if subgraphErr != nil {
			return subgraphErr
		}
		if isSubgraphStart {
			buffer.WriteString(subgraphStr)
			buffer.WriteString(str)
		} else {
			buffer.WriteString(str)
			buffer.WriteString(subgraphStr)
		}
		return nil
	})

	if str, err := printer.pp.FinishNodeSection(printer.gr); err == nil {
		buffer.WriteString(str)
	} else {
		return "", err
	}

	if str, err := printer.pp.StartEdgeSection(printer.gr); err == nil {
		buffer.WriteString(str)
	} else {
		return "", err
	}

	printer.gr.RootFirstWalk(func(id string, _ interface{}) error {
		edgeIDs := printer.gr.DownEdges(id)
		for idx := range edgeIDs {
			if str, err := printer.pp.DrawEdge(printer.gr, id, edgeIDs[idx]); err == nil {
				buffer.WriteString(str)
			} else {
				return err
			}
		}
		return nil
	})

	if str, err := printer.pp.FinishEdgeSection(printer.gr); err == nil {
		buffer.WriteString(str)
	} else {
		return "", err
	}
	if str, err := printer.pp.FinishPP(printer.gr); err == nil {
		buffer.WriteString(str)
	} else {
		return "", err
	}
	return buffer.String(), nil
}

func maybeAppend(b *bytes.Buffer, f func() (string, error)) error {
	str, err := f()
	if err != nil {
		b.WriteString(str)
	}
	return err
}
