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

// FakeRenderer is a pass-through renderer for testing resources
type FakeRenderer struct {
}

// Value returns a blank string
func (fr *FakeRenderer) Value() (string, bool) {
	return "", true
}

// Render returns whatever content is passed in
func (fr *FakeRenderer) Render(name, content string) (string, error) {
	return content, nil
}
