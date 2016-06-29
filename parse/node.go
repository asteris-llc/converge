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
	"fmt"

	"github.com/hashicorp/hcl/hcl/ast"
)

// Node represents a node in the parsed module
type Node struct {
	*ast.ObjectItem
}

// Validate this node
func (n *Node) Validate() error {
	switch len(n.Keys) {
	case 0:
		return fmt.Errorf("%s: no keys", n.Pos())

	case 1:
		return fmt.Errorf("%s: missing name", n.Pos())

	case 2:
		if n.IsModule() {
			return fmt.Errorf("%s: missing source or name in module call", n.Pos())
		}

	default:
		if n.IsModule() && len(n.Keys) == 3 {
			break
		}

		return fmt.Errorf("%s: too many keys", n.Pos())
	}

	return nil
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

// Source returns where a module call is to be loaded from
func (n *Node) Source() string {
	if n.IsModule() {
		return n.Keys[1].Token.Value().(string)
	}
	return ""
}

func (n *Node) String() string {
	return fmt.Sprintf(
		"%s.%s",
		n.Kind(),
		n.Name(),
	)
}
