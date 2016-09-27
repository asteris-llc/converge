package control

import (
	"errors"
	"fmt"
	"strings"

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

// Preprocessor defines the general preprocessor for hcl control structures
type Preprocessor struct {
	Data []byte
}

// Case represents a case structure from a switch element
type Case struct {
	Name      string
	Predicate string
	InnerNode *parse.Node
	OuterNode *parse.Node
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
func (p *Preprocessor) NewSwitch(n *parse.Node) (*Switch, error) {
	if n.Kind() != keywords["switch"] {
		return nil, fmt.Errorf("expected switch node but got %s", n.Kind())
	}
	s := &Switch{
		Name: n.Name(),
		Node: n,
	}

	branches, err := p.Cases(s)
	if err != nil {
		return nil, err
	}
	s.Branches = branches
	return s, nil
}

// Cases returns a slice of cases
func (p *Preprocessor) Cases(s *Switch) ([]*Case, error) {
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
		newCase, err := p.ParseCase(caseNode)
		if err != nil {
			return nil, err
		}
		cases = append(cases, newCase)
	}
	return cases, nil
}

// ParseCase generates a case statement from an ast node at the switch statement
// level.  The node should be an *ast.ObjectItem whose Val is an *ast.ObjectType
func (p *Preprocessor) ParseCase(n *parse.Node) (*Case, error) {
	if n.Kind() != keywords["case"] {
		return nil, fmt.Errorf("expected `case` but got %s", n.Kind())
	}
	c := &Case{
		Name:      n.Name(),
		Predicate: strings.TrimSpace(n.Keys[1].Token.Value().(string)),
		OuterNode: n,
	}

	return c, nil
}

// InnerText returns the text inside of a *parse.Node whose ObjectItem has a
// value of type *ast.ObjectType.
func (p *Preprocessor) InnerText(n *parse.Node) ([]byte, error) {
	asObjType, ok := n.Val.(*ast.ObjectType)
	if !ok {
		return nil, NewTypeError("*ast.ObjectType", n.Val)
	}

	start := asObjType.Lbrace.Offset + 1
	end := asObjType.Rbrace.Offset - 1
	if end > len(p.Data) {
		return nil, errors.New("index out-of-bounds error")
	}
	return p.Data[start:end], nil
}
