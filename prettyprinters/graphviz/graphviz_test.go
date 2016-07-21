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
	provider.On("VertexGetProperties", mock.Anything).Return(make(graphviz.PropertySet))
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
	provider.On("VertexGetProperties", mock.Anything).Return(make(graphviz.PropertySet))
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
	provider.On("VertexGetProperties", mock.Anything).Return(make(graphviz.PropertySet))
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
	provider.On("VertexGetProperties", mock.Anything).Return(make(graphviz.PropertySet))
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
	provider.On("VertexGetProperties", mock.Anything).Return(make(graphviz.PropertySet))
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
	provider.On("VertexGetProperties", mock.Anything).Return(make(graphviz.PropertySet))
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
	provider.On("VertexGetProperties", mock.Anything).Return(make(graphviz.PropertySet))
	g := graph.New()
	g.Add("test", nil)
	printer := graphviz.NewPrinter(emptyOptsMap, provider)
	_, actualErr := printer.DrawNode(g, "test", stubFlagFunc)
	assert.Equal(t, actualErr, err)
}

func Test_DrawNode_WhenAdditionalAttributes_AddsAttributesTo(t *testing.T) {
	expectedAttrs := graphviz.PropertySet{
		"key1": "val1",
		"key2": "val2",
	}
	provider := new(MockPrintProvider)
	provider.On("VertexGetId", mock.Anything).Return("", nil)
	provider.On("VertexGetLabel", mock.Anything).Return("", nil)
	provider.On("SubgraphMarker", mock.Anything).Return(graphviz.SubgraphMarkerNop)
	provider.On("VertexGetProperties", mock.Anything).Return(expectedAttrs)
	g := graph.New()
	g.Add("test", nil)
	printer := graphviz.NewPrinter(emptyOptsMap, provider)
	dotCode, _ := printer.DrawNode(g, "test", stubFlagFunc)
	fmt.Printf("dotCode = %s\n", dotCode)
	actualAttrs := getDotAttributes(dotCode)
	fmt.Println("expected Attrs:")
	fmt.Println(expectedAttrs)
	fmt.Println("Actual Attrs:")
	fmt.Println(actualAttrs)
	// NB: compareAttrMap does not commute.  As written this will only assert that
	// the found attr map contains at a minimum the expected attributes. This is
	// desireable for this test since we do not want to make assumptions about any
	// additional attributes that should be included (e.g. label), just that we
	// also have the ones that were specified
	assert.True(t, compareAttrMap(expectedAttrs, actualAttrs))
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
	m.On("VertexGetId", mock.Anything).Return("id1", nil)
	m.On("VertexGetLabel", mock.Anything).Return("label1", nil)
	m.On("VertexGetProperties", mock.Anything).Return(make(graphviz.PropertySet))
	m.On("EdgeGetLabel", mock.Anything, mock.Anything).Return("label1", nil)
	m.On("EdgeGetProperties", mock.Anything, mock.Anything).Return(make(map[string]string))
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
