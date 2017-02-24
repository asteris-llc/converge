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

package resource_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/file/content"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExportedFields tests things related to exported fields
func TestExportedFields(t *testing.T) {
	t.Parallel()
	t.Run("exported-fields", func(t *testing.T) {
		t.Parallel()
		t.Run("when-value", func(t *testing.T) {
			t.Parallel()
			expected := []string{"a", "c"}
			actual, err := resource.ExportedFields(TestOuterStruct{})
			require.NoError(t, err)
			assert.Equal(t, expected, fieldNames(actual))
		})
		t.Run("when-pointer", func(t *testing.T) {
			t.Parallel()
			expected := []string{"a", "c"}
			actual, err := resource.ExportedFields(&TestOuterStruct{})
			require.NoError(t, err)
			assert.Equal(t, expected, fieldNames(actual))
		})
		t.Run("when-nil", func(t *testing.T) {
			t.Parallel()
			_, err := resource.ExportedFields(nil)
			assert.Error(t, err)
		})
		t.Run("when-embedded", func(t *testing.T) {
			t.Parallel()
			t.Run("non-overlapping", func(t *testing.T) {
				t.Parallel()
				expected := []string{"a", "b", "x"}
				actual, err := resource.ExportedFields(&TestEmbeddingStruct{})
				require.NoError(t, err)
				assert.Equal(t, expected, fieldNames(actual))
			})
			t.Run("overlapping", func(t *testing.T) {
				t.Parallel()
				expected := []string{"a", "c", "x", "b", "TestEmbeddedStruct.x"}
				actual, err := resource.ExportedFields(&TestEmbeddingOverlap{})
				require.NoError(t, err)
				assert.Equal(t, expected, fieldNames(actual))
			})
		})
	})
	t.Run("reference-fields", func(t *testing.T) {
		t.Parallel()
		expected := []string{"a", "c"}
		actual, err := resource.ExportedFields(TestOuterStruct{})
		require.NoError(t, err)
		assert.Equal(t, expected, lookupNames(actual))
		require.True(t, len(actual) > 1)
		val1 := actual[1]
		assert.Equal(t, "B", val1.FieldName)
		assert.Equal(t, "c", val1.ReferenceName)
	})
	t.Run("values", func(t *testing.T) {
		t.Parallel()
		input := newOuterStruct(3, 4, 5, 6)
		actual, err := resource.ExportedFields(input)
		require.NoError(t, err)
		require.True(t, len(actual) > 1)
		assert.Equal(t, 3, actual[0].Value.Interface().(int))
		assert.Equal(t, 4, actual[1].Value.Interface().(int))
	})

	t.Run("re-export-as-fields", func(t *testing.T) {
		t.Parallel()
		expected := []string{"embedded.b", "embedded.x"}
		input := &TestReexport{}
		actual, err := resource.ExportedFields(input)
		require.NoError(t, err)
		assert.Equal(t, expected, fieldNames(actual))
	})

	t.Run("re-export-with-non-nil-pointer", func(t *testing.T) {
		t.Parallel()
		expected := []string{"embeddedptr.b", "embeddedptr.x"}
		input := newReExportPtr(1, 2, 3)
		actual, err := resource.ExportedFields(input)
		if err != nil {
			t.Log("error: ", err)
		}
		require.NoError(t, err)
		fmt.Println("actual = ", actual)
		assert.Equal(t, expected, fieldNames(actual))
	})

	t.Run("re-export-as-with-nil", func(t *testing.T) {
		t.Parallel()
		expected := []string{"notnilval.b", "notnilval.x"}
		input := newReExportNil(1, 2, 3)
		actual, err := resource.ExportedFields(input)
		if err != nil {
			t.Log("error: ", err)
		}
		require.NoError(t, err)
		fmt.Println("actual = ", actual)
		assert.Equal(t, expected, fieldNames(actual))
	})

}

// TestGenerateLookupMap tests lookup map generation
func TestGenerateLookupMap(t *testing.T) {
	var nilInt *int
	t.Parallel()
	testVals := []*resource.ExportedField{
		mkExportedField("a", "a", "a"),
		mkExportedField("b", "b", "b"),
		mkExportedField("c", "x", "c"),
		mkExportedField("d", "Thing.x", "d"),
		mkExportedField("e", "e", nilInt),
	}
	results, err := resource.GenerateLookupMap(testVals)
	require.NoError(t, err)
	assert.Equal(t, results["a"], "a")
	assert.Equal(t, results["b"], "b")
	assert.Equal(t, results["x"], "c")
	assert.Equal(t, results["Thing.x"], "d")
	assert.Nil(t, results["e"])
	_, found := results["c"]
	assert.False(t, found)
}

