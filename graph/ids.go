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

package graph

import (
	"path"
	"strings"
)

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

// AreSiblingIDs checks if two IDs are siblings
func AreSiblingIDs(a, b string) bool {
	return ParentID(a) == ParentID(b)
}

// BaseID is the end of the ID, so "just" the original part
func BaseID(id string) string {
	return path.Base(id)
}

// IsDescendentID checks if a child is the descendent of a given parent
func IsDescendentID(parent, child string) bool {
	return parent != child && strings.HasPrefix(child, parent)
}
