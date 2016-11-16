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
	"testing"

	"github.com/asteris-llc/converge/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test things related to exported fields
func TestExportedFields(t *testing.T) {
	t.Run("exported-fields", func(t *testing.T) {
		t.Run("when-value", func(t *testing.T) {
			expected := []string{"A", "B"}
			actual, err := resource.ExportedFields(TestOuterStruct{})
			require.NoError(t, err)
			assert.Equal(t, expected, fieldNames(actual))
		})
		t.Run("when-pointer", func(t *testing.T) {
			expected := []string{"A", "B"}
			actual, err := resource.ExportedFields(&TestOuterStruct{})
			require.NoError(t, err)
			assert.Equal(t, expected, fieldNames(actual))
		})
		t.Run("when-nil", func(t *testing.T) {
			_, err := resource.ExportedFields(nil)
			assert.Error(t, err)
		})
		t.Run("when-embedded", func(t *testing.T) {
			expected := []string{"A", "B", "X"}
			actual, err := resource.ExportedFields(&TestEmbeddingStruct{})
			require.NoError(t, err)
			assert.Equal(t, expected, fieldNames(actual))
		})
	})
	t.Run("reference-fields", func(t *testing.T) {
		expected := []string{"a", "c"}
		actual, err := resource.ExportedFields(TestOuterStruct{})
		require.NoError(t, err)
		assert.Equal(t, expected, lookupNames(actual))
		val1 := actual[1]
		assert.Equal(t, "B", val1.FieldName)
		assert.Equal(t, "c", val1.ReferenceName)
	})
	t.Run("values", func(t *testing.T) {
		input := newOuterStruct(3, 4, 5, 6)
		actual, err := resource.ExportedFields(input)
		require.NoError(t, err)
		assert.Equal(t, 3, actual[0].Value.Interface().(int))
		assert.Equal(t, 4, actual[1].Value.Interface().(int))
	})
}

type TestOuterStruct struct {
	A int `export:"a"`
	B int `export:"c"`
	C int
	d int
}

type TestEmbeddingStruct struct {
	TestEmbeddedStruct
	A int `export:"a"`
}

type TestEmbeddedStruct struct {
	A int
	B int `export:"b"`
	X int `export:"x"`
}

func newOuterStruct(a, b, c, d int) *TestOuterStruct {
	return &TestOuterStruct{A: a, B: b, C: c, d: d}
}

func (t *TestOuterStruct) Deconstruct() (int, int, int, int) {
	return t.A, t.B, t.C, t.d
}

func fieldNames(in []*resource.ExportedField) (out []string) {
	for _, f := range in {
		out = append(out, f.FieldName)
	}
	return
}

func lookupNames(in []*resource.ExportedField) (out []string) {
	for _, f := range in {
		out = append(out, f.ReferenceName)
	}
	return
}
