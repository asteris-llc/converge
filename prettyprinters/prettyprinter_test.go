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

package prettyprinters_test

import (
	"fmt"
	"os"

	"golang.org/x/net/context"

	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/load"
	"github.com/asteris-llc/converge/prettyprinters"
	"github.com/asteris-llc/converge/prettyprinters/graphviz"
	"github.com/asteris-llc/converge/prettyprinters/graphviz/providers"
)

func ExampleShowGraphWithDefaultProvider() {
	g := createTestGraph()
	valuePrinter := prettyprinters.New(graphviz.New(graphviz.DefaultOptions(), graphviz.DefaultProvider()))
	valueDotCode, _ := valuePrinter.Show(context.Background(), g)
	fmt.Println(valueDotCode)

	// Output:
	// digraph {

	// "1" [ label="1" ];
	// "2" [ label="2" ];
	// "3" [ label="3" ];
	// "1" -> "2" [ label="" ];
	// "1" -> "3" [ label="" ];
	// }
}

func ExampleShowGraphWithIDProvider() {
	g := createTestGraph()

	namePrinter := prettyprinters.New(graphviz.New(graphviz.DefaultOptions(), graphviz.IDProvider()))
	nameDotCode, _ := namePrinter.Show(context.Background(), g)
	fmt.Println(nameDotCode)

	// Output:
	// digraph {
	// splines = "spline";
	// rankdir = "LR";

	// "a" [ label="a"];
	// "a/b" [ label="a/b"];
	// "a/c" [ label="a/c"];
	// "a" -> "a/b" [ label=""];
	// "a" -> "a/c" [ label=""];
	// }
}

func ExampleLoadAndPrint() {
	g, err := load.Load(context.Background(), os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}
	printer := prettyprinters.New(graphviz.New(graphviz.DefaultOptions(), providers.ResourcePreparer()))
	dotCode, err := printer.Show(context.Background(), g)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(dotCode)
}

func ExampleCustomProvider() {

	g := graph.New()
	g.Add(graph.ID("a"), 1)
	g.Add(graph.ID("a", "b"), 2)
	g.Add(graph.ID("a", "c"), 3)
	g.Add(graph.ID("a", "c", "d"), 4)
	g.Add(graph.ID("a", "c", "e"), 5)
	g.Add(graph.ID("a", "b", "f"), 6)
	g.Add(graph.ID("a", "b", "g"), 7)

	g.Add(graph.ID("a", "c", "d", "h"), 8)
	g.Add(graph.ID("a", "c", "d", "i"), 9)
	g.Add(graph.ID("a", "c", "d", "j"), 10)

	g.Add(graph.ID("a", "c", "e", "k"), 11)
	g.Add(graph.ID("a", "c", "e", "l"), 12)
	g.Add(graph.ID("a", "c", "e", "m"), 13)

	g.Add(graph.ID("a", "b", "f", "n"), 14)
	g.Add(graph.ID("a", "b", "f", "o"), 15)
	g.Add(graph.ID("a", "b", "f", "p"), 16)

	g.Add(graph.ID("a", "b", "g", "q"), 17)
	g.Add(graph.ID("a", "b", "g", "r"), 18)
	g.Add(graph.ID("a", "b", "g", "s"), 19)

	g.Connect(graph.ID("a"), graph.ID("a", "b"))
	g.Connect(graph.ID("a"), graph.ID("a", "c"))
	g.Connect(graph.ID("a", "c"), graph.ID("a", "c", "d"))
	g.Connect(graph.ID("a", "c"), graph.ID("a", "c", "e"))
	g.Connect(graph.ID("a", "b"), graph.ID("a", "b", "f"))
	g.Connect(graph.ID("a", "b"), graph.ID("a", "b", "g"))

	g.Connect(graph.ID("a", "c", "d"), graph.ID("a", "c", "d", "h"))
	g.Connect(graph.ID("a", "c", "d"), graph.ID("a", "c", "d", "i"))
	g.Connect(graph.ID("a", "c", "d"), graph.ID("a", "c", "d", "j"))

	g.Connect(graph.ID("a", "c", "e"), graph.ID("a", "c", "e", "k"))
	g.Connect(graph.ID("a", "c", "e"), graph.ID("a", "c", "e", "l"))
	g.Connect(graph.ID("a", "c", "e"), graph.ID("a", "c", "e", "m"))

	g.Connect(graph.ID("a", "b", "f"), graph.ID("a", "b", "f", "n"))
	g.Connect(graph.ID("a", "b", "f"), graph.ID("a", "b", "f", "o"))
	g.Connect(graph.ID("a", "b", "f"), graph.ID("a", "b", "f", "p"))

	g.Connect(graph.ID("a", "b", "g"), graph.ID("a", "b", "g", "q"))
	g.Connect(graph.ID("a", "b", "g"), graph.ID("a", "b", "g", "r"))
	g.Connect(graph.ID("a", "b", "g"), graph.ID("a", "b", "g", "s"))

	numberPrinter := prettyprinters.New(graphviz.New(graphviz.DefaultOptions(), NumberProvider{}))
	dotCode, _ := numberPrinter.Show(context.Background(), g)
	fmt.Println(dotCode)

	// Output:
	// digraph {
	// splines = "spline";
	// rankdir = "LR";

	// "1" [ label="1"];
	// "14" [ label="14"];
	// "15" [ label="15"];
	// "16" [ label="16"];
	// "17" [ label="17"];
	// "18" [ label="18"];
	// "19" [ label="19"];
	// "8" [ label="8"];
	// "9" [ label="9"];
	// "10" [ label="10"];
	// "11" [ label="11"];
	// "12" [ label="12"];
	// "13" [ label="13"];
	// subgraph cluster_0 {
	// "2" [ label="2"];
	// "6" [ label="6"];
	// "7" [ label="7"];
	// }
	// subgraph cluster_1 {
	// "3" [ label="3"];
	// "4" [ label="4"];
	// "5" [ label="5"];
	// }
	// "7" -> "17" [ label=""];
	// "1" -> "2" [ label=""];
	// "4" -> "9" [ label=""];
	// "4" -> "10" [ label=""];
	// "6" -> "16" [ label=""];
	// "7" -> "18" [ label=""];
	// "1" -> "3" [ label=""];
	// "5" -> "11" [ label=""];
	// "5" -> "12" [ label=""];
	// "6" -> "15" [ label=""];
	// "2" -> "6" [ label=""];
	// "4" -> "8" [ label=""];
	// "5" -> "13" [ label=""];
	// "7" -> "19" [ label=""];
	// "3" -> "4" [ label=""];
	// "3" -> "5" [ label=""];
	// "2" -> "7" [ label=""];
	// "6" -> "14" [ label=""];
	// }

}

type NumberProvider struct {
	graphviz.BasicProvider
}

func (p NumberProvider) SubgraphMarker(e graphviz.GraphEntity) graphviz.SubgraphMarkerKey {
	val := e.Value.(int)

	if val == 2 || val == 3 {
		return graphviz.SubgraphMarkerStart
	}

	if val == 4 || val == 5 || val == 6 || val == 7 {
		return graphviz.SubgraphMarkerEnd
	}
	return graphviz.SubgraphMarkerNOP
}

func createTestGraph() *graph.Graph {
	g := graph.New()
	g.Add(graph.ID("a"), 1)
	g.Add(graph.ID("a", "b"), 2)
	g.Add(graph.ID("a", "c"), 3)
	g.Connect(graph.ID("a"), graph.ID("a", "b"))
	g.Connect(graph.ID("a"), graph.ID("a", "c"))
	return g
}
