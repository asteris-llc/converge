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

package preprocessor_test

import (
	"reflect"
	"testing"

	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/graph/node"
	"github.com/asteris-llc/converge/render/preprocessor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInitsWhenEmptySlice ensures that an emptly slice returns an empty slice
func TestInitsWhenEmptySlice(t *testing.T) {
	t.Parallel()
	assert.Nil(t, preprocessor.Inits([]string{}))
}

// TestHasField ensures the correct behavior when looking for a field on a
// struct
func TestHasField(t *testing.T) {
	t.Parallel()
	t.Run("TestHasFieldWhenStructReturnsFieldPresentWhenPresent", func(t *testing.T) {
		assert.True(t, preprocessor.HasField(TestStruct{}, "FieldA"))
		assert.False(t, preprocessor.HasField(TestStruct{}, "FieldB"))

	})

	t.Run("TestHasFieldWhenEmbeddedStructReturnsEmbeddedFieldPresent", func(t *testing.T) {
		type Embedded struct {
			A struct{}
		}

		type Embedding struct {
			*Embedded
			B struct{}
		}
		assert.True(t, preprocessor.HasField(Embedding{}, "B"))
		assert.True(t, preprocessor.HasField(Embedding{}, "A"))
	})

	t.Run("TestHasFieldWhenStructPtrReturnsFieldPresentWhenPresent", func(t *testing.T) {
		assert.True(t, preprocessor.HasField(&TestStruct{}, "FieldA"))
		assert.False(t, preprocessor.HasField(&TestStruct{}, "FieldB"))
	})

	t.Run("TestHasFieldWhenGivenAsLowerCaseAndIsCapitalReturnsTrue", func(t *testing.T) {
		assert.True(t, preprocessor.HasField(&TestStruct{}, "fieldA"))
		assert.True(t, preprocessor.HasField(&TestStruct{}, "fielda"))
		assert.False(t, preprocessor.HasField(&TestStruct{}, "fieldB"))
	})

	t.Run("TestHasFieldWhenNilPtrReturnsTrue", func(t *testing.T) {
		var test *TestStruct
		assert.True(t, preprocessor.HasField(test, "FieldA"))
		assert.False(t, preprocessor.HasField(test, "FieldB"))
	})

}

// TestVertexSplit ensures the correct behavior when stripping off trailing
// fields from a qualified vertex name
func TestVertexSplit(t *testing.T) {
	t.Parallel()
	t.Run("TestVertexSplitWhenMatchingSubstringReturnsPrefixAndRest", func(t *testing.T) {
		s := "a.b.c.d.e"
		g := graph.New()
		g.Add(node.New("a", "a"))
		g.Add(node.New("a.b", "a.b."))
		g.Add(node.New("a.c.d.", "a.c.d"))
		g.Add(node.New("a.b.c", "a.b.c"))
		pfx, rest, found := preprocessor.VertexSplit(g, s)
		assert.Equal(t, "a.b.c", pfx)
		assert.Equal(t, "d.e", rest)
		assert.True(t, found)
	})

	t.Run("TestVertexSplitWhenExactMatchReturnsPrefix", func(t *testing.T) {
		s := "a.b.c"
		g := graph.New()
		g.Add(node.New("a.b.c", "a.b.c"))
		pfx, rest, found := preprocessor.VertexSplit(g, s)
		assert.Equal(t, "a.b.c", pfx)
		assert.Equal(t, "", rest)
		assert.True(t, found)
	})

	t.Run("TestVertexSplitWhenNoMatchReturnsRest", func(t *testing.T) {
		s := "x.y.z"
		g := graph.New()
		g.Add(node.New("a.b.c", "a.b.c"))
		pfx, rest, found := preprocessor.VertexSplit(g, s)
		assert.Equal(t, "", pfx)
		assert.Equal(t, "x.y.z", rest)
		assert.False(t, found)
	})

}

// TestHasMethod ensures correct behavior in identifying methods on structs
func TestHasMethod(t *testing.T) {
	t.Parallel()
	t.Run("TestHasMethodWhenStruct", func(t *testing.T) {
		assert.True(t, preprocessor.HasMethod(TestStruct{}, "FunctionOnStruct"))
		assert.False(t, preprocessor.HasMethod(TestStruct{}, "FunctionOnPointer"))
		assert.False(t, preprocessor.HasMethod(TestStruct{}, "NonExistantFunc"))
	})

	t.Run("TestHasMethodWhenStructPtr", func(t *testing.T) {
		assert.True(t, preprocessor.HasMethod(&TestStruct{}, "FunctionOnStruct"))
		assert.True(t, preprocessor.HasMethod(&TestStruct{}, "FunctionOnPointer"))
		assert.False(t, preprocessor.HasMethod(&TestStruct{}, "NonExistantFunc"))
	})

	t.Run("TestHasMethodWhenNilPtrReturnsFalse", func(t *testing.T) {
		var test *TestStruct
		assert.True(t, preprocessor.HasMethod(test, "FunctionOnStruct"))
		assert.True(t, preprocessor.HasMethod(test, "FunctionOnPointer"))
		assert.False(t, preprocessor.HasMethod(test, "NonExistantFunc"))
	})

}

// TestEvalMember ensures correct behavior when extracting a field from a struct
func TestEvalMember(t *testing.T) {
	t.Parallel()
	t.Run("TestEvalMemberReturnsValueWhenExists", func(t *testing.T) {
		expected := "foo"
		test := &TestStruct{FieldA: expected}
		actual, err := preprocessor.EvalMember("FieldA", test)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual.Interface().(string))
	})

	t.Run("TestEvalMemberReturnsValueWhenLowerCaseAndExists", func(t *testing.T) {
		expected := "foo"
		test := &TestStruct{FieldA: expected}
		actual, err := preprocessor.EvalMember("fieldA", test)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual.Interface().(string))
	})

	t.Run("TestEvalMemberReturnsErrorWhenNotExists", func(t *testing.T) {
		test := &TestStruct{}
		_, err := preprocessor.EvalMember("MissingField", test)
		assert.Error(t, err)
	})
}

