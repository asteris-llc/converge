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
	emptyGraph = graph.New()
	stubID     = "stub"
)

// TODO: Create a mock factory or something to reduce some of the boilerplate
// code where we have to set default bindings on the mock that we don't care
// about, because we can't just override a binding after calling
// defaultMockProvider().  Probably take a bunch of functions as args and just
// set them to the default implementation if they are nil, the
// defaultMockProvider() would just be the case where every param is nil.

func Test_DrawNode_WhenRenderFunction_CallsRenderFunction(t *testing.T) {
	provider := defaultMockProvider()
	printer := graphviz.New(graphviz.DefaultOptions(), provider)
	printer.DrawNode(emptyGraph, stubID, stubFlagFunc)
	provider.AssertCalled(t, "VertexGetID", mock.Anything)
}

func Test_DrawNode_CallsShouldMark(t *testing.T) {
	provider := defaultMockProvider()
	g := testGraph()
	printer := graphviz.New(graphviz.DefaultOptions(), provider)
	printer.DrawNode(g, stubID, stubFlagFunc)
	provider.AssertCalled(t, "SubgraphMarker", mock.Anything)
}

func Test_DrawNode_WhenMarkerReturnsStart_CallsSubgraphMarkTrue(t *testing.T) {
	provider := new(MockPrintProvider)
	provider.On("VertexGetID", mock.Anything).Return("", nil)
	provider.On("VertexGetLabel", mock.Anything).Return("", nil)
	provider.On("SubgraphMarker", mock.Anything).Return(graphviz.SubgraphMarkerStart)
	provider.On("VertexGetProperties", mock.Anything).Return(make(graphviz.PropertySet))
	calledFlag := false
	calledWith := false
	subgraphMarker := func(_ string, c bool) {
		calledFlag = true
		calledWith = c
	}
	g := testGraph()
	printer := graphviz.New(graphviz.DefaultOptions(), provider)
	printer.DrawNode(g, stubID, subgraphMarker)
	assert.True(t, calledFlag)
	assert.True(t, calledWith)
}

func Test_DrawNode_WhenMarkerReturnsEnd_CallSubgraphMarkFalse(t *testing.T) {
	provider := new(MockPrintProvider)
	provider.On("VertexGetID", mock.Anything).Return("", nil)
	provider.On("VertexGetLabel", mock.Anything).Return("", nil)
	provider.On("SubgraphMarker", mock.Anything).Return(graphviz.SubgraphMarkerEnd)
	provider.On("VertexGetProperties", mock.Anything).Return(make(graphviz.PropertySet))
	calledFlag := false
	calledWith := true
	subgraphMarker := func(_ string, c bool) {
		calledFlag = true
		calledWith = c
	}
	g := testGraph()
	printer := graphviz.New(graphviz.DefaultOptions(), provider)
	printer.DrawNode(g, stubID, subgraphMarker)
	assert.True(t, calledFlag)
	assert.False(t, calledWith)
}

func Test_DrawNode_WhenMarkerReturnsNOP_DoesNotCallSubgraphMark(t *testing.T) {
	provider := new(MockPrintProvider)
	provider.On("VertexGetID", mock.Anything).Return("", nil)
	provider.On("VertexGetLabel", mock.Anything).Return("", nil)
	provider.On("SubgraphMarker", mock.Anything).Return(graphviz.SubgraphMarkerNOP)
	provider.On("VertexGetProperties", mock.Anything).Return(make(graphviz.PropertySet))
	calledFlag := false
	subgraphMarker := func(_ string, c bool) {
		calledFlag = true
	}
	g := testGraph()
	printer := graphviz.New(graphviz.DefaultOptions(), provider)
	printer.DrawNode(g, stubID, subgraphMarker)
	assert.False(t, calledFlag)
}

func Test_DrawNode_SetsNodeNameToVertexID(t *testing.T) {
	vertexID := "testID"
	provider := new(MockPrintProvider)
	provider.On("VertexGetID", mock.Anything).Return(vertexID, nil)
	provider.On("VertexGetLabel", mock.Anything).Return("", nil)
	provider.On("SubgraphMarker", mock.Anything).Return(graphviz.SubgraphMarkerNOP)
	provider.On("VertexGetProperties", mock.Anything).Return(make(graphviz.PropertySet))
	g := graph.New()
	g.Add("test", nil)
	printer := graphviz.New(graphviz.DefaultOptions(), provider)
	dotCode, _ := printer.DrawNode(g, "test", stubFlagFunc)
	actual := getDotNodeID(dotCode)
	assert.Equal(t, vertexID, actual)
}

