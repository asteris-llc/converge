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

package prettyprinters_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	pp "github.com/asteris-llc/converge/prettyprinters"
	"github.com/stretchr/testify/assert"
)

func Test_VisibleString_SetsHiddenToFalse(t *testing.T) {
	r := pp.VisibleString("test string")
	assert.False(t, r.Hidden)
}

func Test_VisibleString_SetsString(t *testing.T) {
	expected := "Test String"
	r := pp.VisibleString(expected)
	assert.Equal(t, expected, r.Contents)
}

func Test_HiddenString_SetsHiddenToTrue(t *testing.T) {
	r := pp.HiddenString("test string")
	assert.True(t, r.Hidden)
}

func Test_HiddenString_SetsString(t *testing.T) {
	expected := "Test string"
	r := pp.HiddenString(expected)
	assert.Equal(t, expected, r.Contents)
}

func TestStringRenderable_VisibleReturnsFalseWhenHidden(t *testing.T) {
	str := pp.HiddenString("test string")
	assert.False(t, str.Visible())
}

func TestStringRenderable_VisibleReturnsTrueWhenNotHidden(t *testing.T) {
	str := pp.VisibleString("test string")
	assert.True(t, str.Visible())
}

func TestStringRenderable_String_ReturnsStringWhenVisible(t *testing.T) {
	expected := "Test string"
	r := pp.VisibleString(expected)
	assert.Equal(t, expected, r.String())
}

func TestStringRenderable_String_ReturnsStringWhenHidden(t *testing.T) {
	expected := ""
	r := pp.HiddenString("Test string")
	assert.Equal(t, expected, r.String())
}

func Test_Renderable_ShowsUpAsStringWhenVisibleAndPrinted(t *testing.T) {
	expected := "Test string"
	r := pp.VisibleString(expected)
	assert.Equal(t, expected, fmt.Sprintf("%v", r))
}

func TestWrappedRenderable_VisibleReturnsTrueWhenBaseValueVisible(t *testing.T) {
	base := pp.VisibleString("test string")
	wrapped := pp.ApplyRenderable(base, stringIdentity)
	assert.True(t, wrapped.Visible())
}

func TestWrappedRenderable_VisibleReturnsFalseWhenBaseValueHidden(t *testing.T) {
	base := pp.HiddenString("test string")
	wrapped := pp.ApplyRenderable(base, stringIdentity)
	assert.False(t, wrapped.Visible())
}

func TestWrappedRenderable_StringCallsWrappedFunction(t *testing.T) {
	called := false
	f := func(s string) string {
		called = true
		return s
	}
	wrapped := pp.ApplyRenderable(pp.VisibleString("test string"), f)
	wrapped.String()
	assert.True(t, called)
}

func TestWrappedRenderable_WrappedFunctionLazilyEvaluated(t *testing.T) {
	called := false
	f := func(s string) string {
		called = true
		return s
	}
	wrapped := pp.ApplyRenderable(pp.VisibleString("test string"), f)
	assert.False(t, called)
	wrapped.String()
	assert.True(t, called)
}

func TestWrappedRenderable_PushesCallsOntoAStack(t *testing.T) {
	expected := []string{"f2", "f1"}
	var calls []string
	f1 := func(s string) string {
		calls = append(calls, "f2")
		return s
	}
	f2 := func(s string) string {
		calls = append(calls, "f1")
		return s
	}
	wrapped := pp.ApplyRenderable(pp.VisibleString("test string"), f1)
	wrapped = pp.ApplyRenderable(wrapped, f2)
	wrapped.String()
	assert.True(t, reflect.DeepEqual(expected, calls))
}

func TestVisibilityWrapper_ReturnsBaseVisibilityWhenNilToggle(t *testing.T) {
	assert.True(t, pp.Untoggle(pp.VisibleString("test string")).Visible())
	assert.False(t, pp.Untoggle(pp.HiddenString("test string")).Visible())
}

func TestVisibilityWrapper_ReturnsVisibleWhenVisibilityToggledOn(t *testing.T) {
	assert.True(t, pp.Unhide(pp.HiddenString("test string")).Visible())
}

func TestVisibilityWrapper_ReturnsHiddenWhenVisibilityToggledOn(t *testing.T) {
	assert.False(t, pp.Hide(pp.VisibleString("test string")).Visible())
}

func TestVisibilityWrapper_WorksWithNestedValues(t *testing.T) {
	wrapped := pp.VisibleString("test string")
	wrapper := pp.Untoggle(wrapped)
	assert.True(t, wrapper.Visible())
	wrapper = pp.Hide(wrapper)
	assert.False(t, wrapper.Visible())
	wrapper = pp.Unhide(wrapper)
	assert.True(t, wrapper.Visible())
	wrapper = pp.Hide(wrapper)
	assert.False(t, wrapper.Visible())
}

func TestVisibilityWrapper_UntoggleStripsLayersOfWrapping(t *testing.T) {
	wrapper := pp.Untoggle(pp.VisibleString("test string"))
	for i := 0; i < 100; i++ {
		wrapper = pp.Hide(wrapper)
	}
	assert.True(t, pp.Untoggle(wrapper).Visible())
}

func ExampleApplyRenderable() {
	o := pp.HiddenString(" foo ")
	t := pp.ApplyRenderable(pp.ApplyRenderable(o, strings.ToUpper), strings.TrimSpace)
	fmt.Println("Before making o visible:")
	fmt.Println(t)
	o.Hidden = false
	fmt.Println("After making o visible:")
	fmt.Println(t)

	// Output:
	// Before making o visible:

	// After making o visible:
	// FOO
}

/// Utility Functions

func stringIdentity(i string) string {
	return i
}
