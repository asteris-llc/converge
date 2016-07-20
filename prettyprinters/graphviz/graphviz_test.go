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
	"fmt"
	"strings"
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

// NewPrinter
func Test_NewPrinter_WhenMissingOptions_UsesDefaultOptions(t *testing.T) {

	printer := graphviz.NewPrinter(emptyOptsMap, nil, nil, nil)
	actual := printer.Options()
	expected := graphviz.DefaultOptions()
	assert.Equal(t, actual, expected)
}

func Test_NewPrinter_WhenProvidedOptions_UsesProvidedOptions(t *testing.T) {
	opts := map[string]string{"rankdir": "TB"} // not the default
	printer := graphviz.NewPrinter(opts, nil, nil, nil)
	setOpts := printer.Options()
	rankdir := setOpts["rankdir"]
	assert.Equal(t, "TB", rankdir)
}

// Draw Node
func Test_DrawNode_WhenRenderFunction_CallsRenderFunction(t *testing.T) {
	calledFlag := false
	idF := func(_ interface{}) (string, error) {
		calledFlag = true
		return "", nil
	}
	printer := graphviz.NewPrinter(emptyOptsMap, idF, nil, stubMarker)
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
	printer := graphviz.NewPrinter(emptyOptsMap, nil, nil, shouldMark)
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
	printer := graphviz.NewPrinter(emptyOptsMap, nil, nil, shouldMark)
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
	printer := graphviz.NewPrinter(emptyOptsMap, nil, nil, shouldMark)
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
	printer := graphviz.NewPrinter(emptyOptsMap, nil, nil, shouldMark)
	printer.DrawNode(g, stubID, subgraphMarker)
	assert.False(t, calledFlag)
}

func Test_DrawNode_WhenIDFunc_SetsIDToIDFunc(t *testing.T) {
	nodeID := "A"
	idF := func(interface{}) (string, error) {
		return nodeID, nil
	}
	g := graph.New()
	g.Add("test", nil)
	printer := graphviz.NewPrinter(emptyOptsMap, idF, nil, nil)
	dotCode, _ := printer.DrawNode(g, "test", stubFlagFunc)
	actual := getDotNodeID(dotCode)
	assert.Equal(t, actual, nodeID)
}

func Test_DrawNode_WhenNoIDFunc_SetsIDStructValue(t *testing.T) {
	nodeVal := "A"
	expected := fmt.Sprintf("%x", nodeVal)
	g := graph.New()
	g.Add("test", nodeVal)
	printer := graphviz.NewPrinter(emptyOptsMap, nil, nil, nil)
	dotCode, _ := printer.DrawNode(g, "test", stubFlagFunc)
	actual := getDotNodeID(dotCode)
	assert.Equal(t, actual, expected)
}

func Test_DrawNode_WhenVertexPrinterReturnsError_ReturnsError(t *testing.T) {
	expectedError := errors.New("test error")
	idF := func(interface{}) (string, error) {
		return "", expectedError
	}
	printer := graphviz.NewPrinter(emptyOptsMap, idF, nil, nil)
	g := graph.New()
	g.Add("test", nil)
	_, actualError := printer.DrawNode(g, "test", stubFlagFunc)
	assert.Equal(t, actualError, expectedError)
}

// DrawEdge

// Stubs / Utility Functions
func stubFlagFunc(string, bool) {
}

func stubPrinter(_ interface{}) (string, error) {
	return "", nil
}

func stubMarker(_ interface{}) graphviz.SubgraphMarkerKey {
	return graphviz.SubgraphMarkerNop
}

func getDotNodeID(s string) string {
	trimmed := strings.TrimSpace(s)
	firstChar := trimmed[0]
	if firstChar == '\'' || firstChar == '"' {
		sep := fmt.Sprintf("%c", firstChar)
		return strings.Split(s, sep)[1]
	}
	return strings.Split(trimmed, " ")[0]
}

func getDotNodeLabel(s string) string {
	labelSplit := strings.Split(s, "label=")
	if len(labelSplit) < 2 {
		return ""
	}
	labelPart := labelSplit[1]
	firstChar := labelPart[0]
	if firstChar == '\'' || firstChar == '"' {
		sep := fmt.Sprintf("%c", firstChar)
		return strings.Split(labelPart, sep)[1]
	}
	return strings.Split(labelPart, " ")[0]
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