// TestLookupCanonicalFieldName ensures the correct behavior when normalizing a
// field name and accessing it through struct named fields, and named anonymous
// embedded structs.
func TestLookupCanonicalFieldName(t *testing.T) {
	t.Parallel()
	t.Run("TestLookupCanonicalFieldNameReturnsCanonicalFieldNameWhenStructOkay", func(t *testing.T) {
		type TestStruct struct {
			ABC struct{} // All upper
			aaa struct{} // all lower

			AcB struct{} // mixed case, initial capital
			bAc struct{} // mixed case, initial lower

			A struct{} // single letter  upper
			b struct{} // single letter lower
		}
		testType := reflect.TypeOf(TestStruct{})

		actual, err := preprocessor.LookupCanonicalFieldName(testType, "ABC")
		assert.NoError(t, err)
		assert.Equal(t, "ABC", actual)
		actual, err = preprocessor.LookupCanonicalFieldName(testType, "AbC")
		assert.NoError(t, err)
		assert.Equal(t, "ABC", actual)
		actual, err = preprocessor.LookupCanonicalFieldName(testType, "abc")
		assert.NoError(t, err)
		assert.Equal(t, "ABC", actual)

		actual, err = preprocessor.LookupCanonicalFieldName(testType, "aaa")
		assert.NoError(t, err)
		assert.Equal(t, "aaa", actual)
		actual, err = preprocessor.LookupCanonicalFieldName(testType, "aAa")
		assert.NoError(t, err)
		assert.Equal(t, "aaa", actual)
		actual, err = preprocessor.LookupCanonicalFieldName(testType, "AAA")
		assert.NoError(t, err)
		assert.Equal(t, "aaa", actual)

		actual, err = preprocessor.LookupCanonicalFieldName(testType, "acb")
		assert.NoError(t, err)
		assert.Equal(t, "AcB", actual)
		actual, err = preprocessor.LookupCanonicalFieldName(testType, "AcB")
		assert.NoError(t, err)
		assert.Equal(t, "AcB", actual)
		actual, err = preprocessor.LookupCanonicalFieldName(testType, "aCb")
		assert.NoError(t, err)
		assert.Equal(t, "AcB", actual)

		actual, err = preprocessor.LookupCanonicalFieldName(testType, "bac")
		assert.NoError(t, err)
		assert.Equal(t, "bAc", actual)
		actual, err = preprocessor.LookupCanonicalFieldName(testType, "BaC")
		assert.NoError(t, err)
		assert.Equal(t, "bAc", actual)
		actual, err = preprocessor.LookupCanonicalFieldName(testType, "BAC")
		assert.NoError(t, err)
		assert.Equal(t, "bAc", actual)

		actual, err = preprocessor.LookupCanonicalFieldName(testType, "A")
		assert.NoError(t, err)
		assert.Equal(t, "A", actual)
		actual, err = preprocessor.LookupCanonicalFieldName(testType, "a")
		assert.NoError(t, err)
		assert.Equal(t, "A", actual)

		actual, err = preprocessor.LookupCanonicalFieldName(testType, "b")
		assert.NoError(t, err)
		assert.Equal(t, "b", actual)
		actual, err = preprocessor.LookupCanonicalFieldName(testType, "B")
		assert.NoError(t, err)
		assert.Equal(t, "b", actual)

	})

	t.Run("TestLookupCanonicalFieldNameReturnsNoErrorWhenOverlappingFieldNames", func(t *testing.T) {
		type TestStruct struct {
			Xyz struct{} // collision initial upper
			XYz struct{} // collision first two upper
		}

		testType := reflect.TypeOf(TestStruct{})
		_, err := preprocessor.LookupCanonicalFieldName(testType, "xyz")
		assert.NoError(t, err)
	})

	t.Run("TestLookupCanonicalFieldNameReturnsCorrectNameWhenAnonymousField", func(t *testing.T) {
		type A struct {
			Foo string
			Bar string
		}
		type B struct {
			A
			Foo string
			Baz string
		}
		testType := reflect.TypeOf(B{})
		_, err := preprocessor.LookupCanonicalFieldName(testType, "bar")
		assert.NoError(t, err)
		_, err = preprocessor.LookupCanonicalFieldName(testType, "baz")
		assert.NoError(t, err)
		_, err = preprocessor.LookupCanonicalFieldName(testType, "foo")
		assert.NoError(t, err)
	})
}

