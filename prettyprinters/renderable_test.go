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
	"testing"

	pp "github.com/asteris-llc/converge/prettyprinters"
	"github.com/stretchr/testify/assert"
)

func Test_VisibleString_SetsString(t *testing.T) {
	expected := "Test String"
	r := pp.VisibleString(expected)
	assert.Equal(t, expected, r.String())
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

func Test_Renderable_ShowsUpAsStringWhenVisibleAndPrinted(t *testing.T) {
	expected := "Test string"
	r := pp.VisibleString(expected)
	assert.Equal(t, expected, fmt.Sprintf("%v", r))
}

func Test_Renderable_ReturnsEmptyStringWhenHidden(t *testing.T) {
	r := pp.RenderableString("anything", false)
	assert.Equal(t, "", r.String())
}
