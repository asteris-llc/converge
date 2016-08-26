package preprocessor_test

import (
	"testing"

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

func Test_HasField_WhenStructPtr_ReturnsFieldPresentWhenPresent(t *testing.T) {
	assert.True(t, preprocessor.HasField(&TestStruct{}, "FieldA"))
	assert.False(t, preprocessor.HasField(&TestStruct{}, "FieldB"))
}

func Test_HasField_WhenNilPtr_ReturnsFalse(t *testing.T) {
	var test *TestStruct
	assert.False(t, preprocessor.HasField(test, "FieldA"))
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

func Test_EvalMember_ReturnsError_WhenNotExists(t *testing.T) {
	test := &TestStruct{}
	_, err := preprocessor.EvalMember("MissingField", test)
	assert.Error(t, err)
}

type TestStruct struct {
	FieldA string
}

func (t TestStruct) FunctionOnStruct()   {}
func (t *TestStruct) FunctionOnPointer() {}
