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
	"github.com/stretchr/testify/mock"
)

var (
	emptyOptsMap = map[string]string{}
	emptyGraph   = graph.New()
	stubID       = "stub"
)

// NewPrinter
func Test_NewPrinter_WhenMissingOptions_UsesDefaultOptions(t *testing.T) {
	provider := defaultMockProvider()
	printer := graphviz.NewPrinter(emptyOptsMap, provider)
	actual := printer.Options()
	expected := graphviz.DefaultOptions()
	assert.Equal(t, actual, expected)
}

func Test_NewPrinter_WhenProvidedOptions_UsesProvidedOptions(t *testing.T) {
	opts := map[string]string{"rankdir": "TB"} // not the default
	printer := graphviz.NewPrinter(opts, defaultMockProvider())
	setOpts := printer.Options()
	rankdir := setOpts["rankdir"]
	assert.Equal(t, "TB", rankdir)
}

// Draw Node
func Test_DrawNode_WhenRenderFunction_CallsRenderFunction(t *testing.T) {
	provider := defaultMockProvider()
	printer := graphviz.NewPrinter(emptyOptsMap, provider)
	printer.DrawNode(emptyGraph, stubID, stubFlagFunc)
	provider.AssertCalled(t, "VertexGetId", mock.Anything)
}

func Test_DrawNode_CallsShouldMark(t *testing.T) {
	provider := defaultMockProvider()
	g := testGraph()
	printer := graphviz.NewPrinter(emptyOptsMap, provider)
	printer.DrawNode(g, stubID, stubFlagFunc)
	provider.AssertCalled(t, "SubgraphMarker", mock.Anything)
}

func Test_DrawNode_WhenMarkerReturnsStart_CallsSubgraphMarkTrue(t *testing.T) {
	provider := new(MockPrintProvider)
	provider.On("VertexGetId", mock.Anything).Return("", nil)
	provider.On("VertexGetLabel", mock.Anything).Return("", nil)
	provider.On("SubgraphMarker", mock.Anything).Return(graphviz.SubgraphMarkerStart)
	calledFlag := false
	calledWith := false
	subgraphMarker := func(_ string, c bool) {
		calledFlag = true
		calledWith = c
	}
	g := testGraph()
	printer := graphviz.NewPrinter(emptyOptsMap, provider)
	printer.DrawNode(g, stubID, subgraphMarker)
	assert.True(t, calledFlag)
	assert.True(t, calledWith)
}

func Test_DrawNode_WhenMarkerReturnsEnd_CallSubgraphMarkFalse(t *testing.T) {
	provider := new(MockPrintProvider)
	provider.On("VertexGetId", mock.Anything).Return("", nil)
	provider.On("VertexGetLabel", mock.Anything).Return("", nil)
	provider.On("SubgraphMarker", mock.Anything).Return(graphviz.SubgraphMarkerEnd)
	calledFlag := false
	calledWith := true
	subgraphMarker := func(_ string, c bool) {
		calledFlag = true
		calledWith = c
	}
	g := testGraph()
	printer := graphviz.NewPrinter(emptyOptsMap, provider)
	printer.DrawNode(g, stubID, subgraphMarker)
	assert.True(t, calledFlag)
	assert.False(t, calledWith)
}

func Test_DrawNode_WhenMarkerReturnsNop_DoesNotCallSubgraphMark(t *testing.T) {
	provider := new(MockPrintProvider)
	provider.On("VertexGetId", mock.Anything).Return("", nil)
	provider.On("VertexGetLabel", mock.Anything).Return("", nil)
	provider.On("SubgraphMarker", mock.Anything).Return(graphviz.SubgraphMarkerNop)
	calledFlag := false
	subgraphMarker := func(_ string, c bool) {
		calledFlag = true
	}
	g := testGraph()
	printer := graphviz.NewPrinter(emptyOptsMap, provider)
	printer.DrawNode(g, stubID, subgraphMarker)
	assert.False(t, calledFlag)
}

