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

import (
	"bytes"
	"errors"
	"fmt"
)

// Groupable returns a group
type Groupable interface {
	Group() string
}

// ErrMetadataNotUnique indicates that the user attempted to overwrite a node
// metadata field.
var ErrMetadataNotUnique = errors.New("metadata field is non-unique")

// Node tracks the metadata associated with a node in the graph
type Node struct {
	ID    string `json:"id"`
	Group string `json:"group"`

	metadata map[string]interface{}
	value    interface{}
}

// New creates a new node
func New(id string, value interface{}) *Node {
	n := &Node{
		ID:       id,
		value:    value,
		metadata: make(map[string]interface{}),
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

// AddMetadata will allow you to add metadata to the node.  If the key already
// exists it will return ErrMetadataNotUnique to ensure immutability
func (n *Node) AddMetadata(key string, value interface{}) error {
	if n.metadata == nil {
		n.metadata = make(map[string]interface{})
	}
	if found, ok := n.metadata[key]; ok && found != value {
		return ErrMetadataNotUnique
	}
	n.metadata[key] = value
	return nil
}

// LookupMetadata extracts a metdatadata field from the node. If the value is
// found, it returns (value, true), and (nil, false) otherwise.
func (n *Node) LookupMetadata(key string) (interface{}, bool) {
	if n.metadata == nil {
		n.metadata = make(map[string]interface{})
	}
	result, ok := n.metadata[key]
	return result, ok
}

// ShowMetadata will print out the existing metadata.  Used for debugging
func (n *Node) ShowMetadata() string {
	if n == nil {
		return "<nil>"
	}
	var buffer bytes.Buffer
	buffer.Write([]byte(fmt.Sprintf("ID:\t%s\nGroup:\t%s\nValue Type:\t%T\nMetadata:\n", n.ID, n.Group, n.value)))
	for k, v := range n.metadata {
		buffer.Write([]byte(fmt.Sprintf("\t%s => %v\n", k, v)))
	}
	return buffer.String()
}