// TestLookupMapFromStiruct tests lookup map generation from a struct
func TestLookupMapFromStruct(t *testing.T) {
	t.Parallel()
	testInput := &TestEmbeddingOverlap{
		TestEmbeddedStruct: TestEmbeddedStruct{
			A: 1,
			B: 2,
			X: 3,
		},
		A: 4,
		B: 5,
		X: 6,
	}
	refMap, err := resource.LookupMapFromStruct(testInput)
	require.NoError(t, err)
	assert.Equal(t, 4, refMap["a"].(int))
	assert.Equal(t, 5, refMap["c"].(int))
	assert.Equal(t, 6, refMap["x"].(int))
	assert.Equal(t, 2, refMap["b"].(int))
	assert.Equal(t, 3, refMap["TestEmbeddedStruct.x"].(int))
}

// TestWithFileContent is a simple integration test
func TestWithFileContent(t *testing.T) {
	content := &content.Content{
		Content:     "foo",
		Destination: "foo.txt",
	}
	results, err := resource.LookupMapFromInterface(content)
	require.NoError(t, err)
	t.Log(results)
}

// TestLookupMapFromInterface tests getting exported fields from an interface
func TestLookupMapFromInterface(t *testing.T) {
	t.Parallel()
	t.Run("when-pointer", func(t *testing.T) {
		t.Parallel()
		var asInterface TestInterface = &TestEmbeddingOverlap{}
		ifaceResult, err := resource.LookupMapFromInterface(asInterface)
		require.NoError(t, err)
		structResult, err := resource.LookupMapFromStruct(&TestEmbeddingOverlap{})
		require.NoError(t, err)
		assert.True(t, reflect.DeepEqual(ifaceResult, structResult))
	})
	t.Run("when-value", func(t *testing.T) {
		t.Parallel()
		var asInterface TestInterface = TestEmbeddingOverlap{}
		ifaceResult, err := resource.LookupMapFromInterface(asInterface)
		require.NoError(t, err)
		structResult, err := resource.LookupMapFromStruct(TestEmbeddingOverlap{})
		require.NoError(t, err)
		assert.True(t, reflect.DeepEqual(ifaceResult, structResult))
	})
	t.Run("when-nil", func(t *testing.T) {
		t.Parallel()
		var asInterface TestInterface
		_, err := resource.LookupMapFromInterface(asInterface)
		require.Error(t, err)
	})
}

func mkExportedField(name, refName string, val interface{}) *resource.ExportedField {
	return &resource.ExportedField{FieldName: name, ReferenceName: refName, Value: reflect.ValueOf(val)}
}

// TestInterface is for testing
type TestInterface interface{}

// TestOuterStruct is for testing
type TestOuterStruct struct {
	A int `export:"a"`
	B int `export:"c"`
	C int
	d int
}

// TestEmbeddingStruct is for testing
type TestEmbeddingStruct struct {
	TestEmbeddedStruct
	A int `export:"a"`
}

// TestEmbeddingOverlap is for testing
type TestEmbeddingOverlap struct {
	TestEmbeddedStruct
	A int `export:"a"`
	B int `export:"c"`
	X int `export:"x"`
}

// TestEmbeddedStruct is for testing
type TestEmbeddedStruct struct {
	A int
	B int `export:"b"`
	X int `export:"x"`
}

// TestReexport is for testing
type TestReexport struct {
	Embedded TestEmbeddedStruct `re-export-as:"embedded"`
}

type TestReExportPtr struct {
	Embedded *TestEmbeddedStruct `re-export-as:"embeddedptr"`
}

func newReExportPtr(a, b, c int) *TestReExportPtr {
	return &TestReExportPtr{Embedded: &TestEmbeddedStruct{a, b, c}}
}

// TestReExportNil tests re-exporting a nil pointer
type TestReExportNil struct {
	NotNil TestEmbeddedStruct  `re-export-as:"notnilval"`
	NilPtr *TestEmbeddedStruct `re-export-as:"nilval"`
}

func newReExportNil(a, b, c int) *TestReExportNil {
	return &TestReExportNil{NotNil: TestEmbeddedStruct{a, b, c}}
}

func newOuterStruct(a, b, c, d int) *TestOuterStruct {
	return &TestOuterStruct{A: a, B: b, C: c, d: d}
}

func fieldNames(in []*resource.ExportedField) (out []string) {
	for _, f := range in {
		out = append(out, f.ReferenceName)
	}
	return
}

func lookupNames(in []*resource.ExportedField) (out []string) {
	for _, f := range in {
		out = append(out, f.ReferenceName)
	}
	return
}
