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

import "github.com/asteris-llc/converge/graph"

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

	StartSubgraph(*graph.Graph) (string, error)
	FinishSubgraph(*graph.Graph) (string, error)

	DrawNode(*graph.Graph, interface{}, func(*graph.Graph)) (string, error)
}
