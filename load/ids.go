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

package load

import "path"

// ID for a node in the graph
func ID(parts ...string) string {
	return path.Join(parts...)
}

// ParentID for a node in the graph
func ParentID(id string) string {
	return path.Dir(id)
}

// SiblingID for a node in the graph
func SiblingID(id, sibling string) string {
	return ID(ParentID(id), sibling)
}
