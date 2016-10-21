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

package node

// Groupable returns a group
type Groupable interface {
	Group() string
}

// Node tracks the metadata associated with a node in the graph
type Node struct {
	ID    string `json:"id"`
	Group string `json:"group"`

	value interface{}
}

// New creates a new node
func New(id string, value interface{}) *Node {
	n := &Node{
		ID:    id,
		value: value,
	}
	n.setGroup()

	return n
}

// Value gets the inner value of this node
func (n *Node) Value() interface{} {
	return n.value
}

// WithValue returns a copy of the node with the new value set
func (n *Node) WithValue(value interface{}) *Node {
	copied := new(Node)
	*copied = *n
	copied.value = value
	copied.setGroup()

	return copied
}

func (n *Node) setGroup() {
	if groupable, ok := n.value.(Groupable); ok {
		n.Group = groupable.Group()
	}
}
