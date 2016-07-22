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

package main

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/prettyprinters"
	"github.com/asteris-llc/converge/prettyprinters/graphviz"
)

func makeGraph() *graph.Graph {
	g := graph.New()
	g.Add(graph.ID("a"), 1)
	g.Add(graph.ID("a", "b"), 2)
	g.Add(graph.ID("a", "c"), 3)
	g.Connect(graph.ID("a"), graph.ID("a", "b"))
	g.Connect(graph.ID("a"), graph.ID("a", "c"))
	return g
}

func showGraphWithValues(g *graph.Graph) {
	valuePrinter := prettyprinters.New(g, graphviz.New(graphviz.DefaultOptions(), graphviz.DefaultProvider()))
	valueDotCode, _ := valuePrinter.Show()
	fmt.Println("With default value provider")
	fmt.Println(valueDotCode)
}

func showGraphWithIDs(g *graph.Graph) {
	namePrinter := prettyprinters.New(g, graphviz.New(graphviz.DefaultOptions(), graphviz.IDProvider()))
	nameDotCode, _ := namePrinter.Show()
	fmt.Println("With default ID provider")
	fmt.Println(nameDotCode)
}

func main() {
	log.SetOutput(ioutil.Discard)
	g := makeGraph()
	showGraphWithValues(g)
	showGraphWithIDs(g)
}