func Test_DrawNode_WhenVertexIDReturnsError_ReturnsError(t *testing.T) {
	err := errors.New("test error")
	provider := new(MockPrintProvider)
	provider.On("VertexGetID", mock.Anything).Return("", err)
	provider.On("VertexGetLabel", mock.Anything).Return("", nil)
	provider.On("SubgraphMarker", mock.Anything).Return(graphviz.SubgraphMarkerNOP)
	provider.On("VertexGetProperties", mock.Anything).Return(make(graphviz.PropertySet))
	g := graph.New()
	g.Add("test", nil)
	printer := graphviz.New(graphviz.DefaultOptions(), provider)
	_, actualErr := printer.DrawNode(g, "test", stubFlagFunc)
	assert.Equal(t, err, actualErr)
}

func Test_DrawNode_SetsLabelToVertexLabel(t *testing.T) {
	vertexLabel := "test label"
	provider := new(MockPrintProvider)
	provider.On("VertexGetID", mock.Anything).Return("", nil)
	provider.On("VertexGetLabel", mock.Anything).Return(vertexLabel, nil)
	provider.On("SubgraphMarker", mock.Anything).Return(graphviz.SubgraphMarkerNOP)
	provider.On("VertexGetProperties", mock.Anything).Return(make(graphviz.PropertySet))
	g := graph.New()
	g.Add("test", nil)
	printer := graphviz.New(graphviz.DefaultOptions(), provider)
	dotCode, _ := printer.DrawNode(g, "test", stubFlagFunc)
	actual := getDotNodeLabel(dotCode)
	assert.Equal(t, vertexLabel, actual)
}

func Test_DrawNode_WhenVertexLabelReturnsError_ReturnsError(t *testing.T) {
	err := errors.New("test error")
	provider := new(MockPrintProvider)
	provider.On("VertexGetID", mock.Anything).Return("", nil)
	provider.On("VertexGetLabel", mock.Anything).Return("", err)
	provider.On("SubgraphMarker", mock.Anything).Return(graphviz.SubgraphMarkerNOP)
	provider.On("VertexGetProperties", mock.Anything).Return(make(graphviz.PropertySet))
	g := graph.New()
	g.Add("test", nil)
	printer := graphviz.New(graphviz.DefaultOptions(), provider)
	_, actualErr := printer.DrawNode(g, "test", stubFlagFunc)
	assert.Equal(t, err, actualErr)
}

func Test_DrawNode_WhenAdditionalAttributes_AddsAttributesTo(t *testing.T) {
	expectedAttrs := graphviz.PropertySet{
		"key1": "val1",
		"key2": "val2",
	}
	provider := new(MockPrintProvider)
	provider.On("VertexGetID", mock.Anything).Return("", nil)
	provider.On("VertexGetLabel", mock.Anything).Return("", nil)
	provider.On("SubgraphMarker", mock.Anything).Return(graphviz.SubgraphMarkerNOP)
	provider.On("VertexGetProperties", mock.Anything).Return(expectedAttrs)
	g := graph.New()
	g.Add("test", nil)
	printer := graphviz.New(graphviz.DefaultOptions(), provider)
	dotCode, _ := printer.DrawNode(g, "test", stubFlagFunc)
	actualAttrs := getDotAttributes(dotCode)
	// NB: compareAttrMap does not commute.  As written this will only assert that
	// the found attr map contains at a minimum the expected attributes. This is
	// desireable for this test since we do not want to make assumptions about any
	// additional attributes that should be included (e.g. label), just that we
	// also have the ones that were specified
	assert.True(t, compareAttrMap(expectedAttrs, actualAttrs))
}

func Test_DrawEdge_GetsIDForEachNode(t *testing.T) {
	provider := defaultMockProvider()
	printer := graphviz.New(graphviz.DefaultOptions(), provider)
	printer.DrawEdge(edgeTestGraph(), "A", "B")
	provider.AssertCalled(t, "VertexGetID", "A")
	provider.AssertCalled(t, "VertexGetID", "B")
}

