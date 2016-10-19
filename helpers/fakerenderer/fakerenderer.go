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

import "github.com/asteris-llc/converge/resource"

// FakeRenderer is a pass-through renderer for testing resources
type FakeRenderer struct {
	ID           string
	DotValue     resource.Value
	ValuePresent bool
}

// GetID returns the ID of this renderer
func (fr *FakeRenderer) GetID() string {
	return fr.ID
}

// Value returns the dotvalue
func (fr *FakeRenderer) Value() (resource.Value, bool) {
	return fr.DotValue, fr.ValuePresent
}

// Render returns whatever content is passed in
func (fr *FakeRenderer) Render(name, content string) (string, error) {
	return content, nil
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

// NewWithID gets a FakeRenderer with the specified ID
func NewWithID(id string) *FakeRenderer {
	fr := New()
	fr.ID = id

	return fr
}