// TestEvalTerms tests pulling field values from a struct in different scenarios
func TestEvalTerms(t *testing.T) {
	t.Parallel()
	t.Run("well-formed", func(t *testing.T) {
		type C struct {
			CVal string
		}
		type B struct {
			BVal string
			BC   *C
		}

		type A struct {
			AVal string
			AB   *B
		}
		a := &A{AVal: "a", AB: &B{BVal: "b", BC: &C{CVal: "c"}}}
		val, err := preprocessor.EvalTerms(a, "AB", "BVal")
		assert.NoError(t, err)
		assert.Equal(t, val, "b")

		val, err = preprocessor.EvalTerms(a, "AB", "BC", "CVal")
		assert.NoError(t, err)
		assert.Equal(t, val, "c")

		val, err = preprocessor.EvalTerms(a, "ab", "bc", "cval")
		assert.NoError(t, err)
		assert.Equal(t, val, "c")

		val, err = preprocessor.EvalTerms(a, "AVal")
		assert.NoError(t, err)
		assert.Equal(t, val, "a")
	})

	t.Run("non-overlapping-anonymous", func(t *testing.T) {
		type A struct {
			AOnly  string
			ACOnly string
			ABOnly string
		}
		type B struct {
			A
			ABOnly string
			BOnly  string
			BCOnly string
		}
		type C struct {
			A
			B
			COnly  string
			BCOnly string
			ACOnly string
		}

		val := C{
			B: B{
				A: A{
					AOnly:  "c.b.a.aonly",
					ABOnly: "c.b.a.abonly",
					ACOnly: "c.b.a.aconly",
				},
				ABOnly: "c.b.abonly",
				BOnly:  "c.b.bonly",
				BCOnly: "c.b.bconly",
			},
			A: A{
				AOnly:  "c.a.aonly",
				ABOnly: "c.a.abonly",
				ACOnly: "c.a.aconly",
			},
			COnly:  "c.conly",
			BCOnly: "c.bconly",
			ACOnly: "c.aconly",
		}

		t.Run("conly", func(t *testing.T) {
			result, err := preprocessor.EvalTerms(val, "conly")
			assert.NoError(t, err)
			assert.Equal(t, result, "c.conly")
		})

		t.Run("bconly", func(t *testing.T) {
			result, err := preprocessor.EvalTerms(val, "bconly")
			assert.NoError(t, err)
			assert.Equal(t, result, "c.bconly")
		})

		t.Run("aconly", func(t *testing.T) {
			result, err := preprocessor.EvalTerms(val, "aconly")
			assert.NoError(t, err)
			assert.Equal(t, result, "c.aconly")
		})

		t.Run("abonly", func(t *testing.T) {
			_, err := preprocessor.EvalTerms(val, "abonly")
			assert.Error(t, err)
		})

		t.Run("b-abonly", func(t *testing.T) {
			result, err := preprocessor.EvalTerms(val, "b", "abonly")
			assert.NoError(t, err)
			assert.Equal(t, result, "c.b.abonly")
		})
	})

	t.Run("nested-anonymous-no-overlap", func(t *testing.T) {
		type A struct {
			AField string
		}
		type B struct {
			BField string
		}
		type C struct {
			B
			CField string
		}
		type D struct {
			A
			C
			DField string
		}
		val := D{
			A: A{
				AField: "afield",
			},
			C: C{
				B: B{
					BField: "bfield",
				},
				CField: "cfield",
			},
			DField: "dfield",
		}

		result, err := preprocessor.EvalTerms(val, "dfield")
		assert.NoError(t, err)
		assert.Equal(t, result, "dfield")

		result, err = preprocessor.EvalTerms(val, "cfield")
		assert.NoError(t, err)
		assert.Equal(t, result, "cfield")

		result, err = preprocessor.EvalTerms(val, "bfield")
		assert.NoError(t, err)
		assert.Equal(t, result, "bfield")

		result, err = preprocessor.EvalTerms(val, "afield")
		assert.NoError(t, err)
		assert.Equal(t, result, "afield")

	})

	t.Run("anonymous-overlapping", func(t *testing.T) {
		type A struct {
			Foo     string
			FooA    string
			Overlap string
		}
		type B struct {
			Foo     string
			FooB    string
			Overlap string
		}
		type C struct {
			B
			FooC    string
			Overlap string
			A
		}
		val := C{
			A:       A{Foo: "a.foo", FooA: "a.fooa", Overlap: "a"},
			B:       B{Foo: "b.foo", FooB: "b.foob", Overlap: "b"},
			FooC:    "c.fooc",
			Overlap: "c",
		}

		t.Run("errors-on-foo", func(t *testing.T) {
			_, err := preprocessor.EvalTerms(val, "foo")
			assert.Error(t, err)
		})

		t.Run("when-fooc", func(t *testing.T) {
			result, err := preprocessor.EvalTerms(val, "fooc")
			assert.NoError(t, err)
			assert.Equal(t, result, "c.fooc")
		})

		t.Run("when-fooa", func(t *testing.T) {
			result, err := preprocessor.EvalTerms(val, "fooa")
			assert.NoError(t, err)
			assert.Equal(t, result, "a.fooa")
		})

		t.Run("when-a-fooa", func(t *testing.T) {
			result, err := preprocessor.EvalTerms(val, "a", "fooa")
			assert.NoError(t, err)
			assert.Equal(t, result, "a.fooa")
		})

		t.Run("when-foob", func(t *testing.T) {
			result, err := preprocessor.EvalTerms(val, "foob")
			assert.NoError(t, err)
			assert.Equal(t, result, "b.foob")
		})

		t.Run("when-b-foob", func(t *testing.T) {
			result, err := preprocessor.EvalTerms(val, "b", "foob")
			assert.NoError(t, err)
			assert.Equal(t, result, "b.foob")
		})

		t.Run("when-a-foo", func(t *testing.T) {
			result, err := preprocessor.EvalTerms(val, "a", "foo")
			assert.NoError(t, err)
			assert.Equal(t, result, "a.foo")
		})

		t.Run("when-b-foo", func(t *testing.T) {
			result, err := preprocessor.EvalTerms(val, "b", "foo")
			assert.NoError(t, err)
			assert.Equal(t, result, "b.foo")
		})

		t.Run("when-overlap", func(t *testing.T) {
			result, err := preprocessor.EvalTerms(val, "overlap")
			assert.NoError(t, err)
			assert.Equal(t, result, "c")
		})

		t.Run("when-a-overlap", func(t *testing.T) {
			result, err := preprocessor.EvalTerms(val, "a", "overlap")
			assert.NoError(t, err)
			assert.Equal(t, result, "a")
		})

		t.Run("when-b-overlap", func(t *testing.T) {
			result, err := preprocessor.EvalTerms(val, "b", "overlap")
			assert.NoError(t, err)
			assert.Equal(t, result, "b")
		})
	})
}