func Test_DrawEdge_SetsSourceAndDestVertexToSourceAndDest(t *testing.T) {
	provider := new(MockPrintProvider)
	provider.On("VertexGetID", "A").Return("A", nil)
	provider.On("VertexGetID", "B").Return("B", nil)
	provider.On("EdgeGetProperties", mock.Anything, mock.Anything).Return(make(graphviz.PropertySet))
	provider.On("EdgeGetLabel", mock.Anything, mock.Anything).Return("", nil)
	printer := graphviz.New(graphviz.DefaultOptions(), provider)
	dotSource, _ := printer.DrawEdge(edgeTestGraph(), "A", "B")
	sourceVertex, destVertex := parseDotEdge(dotSource)
	assert.Equal(t, "A", sourceVertex)
	assert.Equal(t, "B", destVertex)
}

func Test_DrawEdge_WhenVertexIDReturnsError_ReturnsError(t *testing.T) {
	err := errors.New("test error")
	provider := new(MockPrintProvider)
	provider.On("VertexGetID", mock.Anything).Return("", err)
	provider.On("EdgeGetProperties", mock.Anything, mock.Anything).Return(make(graphviz.PropertySet))
	provider.On("EdgeGetLabel", mock.Anything, mock.Anything).Return("", nil)
	printer := graphviz.New(graphviz.DefaultOptions(), provider)
	_, actualErr := printer.DrawEdge(edgeTestGraph(), "test1", "test2")
	assert.Equal(t, actualErr, err)
}

func Test_DrawEdge_SetsLabelToEdgeLabel(t *testing.T) {
	edgeLabel := "test label"
	provider := new(MockPrintProvider)
	provider.On("VertexGetID", mock.Anything).Return("test", nil)
	provider.On("EdgeGetLabel", mock.Anything, mock.Anything).Return(edgeLabel, nil)
	provider.On("EdgeGetProperties", mock.Anything, mock.Anything).Return(make(graphviz.PropertySet))
	printer := graphviz.New(graphviz.DefaultOptions(), provider)
	dotCode, _ := printer.DrawEdge(edgeTestGraph(), "test", "test")
	actual := getDotNodeLabel(dotCode)
	assert.Equal(t, edgeLabel, actual)
}

func Test_DrawEdge_WhenEdgeLabelReturnsError_ReturnsError(t *testing.T) {
	err := errors.New("test error")
	provider := new(MockPrintProvider)
	provider.On("VertexGetID", mock.Anything).Return("", nil)
	provider.On("EdgeGetLabel", mock.Anything, mock.Anything).Return("", err)
	provider.On("EdgeGetProperties", mock.Anything, mock.Anything).Return(make(graphviz.PropertySet))
	printer := graphviz.New(graphviz.DefaultOptions(), provider)
	_, actualErr := printer.DrawEdge(edgeTestGraph(), "test", "test")
	assert.Equal(t, err, actualErr)
}

func Test_DrawEdge_WhenAdditionalAttributes_AddsAttributesToEdge(t *testing.T) {
	expectedAttrs := graphviz.PropertySet{
		"key1": "val1",
		"key2": "val2",
	}
	provider := new(MockPrintProvider)
	provider.On("VertexGetID", mock.Anything).Return("", nil)
	provider.On("EdgeGetLabel", mock.Anything, mock.Anything).Return("", nil)
	provider.On("EdgeGetProperties", mock.Anything, mock.Anything).Return(expectedAttrs)
	printer := graphviz.New(graphviz.DefaultOptions(), provider)
	dotCode, _ := printer.DrawEdge(edgeTestGraph(), "test", "test")
	actualAttrs := getDotAttributes(dotCode)
	// NB: compareAttrMap does not commute.  As written this will only assert that
	// the found attr map contains at a minimum the expected attributes. This is
	// desireable for this test since we do not want to make assumptions about any
	// additional attributes that should be included (e.g. label), just that we
	// also have the ones that were specified
	assert.True(t, compareAttrMap(expectedAttrs, actualAttrs))
}

