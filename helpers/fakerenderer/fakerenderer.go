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

package fakerenderer

import (
	"fmt"
	"strconv"
)

// FakeRenderer is a pass-through renderer for testing resources
type FakeRenderer struct {
	DotValue     string
	ValuePresent bool
}

// Value returns a blank string
func (fr *FakeRenderer) Value() (string, bool) {
	return fr.DotValue, fr.ValuePresent
}

// Render returns whatever content is passed in
func (fr *FakeRenderer) Render(name, content string) (string, error) {
	return content, nil
}

// RequiredRender returns an error if content is an empty string
func (fr *FakeRenderer) RequiredRender(name, content string) (string, error) {
	if content == "" {
		return "", fmt.Errorf("%s is required", name)
	}
	return content, nil
}

// RenderBool renders a boolean value
func (fr *FakeRenderer) RenderBool(name, content string) (bool, error) {
	if content == "" {
		return false, nil
	}
	return strconv.ParseBool(content)
}

// RenderStringSlice renders the slice of strings passed in
func (fr *FakeRenderer) RenderStringSlice(name string, content []string) ([]string, error) {
	return content, nil
}

// RenderStringMapToStringSlice renders a map of strings to strings as a string
// slice
func (fr *FakeRenderer) RenderStringMapToStringSlice(name string, content map[string]string, toString func(string, string) string) ([]string, error) {
	return []string{}, nil
}

// New gets a default FakeRenderer
func New() *FakeRenderer {
	return new(FakeRenderer)
}

// NewWithValue gets a FakeRenderer with the appropriate value set
func NewWithValue(val string) *FakeRenderer {
	fr := New()
	fr.DotValue = val
	fr.ValuePresent = true

	return fr
}
