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

package control

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/hashicorp/hcl/hcl/ast"

	"github.com/asteris-llc/converge/parse"
)

// we might want to change the keywords later, so keep them in a map, later we
// can replace map lookups with the final keyword
var keywords = map[string]string{
	"switch":  "switch",
	"case":    "case",
	"default": "default",
}

// Switch represents a switch element
type Switch struct {
	Name     string
	Branches []*Case
	Node     *parse.Node
}

// IsSwitchNode returns true if the parse node represents a switch statement
func IsSwitchNode(n *parse.Node) bool {
	if len(n.Keys) < 0 {
		return false
	}
	return n.Kind() == keywords["switch"]
}

// NewSwitch constructs a *Switch from a switch node
func NewSwitch(n *parse.Node, data []byte) (*Switch, error) {
	if n.Kind() != keywords["switch"] {
		return nil, fmt.Errorf("expected switch node but got %s", n.Kind())
	}
	s := &Switch{
		Name: n.Name(),
		Node: n,
	}
	branches, err := Cases(s, data)
	if err != nil {
		return nil, err
	}
	s.Branches = branches
	return s, nil
}

// GenerateNode generates a parse.Node for the macro-expanded placeholder from
// the switch statement
func (s *Switch) GenerateNode() (*parse.Node, error) {
	var quotedBranches []string
	for _, branch := range s.Branches {
		quotedBranches = append(quotedBranches, fmt.Sprintf("%q", branch.Name))
	}
	switchHCL := fmt.Sprintf(
		"macro.switch %q {branches = [ %s ]}",
		s.Name,
		strings.Join(quotedBranches, ","),
	)
	nodes, err := parse.Parse([]byte(switchHCL))
	if err != nil {
		return nil, err
	}
	if len(nodes) != 1 {
		return nil, errors.New("expanded macro did not parse to a single node")
	}
	return nodes[0], nil
}

// Cases returns a slice of cases
func Cases(s *Switch, data []byte) ([]*Case, error) {
	var cases []*Case
	asObjType, ok := s.Node.Val.(*ast.ObjectType)
	if !ok {
		return nil, NewTypeError("*ast.ObjectType", s.Node.Val)
	}
	for _, item := range asObjType.List.Items {
		caseNode := parse.NewNode(item)
		if itemErr := caseNode.Validate(); itemErr != nil {
			return nil, itemErr
		}
		newCase, err := ParseSwitchConditional(caseNode, data)
		if err != nil {
			return nil, err
		}
		cases = append(cases, newCase)
	}
	return cases, nil
}

// ParseSwitchConditional generates a case statement from an ast node at the
// switch statement level.  The node should be an *ast.ObjectItem whose Val is
// an *ast.ObjectType
func ParseSwitchConditional(n *parse.Node, data []byte) (*Case, error) {
	if n.Kind() == keywords["case"] {
		return ParseCase(n, data)
	}
	if n.Kind() == keywords["default"] {
		return parseDefault(n, data)
	}
	return nil, fmt.Errorf("expected `case` but got %s", n.Kind())
}

// InnerText returns the text inside of a *parse.Node whose ObjectItem has a
// value of type *ast.ObjectType.
func InnerText(n *parse.Node, data []byte) ([]byte, error) {
	asObjType, ok := n.Val.(*ast.ObjectType)
	if !ok {
		return nil, NewTypeError("*ast.ObjectType", n.Val)
	}
	start := asObjType.Lbrace.Offset + 1
	end := asObjType.Rbrace.Offset - 1
	if end > len(data) {
		return nil, errors.New("index out-of-bounds error")
	}
	return data[start:end], nil
}
