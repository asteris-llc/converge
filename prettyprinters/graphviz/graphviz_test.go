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

package graphviz_test

import (
	"errors"
	"testing"

	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/prettyprinters/graphviz"
	"github.com/stretchr/testify/assert"
)

var (
	emptyOptsMap = map[string]string{}
	emptyGraph   = graph.New()
	stubID       = "stub"
)

func Test_NewGraphvizPrinter_WhenMissingOptions_UsesDefaultOptions(t *testing.T) {

	printer := graphviz.NewPrinter(emptyOptsMap, stubPrinter, stubMarker)
	actual := printer.Options()
	expected := graphviz.DefaultOptions()
	assert.Equal(t, actual, expected)
}

func Test_NewGraphvizPrinter_WhenProvidedOptions_UsesProvidedOptions(t *testing.T) {
	opts := map[string]string{"rankdir": "TB"} // not the default
	printer := graphviz.NewPrinter(opts, stubPrinter, stubMarker)
	setOpts := printer.Options()
	rankdir := setOpts["rankdir"]
	assert.Equal(t, "TB", rankdir)
}

func Test_DrawNode_WhenRenderFunction_CallsRenderFunction(t *testing.T) {
	calledFlag := false
	renderF := func(_ interface{}) (string, error) {
		calledFlag = true
		return "", nil
	}
	printer := graphviz.NewPrinter(emptyOptsMap, renderF, stubMarker)
	printer.DrawNode(emptyGraph, stubID, stubFlagFunc)
	assert.True(t, calledFlag)
}

func Test_DrawNode_CallsShouldMark(t *testing.T) {
	calledFlag := false
	shouldMark := func(_ interface{}) graphviz.SubgraphMarkerKey {
		calledFlag = true
		return graphviz.SubgraphMarkerNop
	}
	g := testGraph()
	printer := graphviz.NewPrinter(emptyOptsMap, nil, shouldMark)
	printer.DrawNode(g, stubID, stubFlagFunc)
	assert.True(t, calledFlag)
}

func Test_DrawNode_WhenMarkerReturnsStart_CallsSubgraphMarTruek(t *testing.T) {
	calledFlag := false
	calledWith := false
	subgraphMarker := func(_ string, c bool) {
		calledFlag = true
		calledWith = c
	}
	shouldMark := func(_ interface{}) graphviz.SubgraphMarkerKey {
		return graphviz.SubgraphMarkerStart
	}
	g := testGraph()
	printer := graphviz.NewPrinter(emptyOptsMap, nil, shouldMark)
	printer.DrawNode(g, stubID, subgraphMarker)
	assert.True(t, calledFlag)
	assert.True(t, calledWith)
}

func Test_DrawNode_WhenMarkerReturnsEnd_CallSubgraphMarkFalse(t *testing.T) {
	calledFlag := false
	calledWith := true
	subgraphMarker := func(_ string, c bool) {
		calledFlag = true
		calledWith = c
	}
	shouldMark := func(_ interface{}) graphviz.SubgraphMarkerKey {
		return graphviz.SubgraphMarkerEnd
	}
	g := testGraph()
	printer := graphviz.NewPrinter(emptyOptsMap, nil, shouldMark)
	printer.DrawNode(g, stubID, subgraphMarker)
	assert.True(t, calledFlag)
	assert.False(t, calledWith)
}

func Test_DrawNode_WhenMarkerReturnsNop_DoesNotCallSubgraphMark(t *testing.T) {
	calledFlag := false
	subgraphMarker := func(_ string, c bool) {
		calledFlag = true
	}
	shouldMark := func(_ interface{}) graphviz.SubgraphMarkerKey {
		return graphviz.SubgraphMarkerNop
	}
	g := testGraph()
	printer := graphviz.NewPrinter(emptyOptsMap, nil, shouldMark)
	printer.DrawNode(g, stubID, subgraphMarker)
	assert.False(t, calledFlag)
}

func Test_DrawNode_WhenNoAttributes_ReturnsDotFormattedNode(t *testing.T) {
	expected := "\"A\";\n"
	vertexPrinter := func(interface{}) (string, error) {
		return "A", nil
	}
	g := graph.New()
	g.Add("test", nil)
	printer := graphviz.NewPrinter(emptyOptsMap, vertexPrinter, nil)
	actual, _ := printer.DrawNode(g, "test", stubFlagFunc)
	assert.Equal(t, actual, expected)
}

func Test_DrawNode_WhenVertexPrinterReturnsError_ReturnsError(t *testing.T) {
	expectedError := errors.New("test error")
	vertexPrinter := func(interface{}) (string, error) {
		return "", expectedError
	}
	printer := graphviz.NewPrinter(emptyOptsMap, vertexPrinter, nil)
	g := graph.New()
	g.Add("test", nil)
	_, actualError := printer.DrawNode(g, "test", stubFlagFunc)
	assert.Equal(t, actualError, expectedError)
}

func stubFlagFunc(string, bool) {
}

func stubPrinter(_ interface{}) (string, error) {
	return "", nil
}

func stubMarker(_ interface{}) graphviz.SubgraphMarkerKey {
	return graphviz.SubgraphMarkerNop
}

func testGraph() *graph.Graph {
	g := graph.New()
	g.Add("root", nil)
	g.Add("child1", nil)
	g.Add("child2", nil)
	g.Connect("root", "child1")
	g.Connect("root", "child2")
	return g
}
