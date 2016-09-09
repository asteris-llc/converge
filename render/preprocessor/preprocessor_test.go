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
	"errors"
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

func Test_HasField_WhenGivenAsLowerCaseAndIsCapital_ReturnsTrueI(t *testing.T) {
	assert.True(t, preprocessor.HasField(&TestStruct{}, "fieldA"))
	assert.False(t, preprocessor.HasField(&TestStruct{}, "fielda"))
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

func Test_MethodReturnType_ReturnsErrorWhenNotFuncType(t *testing.T) {
	_, err := preprocessor.MethodReturnType(reflect.TypeOf((*int)(nil)))
	assert.Error(t, err)
}

func Test_MethodReturnType_ReturnsTypeSliceForReturnArity1(t *testing.T) {
	expected := []reflect.Type{
		reflect.TypeOf((*int)(nil)).Elem(),
	}
	methodType := reflect.TypeOf((&TestStruct{}).SingleReturnFunction)
	actual, err := preprocessor.MethodReturnType(methodType)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func Test_MethodReturnType_ReturnsTypeSliceForMultiReturn(t *testing.T) {
	expected := []reflect.Type{
		reflect.TypeOf((*int)(nil)).Elem(),
		reflect.TypeOf((*error)(nil)).Elem(),
	}
	methodType := reflect.TypeOf((&TestStruct{}).MultiReturnFunction2)
	actual, err := preprocessor.MethodReturnType(methodType)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
	expected = append([]reflect.Type{reflect.TypeOf((*int)(nil)).Elem()}, expected...)
	methodType = reflect.TypeOf((&TestStruct{}).MutliReturnFunction3)
	actual, err = preprocessor.MethodReturnType(methodType)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
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

func Test_EvalMethod_ReturnsErrorWhenNoMethod(t *testing.T) {
	obj := &TestStruct{}
	_, err := preprocessor.EvalMethod("DoesNotExist", obj)
	assert.Error(t, err)
}

func Test_EvalMethod_ReturnsValueNoErrorWhenSingleReturn(t *testing.T) {
	obj := &TestStruct{}
	expectedValue := 1
	result, err := preprocessor.EvalMethod("SingleReturnFunction", obj)
	assert.NoError(t, err)
	assert.Equal(t, expectedValue, result.Interface().(int))
}

func Test_EvalMethod_ReturnsError_WhenSingleErrorReturn(t *testing.T) {
	obj := &TestStruct{}
	_, err := preprocessor.EvalMethod("SingleReturnError", obj)
	assert.Error(t, err)
	assert.Equal(t, errTestReturn, err)
}

func Test_EvalMethod_ReturnsValError_WhenMultiReturn2(t *testing.T) {
	obj := &TestStruct{}
	expectedVal := 1
	val, err := preprocessor.EvalMethod("MultiReturnFunction2Err", obj)
	assert.Error(t, err)
	assert.Equal(t, errTestReturn, err)
	assert.Equal(t, expectedVal, val.Interface().(int))
}

func Test_EvalMethod_ReturnsValueSlice_WhenMultiReturn3(t *testing.T) {
	obj := &TestStruct{}
	expected := []interface{}{1, 2}
	val, err := preprocessor.EvalMethod("MutliReturnFunction3", obj)
	assert.NoError(t, err)
	assert.Equal(t, expected, val.Interface().([]interface{}))
}

func Test_EvalMethod_ReturnsValueSlice_WhenMultiReturnErr3(t *testing.T) {
	obj := &TestStruct{}
	expected := []interface{}{1, 2}
	val, err := preprocessor.EvalMethod("MutliReturnFunction3Err", obj)
	assert.Error(t, err)
	assert.Equal(t, errTestReturn, err)
	assert.Equal(t, expected, val.Interface().([]interface{}))
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
	val, err = preprocessor.EvalTerms(a, "AVal")
	assert.NoError(t, err)
	assert.Equal(t, val, "a")
}

var errTestReturn = errors.New("returned error")

type TestStruct struct {
	FieldA string
}

func (t TestStruct) FunctionOnStruct() {
}

func (t *TestStruct) FunctionOnPointer() {
}

func (t *TestStruct) SingleReturnFunction() int {
	return 1
}
func (t *TestStruct) MultiReturnFunction2() (int, error) {
	return 1, nil
}
func (t *TestStruct) MultiReturnFunction2Err() (int, error) {
	return 1, errTestReturn
}

func (t *TestStruct) MutliReturnFunction3() (int, int, error) {
	return 1, 2, nil
}

func (t *TestStruct) MutliReturnFunction3Err() (int, int, error) {
	return 1, 2, errTestReturn
}

func (t *TestStruct) SingleReturnError() error {
	return errTestReturn
}