func Test_DrawNode_SetsNodeNameToVertexId(t *testing.T) {
	vertexID := "testID"
	provider := new(MockPrintProvider)
	provider.On("VertexGetId", mock.Anything).Return(vertexID, nil)
	provider.On("VertexGetLabel", mock.Anything).Return("", nil)
	provider.On("SubgraphMarker", mock.Anything).Return(graphviz.SubgraphMarkerNop)
	g := graph.New()
	g.Add("test", nil)
	printer := graphviz.NewPrinter(emptyOptsMap, provider)
	dotCode, _ := printer.DrawNode(g, "test", stubFlagFunc)
	actual := getDotNodeID(dotCode)
	assert.Equal(t, actual, vertexID)
}

func Test_DrawNode_WhenVertexIDReturnsError_ReturnsError(t *testing.T) {
	err := errors.New("test error")
	provider := new(MockPrintProvider)
	provider.On("VertexGetId", mock.Anything).Return("", err)
	provider.On("VertexGetLabel", mock.Anything).Return("", nil)
	provider.On("SubgraphMarker", mock.Anything).Return(graphviz.SubgraphMarkerNop)
	g := graph.New()
	g.Add("test", nil)
	printer := graphviz.NewPrinter(emptyOptsMap, provider)
	_, actualErr := printer.DrawNode(g, "test", stubFlagFunc)
	assert.Equal(t, actualErr, err)
}

func Test_DrawNode_SetsLabelToVertexLabel(t *testing.T) {
	vertexLabel := "test label"
	provider := new(MockPrintProvider)
	provider.On("VertexGetId", mock.Anything).Return("", nil)
	provider.On("VertexGetLabel", mock.Anything).Return(vertexLabel, nil)
	provider.On("SubgraphMarker", mock.Anything).Return(graphviz.SubgraphMarkerNop)
	g := graph.New()
	g.Add("test", nil)
	printer := graphviz.NewPrinter(emptyOptsMap, provider)
	dotCode, _ := printer.DrawNode(g, "test", stubFlagFunc)
	actual := getDotNodeLabel(dotCode)
	assert.Equal(t, actual, vertexLabel)
}

func Test_DrawNode_WhenVertexLabelReturnsError_ReturnsError(t *testing.T) {
	err := errors.New("test error")
	provider := new(MockPrintProvider)
	provider.On("VertexGetId", mock.Anything).Return("", nil)
	provider.On("VertexGetLabel", mock.Anything).Return("", err)
	provider.On("SubgraphMarker", mock.Anything).Return(graphviz.SubgraphMarkerNop)
	g := graph.New()
	g.Add("test", nil)
	printer := graphviz.NewPrinter(emptyOptsMap, provider)
	_, actualErr := printer.DrawNode(g, "test", stubFlagFunc)
	assert.Equal(t, actualErr, err)
}

// DrawEdge

// Stubs / Utility Functions

type MockPrintProvider struct {
	mock.Mock
}

func (m *MockPrintProvider) VertexGetId(i interface{}) (string, error) {
	args := m.Called(i)
	return args.String(0), args.Error(1)
}

func (m *MockPrintProvider) VertexGetLabel(i interface{}) (string, error) {
	args := m.Called(i)
	return args.String(0), args.Error(1)
}

func (m *MockPrintProvider) SubgraphMarker(i interface{}) graphviz.SubgraphMarkerKey {
	args := m.Called(i)
	return args.Get(0).(graphviz.SubgraphMarkerKey)
}

func defaultMockProvider() *MockPrintProvider {
	m := new(MockPrintProvider)
	m.On("VertexGetId", mock.Anything).Return("id1", nil)
	m.On("VertexGetLabel", mock.Anything).Return("label1", nil)
	m.On("SubgraphMarker", mock.Anything).Return(graphviz.SubgraphMarkerNop)
	return m
}

func stubFlagFunc(_ string, _ bool) {
}

func stubPrinter(interface{}) (string, error) {
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
