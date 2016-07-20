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
	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/prettyprinters"
)

var (
	//BadOptionsError is returned when a GraphvizOptionMap contains invalid keys
	//or values, or is missing required values.
	BadOptionsError = error.Error

	//BadGraphError is returned when a provided graph.Graph is not valid
	BadGraphError = error.Error
)

type GraphvizOptions map[string]string
type VertexPrinter func(interface{}) (string, error)

type GraphvizPrinter struct {
	prettyprinters.DigraphPrettyPrinter
	optsMap  GraphvizOptions
	vPrinter VertexPrinter
}

var defaultOptions GraphvizOptions = map[string]string{
	"splines": "curved",
	"rankdir": "LR",
}

func DefaultOptions() GraphvizOptions {
	return defaultOptions
}

func checkOptions(opts GraphvizOptions) {
}

func NewPrinter(opts GraphvizOptions, printer VertexPrinter) (*GraphvizPrinter, error) {
	opts = mergeDefaultOptions(opts)
	return &GraphvizPrinter{
		optsMap:  opts,
		vPrinter: printer,
	}, nil
}

func (g *GraphvizPrinter) Options() GraphvizOptions {
	return g.optsMap
}

/* FIXME: Stubs*/
func (*GraphvizPrinter) StartPP(*graph.Graph) (string, error) {
	return "", nil
}

func (*GraphvizPrinter) FinishPP(*graph.Graph) (string, error) {
	return "", nil
}

func (*GraphvizPrinter) StartSubgraph(*graph.Graph) (string, error) {
	return "", nil
}

func (*GraphvizPrinter) FinishSubgraph(*graph.Graph) (string, error) {
	return "", nil
}

func (*GraphvizPrinter) DrawNode(*graph.Graph, interface{}, func(*graph.Graph)) (string, error) {
	return "", nil
}

// Utility Functions

//mergeDefaultOptions iterates over an option map and copies the defaults for
//any missing entries.
func mergeDefaultOptions(opts GraphvizOptions) GraphvizOptions {
	for k, v := range DefaultOptions() {
		if _, ok := opts[k]; !ok {
			opts[k] = v
		}
	}
	return opts
}
