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

package prettyprinters

import "fmt"

// Renderable provides an interface for printable objects
type Renderable interface {
	// The Renderable interface should provide an instance of String() to render
	// the object.
	fmt.Stringer
}

// VisibleRenderable allows checking if a string is visible
type VisibleRenderable interface {
	Renderable

	// Visible returns true if the object should be rendered, and false
	// otherwise.  If a consumer chooses to ignore this value, the instance should
	// still provide a valid string value.
	Visible() bool
}

// StringRenderable provides a Renderable wrapper around strings.
type stringRenderable struct {
	Hidden   bool
	Contents string
}

// Visible returns the embedded Hidden field
func (r *stringRenderable) Visible() bool {
	return !r.Hidden
}

// String returns the contents of the string
func (r *stringRenderable) String() string {
	if r.Visible() {
		return r.Contents
	}

	return ""
}

// VisibleString creates a new Renderable that is visible
func VisibleString(s string) VisibleRenderable {
	return RenderableString(s, true)
}

// SprintfRenderable creates a new visible Renderable from a Sprintf call
func SprintfRenderable(visible bool, fmtStr string, args ...interface{}) VisibleRenderable {
	if !visible {
		return HiddenString()
	}

	return VisibleString(fmt.Sprintf(fmtStr, args...))
}

// HiddenString creates a non-renderable string
func HiddenString() VisibleRenderable {
	return RenderableString("", false)
}

// RenderableString creates a RenderableString with visibility on or off
// depending on the value of a boolean parameter.
func RenderableString(s string, visible bool) VisibleRenderable {
	return &stringRenderable{
		Hidden:   !visible,
		Contents: s,
	}
}
