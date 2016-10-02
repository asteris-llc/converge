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

// GenerateNode generates a parse.Node for the macro-expanded placeholde from
// the case clause
func (c *Case) GenerateNode() (*parse.Node, error) {
	switchHCL := fmt.Sprintf("macro.case %q {\n", c.Name)
	switchHCL = fmt.Sprintf("%s\tpredicate = %q\n", switchHCL, c.Predicate)
	switchHCL = fmt.Sprintf("%s\tname = %q\n", switchHCL, c.Name)
	switchHCL = fmt.Sprintf("%s}", switchHCL)
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
