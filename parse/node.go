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

package parse

import (
	"errors"
	"fmt"
	"reflect"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
)

// ErrNotFound is returned from Get and friends when the key does not exist
var ErrNotFound = errors.New("key does not exist")

// Node represents a node in the parsed module
type Node struct {
	*ast.ObjectItem

	values map[string]interface{}
	once   sync.Once
}

// NewNode constructs a new Node from the given ObjectItem
func NewNode(item *ast.ObjectItem) *Node {
	return &Node{ObjectItem: item}
}

// Validate this node
func (n *Node) Validate() error {
	if n == nil {
		return errors.New("node is empty, check for bad input")
	}

	switch len(n.Keys) {
	case 0:
		return fmt.Errorf("%s: no keys", n.Pos())

	case 1:
		if n.IsDefault() {
			break
		}
		return fmt.Errorf("%s: missing name", n.Pos())

	case 2:
		if n.IsModule() {
			return fmt.Errorf("%s: missing source or name in module call", n.Pos())
		}

		if n.IsDefault() {
			return fmt.Errorf("%s: too many keys", n.Pos())
		}

		if n.IsCase() {
			return fmt.Errorf("%s: missing name or predicate in case", n.Pos())
		}

	default:
		if n.IsModule() && len(n.Keys) == 3 {
			break
		}

		if n.IsCase() && len(n.Keys) == 3 {
			break
		}

		return fmt.Errorf("%s: too many keys", n.Pos())
	}

	return n.setValues()
}

// Kind returns the kind of resource this is
func (n *Node) Kind() string {
	return n.Keys[0].Token.Value().(string)
}

// Name returns the name of the resource
func (n *Node) Name() string {
	return n.Keys[len(n.Keys)-1].Token.Value().(string)
}

// IsModule tests whether this node is a module call
func (n *Node) IsModule() bool {
	return n.Kind() == "module"
}

// IsCase tests whether this node is a case statement
func (n *Node) IsCase() bool {
	return n.Kind() == "case"
}

// IsDefault tests whether this node is a default case statement
func (n *Node) IsDefault() bool {
	return n.Kind() == "default"
}

// Source returns where a module call is to be loaded from
func (n *Node) Source() string {
	if n.IsModule() {
		return n.Keys[1].Token.Value().(string)
	}
	return ""
}

func (n *Node) setValues() (err error) {
	n.once.Do(func() {
		n.values = map[string]interface{}{}

		err = hcl.DecodeObject(&n.values, n.Val)
	})

	return err
}

// Get a value from the values
func (n *Node) Get(key string) (val interface{}, err error) {
	if err := n.setValues(); err != nil {
		return nil, err
	}

	val, ok := n.values[key]
	if !ok {
		return val, ErrNotFound
	}

	return val, nil
}

// GetString retrieves string value from the values
func (n *Node) GetString(key string) (val string, err error) {
	raw, err := n.Get(key)
	if err != nil {
		return "", err
	}

	val, ok := raw.(string)
	if !ok {
		return "", n.badTypeError(key, "string", raw)
	}

	return val, nil
}

// GetStringSlice retrieves a slice of string from the values
func (n *Node) GetStringSlice(key string) (val []string, err error) {
	raw, err := n.Get(key)
	if err != nil {
		return nil, err
	}

	interfaces, ok := raw.([]interface{})
	if !ok {
		return nil, n.badTypeError(key, "slice", raw)
	}

	for i, iface := range interfaces {
		item, ok := iface.(string)
		if !ok {
			return nil, n.badTypeError(fmt.Sprintf("%s.%d", key, i), "string", iface)
		}

		val = append(val, item)
	}

	return val, nil
}

// GetStrings retrieves all the strings in the node
func (n *Node) GetStrings() (vals []string, err error) {
	if err := n.setValues(); err != nil {
		return nil, err
	}

	toConsider := []interface{}{}
	for _, val := range n.values {
		toConsider = append(toConsider, val)
	}

	for len(toConsider) > 0 {
		val := toConsider[0]
		toConsider = toConsider[1:]

		switch val.(type) {
		case string:
			vals = append(vals, val.(string))

		case []map[string]interface{}:
			for _, sub := range val.([]map[string]interface{}) {
				toConsider = append(toConsider, interface{}(sub))
			}

		case map[string]interface{}:
			for key, value := range val.(map[string]interface{}) {
				toConsider = append(toConsider, key)
				toConsider = append(toConsider, value)
			}

		case []interface{}:
			toConsider = append(toConsider, val.([]interface{})...)

		default:
			log.WithField("type", fmt.Sprintf("%T", val)).WithField("val", val).Debug("unknown value")
		}
	}

	return vals, nil
}

func (n *Node) badTypeError(key, typ string, val interface{}) error {
	article := func(x string) string {
		switch x[0] {
		case 'a', 'e', 'i', 'o', 'u':
			return "an"
		default:
			return "a"
		}
	}

	valTyp := reflect.TypeOf(val).String()

	return fmt.Errorf(
		"%q is not %s %s, it is %s %s",
		key,
		article(typ), typ,
		article(valTyp), valTyp,
	)
}

func (n *Node) String() string {
	return fmt.Sprintf(
		"%s.%s",
		n.Kind(),
		n.Name(),
	)
}
