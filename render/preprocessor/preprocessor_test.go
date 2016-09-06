package preprocessor_test

import (
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

func Test_EvalMember_ReturnsError_WhenNotExists(t *testing.T) {
	test := &TestStruct{}
	_, err := preprocessor.EvalMember("MissingField", test)
	assert.Error(t, err)
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

type TestStruct struct {
	FieldA string
}

func (t TestStruct) FunctionOnStruct()   {}
func (t *TestStruct) FunctionOnPointer() {}
