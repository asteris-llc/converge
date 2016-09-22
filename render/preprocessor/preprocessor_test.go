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
	"github.com/asteris-llc/converge/render/preprocessor"
	"github.com/stretchr/testify/assert"
)

func Test_Inits_WhenEmptySlice(t *testing.T) {
	assert.Nil(t, preprocessor.Inits([]string{}))
}

func Test_HasField_WhenStruct_ReturnsFieldPresentWhenPresent(t *testing.T) {
	assert.True(t, preprocessor.HasField(TestStruct{}, "FieldA"))
	assert.False(t, preprocessor.HasField(TestStruct{}, "FieldB"))

}

func Test_HasField_WhenEmbeddedStruct_ReturnsEmbeddedFieldPresent(t *testing.T) {
	type Embedded struct {
		A struct{}
	}

	type Embedding struct {
		*Embedded
		B struct{}
	}
	assert.True(t, preprocessor.HasField(Embedding{}, "B"))
	assert.True(t, preprocessor.HasField(Embedding{}, "A"))
}

func Test_VertexSplit_WhenMatchingSubstring_ReturnsPrefixAndRest(t *testing.T) {
	s := "a.b.c.d.e"
	g := graph.New()
	g.Add("a", "a")
	g.Add("a.b", "a.b.")
	g.Add("a.c.d.", "a.c.d")
	g.Add("a.b.c", "a.b.c")
	pfx, rest, found := preprocessor.VertexSplit(g, s)
	assert.Equal(t, "a.b.c", pfx)
	assert.Equal(t, "d.e", rest)
	assert.True(t, found)
}

func Test_VertexSplit_WhenExactMatch_ReturnsPrefix(t *testing.T) {
	s := "a.b.c"
	g := graph.New()
	g.Add("a.b.c", "a.b.c")
	pfx, rest, found := preprocessor.VertexSplit(g, s)
	assert.Equal(t, "a.b.c", pfx)
	assert.Equal(t, "", rest)
	assert.True(t, found)
}

func Test_VertexSplit_WhenNoMatch_ReturnsRest(t *testing.T) {
	s := "x.y.z"
	g := graph.New()
	g.Add("a.b.c", "a.b.c")
	pfx, rest, found := preprocessor.VertexSplit(g, s)
	assert.Equal(t, "", pfx)
	assert.Equal(t, "x.y.z", rest)
	assert.False(t, found)
}

func Test_HasField_WhenStructPtr_ReturnsFieldPresentWhenPresent(t *testing.T) {
	assert.True(t, preprocessor.HasField(&TestStruct{}, "FieldA"))
	assert.False(t, preprocessor.HasField(&TestStruct{}, "FieldB"))
}

func Test_HasField_WhenGivenAsLowerCaseAndIsCapital_ReturnsTrue(t *testing.T) {
	assert.True(t, preprocessor.HasField(&TestStruct{}, "fieldA"))
	assert.True(t, preprocessor.HasField(&TestStruct{}, "fielda"))
	assert.False(t, preprocessor.HasField(&TestStruct{}, "fieldB"))
}

func Test_HasField_WhenNilPtr_ReturnsTrue(t *testing.T) {
	var test *TestStruct
	assert.True(t, preprocessor.HasField(test, "FieldA"))
	assert.False(t, preprocessor.HasField(test, "FieldB"))
}

func Test_HasMethod_WhenStruct(t *testing.T) {
	assert.True(t, preprocessor.HasMethod(TestStruct{}, "FunctionOnStruct"))
	assert.False(t, preprocessor.HasMethod(TestStruct{}, "FunctionOnPointer"))
	assert.False(t, preprocessor.HasMethod(TestStruct{}, "NonExistantFunc"))
}

func Test_HasMethod_WhenStructPtr(t *testing.T) {
	assert.True(t, preprocessor.HasMethod(&TestStruct{}, "FunctionOnStruct"))
	assert.True(t, preprocessor.HasMethod(&TestStruct{}, "FunctionOnPointer"))
	assert.False(t, preprocessor.HasMethod(&TestStruct{}, "NonExistantFunc"))
}

func Test_HasMethod_WhenNilPtr_ReturnsFalse(t *testing.T) {
	var test *TestStruct
	assert.True(t, preprocessor.HasMethod(test, "FunctionOnStruct"))
	assert.True(t, preprocessor.HasMethod(test, "FunctionOnPointer"))
	assert.False(t, preprocessor.HasMethod(test, "NonExistantFunc"))
}

