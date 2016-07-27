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
	g := graph.New()
	g.Add(graph.ID("a"), 1)
	g.Add(graph.ID("a", "b"), 2)
	g.Add(graph.ID("a", "c"), 3)
	g.Connect(graph.ID("a"), graph.ID("a", "b"))
	g.Connect(graph.ID("a"), graph.ID("a", "c"))

	valuePrinter := prettyprinters.New(graphviz.New(graphviz.DefaultOptions(), graphviz.DefaultProvider()))
	valueDotCode, _ := valuePrinter.Show(g)
	fmt.Println(valueDotCode)

	// Output:
	// digraph {
	// splines = "spline";
	// rankdir = "LR";
	//
	// "1" [ label="1"];
	// "2" [ label="2"];
	// "3" [ label="3"];
	// "1" -> "2" [ label=""];
	// "1" -> "3" [ label=""];
	// }
}

func ExampleShowGraphWithIDProvider() {
	g := graph.New()
	g.Add(graph.ID("a"), 1)
	g.Add(graph.ID("a", "b"), 2)
	g.Add(graph.ID("a", "c"), 3)
	g.Connect(graph.ID("a"), graph.ID("a", "b"))
	g.Connect(graph.ID("a"), graph.ID("a", "c"))

	namePrinter := prettyprinters.New(graphviz.New(graphviz.DefaultOptions(), graphviz.IDProvider()))
	nameDotCode, _ := namePrinter.Show(g)
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
	dotCode, err := printer.Show(g)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(dotCode)
}
