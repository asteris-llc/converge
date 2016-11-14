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

	"github.com/asteris-llc/converge/parse"
	"github.com/pkg/errors"
)

// Case represents a case structure from a switch element.  Each case may have
// multiple nodes that will be expanded from the predicate.
type Case struct {
	Name       string
	Predicate  string
	InnerNodes []*parse.Node
}

// GenerateNode generates a parse.Node for the macro-expanded placeholder from
// the case clause
func (c *Case) GenerateNode() (*parse.Node, error) {
	switchHCL := fmt.Sprintf(
		"macro.case %q {name = %q}",
		c.Name,
		c.Name,
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

// ParseCase will parse a `case` or `default` node
func ParseCase(n *parse.Node, data []byte) (*Case, error) {
	if n.Name() == keywords["default"] {
		return nil, errors.New("case name cannot be 'default'")
	}

	innerText, err := InnerText(n, data)
	if err != nil {
		return nil, err
	}
	parsed, err := parse.Parse(innerText)
	if err != nil {
		return nil, err
	}

	return &Case{
		Name:       n.Name(),
		Predicate:  strings.TrimSpace(n.Keys[1].Token.Value().(string)),
		InnerNodes: parsed,
	}, nil
}

func parseDefault(n *parse.Node, data []byte) (*Case, error) {
	innerText, err := InnerText(n, data)
	if err != nil {
		return nil, err
	}
	parsed, err := parse.Parse(innerText)
	if err != nil {
		return nil, err
	}
	return &Case{
		Name:       "default",
		Predicate:  "true",
		InnerNodes: parsed,
	}, nil
}