func Test_EvalMember_ReturnsValueWhenExists(t *testing.T) {
	expected := "foo"
	test := &TestStruct{FieldA: expected}
	actual, err := preprocessor.EvalMember("FieldA", test)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual.Interface().(string))
}

func Test_EvalMember_ReturnsValueWhenLowerCaseAndExists(t *testing.T) {
	expected := "foo"
	test := &TestStruct{FieldA: expected}
	actual, err := preprocessor.EvalMember("fieldA", test)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual.Interface().(string))
}

func Test_EvalMember_ReturnsError_WhenNotExists(t *testing.T) {
	test := &TestStruct{}
	_, err := preprocessor.EvalMember("MissingField", test)
	assert.Error(t, err)
}

func Test_LookupCanonicalFieldName_ReturnsCanonicalFieldName_WhenStructOkay(t *testing.T) {
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

}

func Test_LookupCanonicalFieldName_ReturnsNoError_WhenOverlappingFieldNames(t *testing.T) {
	type TestStruct struct {
		Xyz struct{} // collision initial upper
		XYz struct{} // collision first two upper
	}

	testType := reflect.TypeOf(TestStruct{})
	_, err := preprocessor.LookupCanonicalFieldName(testType, "xyz")
	assert.NoError(t, err)
}

func Test_LookupCanonicalFieldName_ReturnsCorrectNameWhenAnonymousField(t *testing.T) {
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
}

func Test_EvalTerms(t *testing.T) {
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
}

func Test_EvalTerms_WhenAnonymousEmbeddedStructs(t *testing.T) {
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

	result, err := preprocessor.EvalTerms(val, "conly")
	assert.NoError(t, err)
	assert.Equal(t, result, "c.conly")

	result, err = preprocessor.EvalTerms(val, "bconly")
	assert.NoError(t, err)
	assert.Equal(t, result, "c.bconly")

	result, err = preprocessor.EvalTerms(val, "aconly")
	assert.NoError(t, err)
	assert.Equal(t, result, "c.aconly")

	_, err = preprocessor.EvalTerms(val, "abonly")
	assert.Error(t, err)

	result, err = preprocessor.EvalTerms(val, "b", "abonly")
	assert.NoError(t, err)
	assert.Equal(t, result, "c.b.abonly")
}

func Test_EvalTerms_WithAnonymousFields(t *testing.T) {
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

}

func Test_EvalTerms_HandlesOverlappingFieldsNames(t *testing.T) {
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
		A
		B
		FooC    string
		Overlap string
	}
	val := C{
		A:       A{Foo: "a.foo", FooA: "a.fooa", Overlap: "a"},
		B:       B{Foo: "b.foo", FooB: "b.foob", Overlap: "b"},
		FooC:    "c.fooc",
		Overlap: "c",
	}

	_, err := preprocessor.EvalTerms(val, "foo")
	assert.Error(t, err)

	result, err := preprocessor.EvalTerms(val, "fooc")
	assert.NoError(t, err)
	assert.Equal(t, result, "c.fooc")

	result, err = preprocessor.EvalTerms(val, "fooa")
	assert.NoError(t, err)
	assert.Equal(t, result, "a.fooa")

	result, err = preprocessor.EvalTerms(val, "a", "fooa")
	assert.NoError(t, err)
	assert.Equal(t, result, "a.fooa")

	result, err = preprocessor.EvalTerms(val, "foob")
	assert.NoError(t, err)
	assert.Equal(t, result, "b.foob")

	result, err = preprocessor.EvalTerms(val, "b", "foob")
	assert.NoError(t, err)
	assert.Equal(t, result, "b.foob")

	result, err = preprocessor.EvalTerms(val, "a", "foo")
	assert.NoError(t, err)
	assert.Equal(t, result, "a.foo")

	result, err = preprocessor.EvalTerms(val, "b", "foo")
	assert.NoError(t, err)
	assert.Equal(t, result, "b.foo")

	result, err = preprocessor.EvalTerms(val, "overlap")
	assert.NoError(t, err)
	assert.Equal(t, result, "c")

	result, err = preprocessor.EvalTerms(val, "a", "overlap")
	assert.NoError(t, err)
	assert.Equal(t, result, "a")

	result, err = preprocessor.EvalTerms(val, "b", "overlap")
	assert.NoError(t, err)
	assert.Equal(t, result, "b")
}

type TestStruct struct {
	FieldA string
}

func (t TestStruct) FunctionOnStruct()   {}
func (t *TestStruct) FunctionOnPointer() {}