func Test_StartNodeSection_ReturnsEmptyString(t *testing.T) {
	provider := defaultMockProvider()
	printer := graphviz.New(graphviz.DefaultOptions(), provider)
	expected := ""
	actual, err := printer.StartNodeSection(emptyGraph)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

/////////////////////////////////
// Stubs / Utility Functions

type MockPrintProvider struct {
	mock.Mock
}

func (m *MockPrintProvider) VertexGetID(i interface{}) (string, error) {
	args := m.Called(i)
	return args.String(0), args.Error(1)
}

func (m *MockPrintProvider) VertexGetLabel(i interface{}) (string, error) {
	args := m.Called(i)
	return args.String(0), args.Error(1)
}

func (m *MockPrintProvider) VertexGetProperties(i interface{}) graphviz.PropertySet {
	args := m.Called(i)
	return args.Get(0).(graphviz.PropertySet)
}

func (m *MockPrintProvider) SubgraphMarker(i interface{}) graphviz.SubgraphMarkerKey {
	args := m.Called(i)
	return args.Get(0).(graphviz.SubgraphMarkerKey)
}

func (m *MockPrintProvider) EdgeGetLabel(i, j interface{}) (string, error) {
	args := m.Called(i, j)
	return args.String(0), args.Error(1)
}

func (m *MockPrintProvider) EdgeGetProperties(i, j interface{}) graphviz.PropertySet {
	args := m.Called(i, j)
	return args.Get(0).(graphviz.PropertySet)
}

func defaultMockProvider() *MockPrintProvider {
	m := new(MockPrintProvider)
	m.On("VertexGetID", mock.Anything).Return("id1", nil)
	m.On("VertexGetLabel", mock.Anything).Return("label1", nil)
	m.On("VertexGetProperties", mock.Anything).Return(make(graphviz.PropertySet))
	m.On("EdgeGetLabel", mock.Anything, mock.Anything).Return("label1", nil)
	m.On("EdgeGetProperties", mock.Anything, mock.Anything).Return(make(graphviz.PropertySet))
	m.On("SubgraphMarker", mock.Anything).Return(graphviz.SubgraphMarkerNOP)
	return m
}

func stubFlagFunc(_ string, _ bool) {
}

func stubPrinter(interface{}) (string, error) {
	return "", nil
}

func stubMarker(_ interface{}) graphviz.SubgraphMarkerKey {
	return graphviz.SubgraphMarkerNOP
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

func getAttributeSubstr(s string) (string, bool) {
	start := strings.IndexRune(s, '[')
	end := strings.IndexRune(s, ']')
	if start == -1 || end == -1 {
		return "", false
	}
	return s[start+1 : end], true
}

func stripQuotes(s string) string {
	if s[0] == '"' || s[0] == '\'' {
		return s[1 : len(s)-1]
	}
	return s
}

func get_kv(attr string) (string, string) {
	pair := strings.Split(attr, "=")
	key := stripQuotes(strings.TrimSpace(pair[0]))
	val := stripQuotes(strings.TrimSpace(pair[1]))
	return key, val
}

func getDotAttributes(s string) map[string]string {
	results := make(map[string]string)
	attributes, found := getAttributeSubstr(s)

	if !found {
		return results
	}

	attributePairs := strings.Split(attributes, ",")

	for pair := range attributePairs {
		key, value := get_kv(attributePairs[pair])
		results[key] = value
	}
	return results
}

func parseDotEdge(e string) (string, string) {
	var dest string
	ef := strings.Split(strings.TrimSpace(e), "->")
	source := stripQuotes(strings.TrimSpace(ef[0]))
	destRaw := strings.TrimSpace(ef[1])
	quoteWrapper := destRaw[0]
	if quoteWrapper == '"' || quoteWrapper == '\'' {
		destRaw = destRaw[1:len(destRaw)]
		idx := strings.IndexByte(destRaw, quoteWrapper)
		dest = destRaw[0:idx]
	} else {
		dest = strings.TrimSpace(destRaw)
		idx := strings.IndexAny(dest, " \t")
		dest = dest[0:idx]
	}
	return source, dest
}

func compareAttrMap(a, b map[string]string) bool {
	for key, refVal := range a {
		foundVal, found := b[key]
		if !found {
			fmt.Printf("key %s missing in dest\n", key)
			return false
		}
		if refVal != foundVal {
			fmt.Printf("mismatched values: refVal = \"%s\", foundVal = \"%s\"\n", refVal, foundVal)
		}
	}
	return true
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

func edgeTestGraph() *graph.Graph {
	g := graph.New()
	g.Add("A", "A")
	g.Add("B", "B")
	g.Add("C", "C")
	g.Connect("A", "B")
	g.Connect("A", "C")
	return g
}