// Test cases where we have an embedded interface (like resource.Task)
func TestAnonymousEmbddedInterfaces(t *testing.T) {
	t.Run("has path finds non-embedded fields", func(t *testing.T) {
		s := &TestAnonEmbedded{TestInterface: &TestImplA{}, FieldD: "d"}
		assert.NoError(t, preprocessor.HasPath(s, "FieldD"))
	})
	t.Run("has path finds embedded interface", func(t *testing.T) {
		s := &TestAnonEmbedded{TestInterface: &TestImplA{}, FieldD: "d"}
		assert.NoError(t, preprocessor.HasPath(s, "TestInterface"))
	})
	t.Run("FieldMap returns fields for correct embedded struct", func(t *testing.T) {
		t.Run("when EmbeddableStruct", func(t *testing.T) {
			s := &TestAnonEmbedded{
				EmbeddableStruct: &EmbeddableStruct{},
			}
			fieldMap := preprocessor.FieldMap(s)
			val, ok := fieldMap["Embedded"]
			require.True(t, ok)
			assert.Equal(t, s.Embedded, val.Interface().(string))
			val, ok = fieldMap["EmbeddableStruct"]
			require.True(t, ok)
			assert.Equal(t, s.EmbeddableStruct, val.Interface().(*EmbeddableStruct))
		})
		t.Run("when TestImplA", func(t *testing.T) {
			s := &TestAnonEmbedded{
				TestInterface: &TestImplA{FieldA: "a"},
			}
			fieldMap := preprocessor.FieldMap(s)
			val, ok := fieldMap["TestInterface"]
			require.True(t, ok)
			assert.Equal(t, s.TestInterface, val.Interface().(TestInterface))
			val, ok = fieldMap["FieldA"]
			require.True(t, ok)
			assert.Equal(t, s.TestInterface.(*TestImplA).FieldA, val.Interface().(string))
			_, ok = fieldMap["FieldB"]
			assert.False(t, ok)
		})
		t.Run("when TestImplB", func(t *testing.T) {
			s := &TestAnonEmbedded{
				TestInterface: &TestImplB{FieldB: "b"},
			}
			fieldMap := preprocessor.FieldMap(s)
			val, ok := fieldMap["TestInterface"]
			require.True(t, ok)
			assert.Equal(t, s.TestInterface, val.Interface().(TestInterface))
			val, ok = fieldMap["FieldB"]
			require.True(t, ok)
			assert.Equal(t, s.TestInterface.(*TestImplB).FieldB, val.Interface().(string))
			_, ok = fieldMap["FieldA"]
			assert.False(t, ok)
		})
		t.Run("when nil", func(t *testing.T) {
			var s *TestAnonEmbedded
			fieldMap := preprocessor.FieldMap(s)
			_, ok := fieldMap["FieldD"]
			assert.False(t, ok)
			_, ok = fieldMap["TestInterface"]
			assert.False(t, ok)
			_, ok = fieldMap["EmbeddableStruct"]
			assert.False(t, ok)
		})
	})
	t.Run("has path finds embedded interface field", func(t *testing.T) {
		s := &TestAnonEmbedded{TestInterface: &TestImplA{}, FieldD: "d"}
		assert.NoError(t, preprocessor.HasPath(s, "fielda"))
	})
	t.Run("EvalMember gets value from an embedded interface", func(t *testing.T) {
		expected := "foo"
		test := &TestAnonEmbedded{
			TestInterface: &TestImplA{FieldA: expected},
		}
		actual, err := preprocessor.EvalMember("FieldA", test)
		require.NoError(t, err)
		assert.Equal(t, expected, actual.Interface().(string))
	})

	t.Run("EvalTerms", func(t *testing.T) {
		t.Run("Returns field inside of an anonymous struct", func(t *testing.T) {
			s := &TestAnonEmbedded{
				TestInterface: &TestImplA{FieldA: "a"},
			}
			val, err := preprocessor.EvalTerms(s, "FieldA")
			assert.NoError(t, err)
			assert.Equal(t, val, "a")
		})
	})
}

type TestInterface interface {
	InterfaceFunc() string
}

type TestImplA struct {
	FieldA      string
	FieldShared string
}

func (a *TestImplA) InterfaceFunc() string { return "TestImplA" }

type TestImplB struct {
	FieldB      string
	FieldShared string
}

func (a *TestImplB) InterfaceFunc() string { return "TestImplB" }

type EmbeddableStruct struct {
	Embedded string
}

type TestAnonEmbedded struct {
	TestInterface
	*EmbeddableStruct
	FieldD string
}

// TestStruct is a stub object
type TestStruct struct {
	FieldA string
}

// FunctionOnStruct is a stub function
func (t TestStruct) FunctionOnStruct() {}

// FunctionOnPointer is a stub functino
func (t *TestStruct) FunctionOnPointer() {}
