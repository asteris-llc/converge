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

func main() {
	log.SetOutput(ioutil.Discard)
	g := makeGraph()
	printer := prettyprinters.New(g, graphviz.New(graphviz.DefaultOptions(), graphviz.DefaultProvider()))
	dotCode, _ := printer.Show()
	fmt.Println(dotCode)
}
