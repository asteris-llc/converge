package control_test

import (
	"testing"

	"github.com/asteris-llc/converge/parse"
	"github.com/asteris-llc/converge/render/preprocessor/control"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var sampleStatement = `
switch "named-switch" {
	case "eq 1 0" "a" {
		task.query "foo" {
			query = "echo foo"
		}
	}
	case "eq ` + "`foo` `foo`" + ` " "b" {
		task.query "bar" {
			query = "echo bar"
		}
	}
	case "eq 0 1" "c" {
		task.query "baz" {
			query = "echo baz"
		}
	}
}

task.query "query" {
	query = "echo foo"
}
`
var sampleCase = `
case "eq 1 0" "a" {
	task.query "foo" {
		query = "echo foo"
	}
}
`

// TestIsSwitch tests code that checks a parse node to see if it's as switch
func TestIsSwitch(t *testing.T) {
	nodes, err := parse.Parse([]byte(sampleStatement))
	require.NoError(t, err)
	t.Run("returns true when node is a switch", func(t *testing.T) {
		switchNode := nodes[0]
		assert.True(t, control.IsSwitchNode(switchNode))
	})
	t.Run("returns false when node is not switch", func(t *testing.T) {
		notSwitchNode := nodes[1]
		assert.False(t, control.IsSwitchNode(notSwitchNode))
	})
}

// TestLoadSwitch tests loading of a node into a switch structure
func TestLoadSwitch(t *testing.T) {
	sampleCaseSlice := []*control.Case{
		&control.Case{
			Name:      "a",
			Predicate: "eq 1 0",
		},
		&control.Case{
			Name:      "b",
			Predicate: "eq `foo` `foo`",
		},
		&control.Case{
			Name:      "c",
			Predicate: "eq 0 1",
		},
	}
	nodes, err := parse.Parse([]byte(sampleStatement))
	require.NoError(t, err)
	t.Run("returns a switch with the correct name", func(t *testing.T) {
		switchStatement, err := control.NewSwitch(nodes[0])
		assert.NoError(t, err)
		assert.Equal(t, "named-switch", switchStatement.Name)
	})
	t.Run("returns a switch with the inner node set", func(t *testing.T) {
		switchStatement, err := control.NewSwitch(nodes[0])
		assert.NoError(t, err)
		assert.Equal(t, nodes[0], switchStatement.Node)
	})
	t.Run("returns a switch with a list of Cases", func(t *testing.T) {
		switchStatement, err := control.NewSwitch(nodes[0])
		assert.NoError(t, err)
		assert.Equal(t, len(sampleCaseSlice), len(switchStatement.Branches))
		for idx, branch := range switchStatement.Branches {
			c := sampleCaseSlice[idx]
			c.OuterNode = branch.OuterNode
			c.InnerNode = branch.InnerNode
			assert.Equal(t, c, branch)
		}
	})
}

// TestParseCase contains tests for parsing a case statement from an
// *ast.ObjectItem.
func TestParseCase(t *testing.T) {
	p := &control.Preprocessor{Data: []byte(sampleCase)}
	nodes, err := parse.Parse([]byte(sampleCase))
	require.NoError(t, err)
	assert.True(t, len(nodes) > 0)
	caseNode := nodes[0]
	t.Run("sets the name", func(t *testing.T) {
		parsedCase, err := p.ParseCase(caseNode)
		assert.NoError(t, err)
		assert.Equal(t, "a", parsedCase.Name)
	})
	t.Run("sets the predicate", func(t *testing.T) {
		parsedCase, err := p.ParseCase(caseNode)
		assert.NoError(t, err)
		assert.Equal(t, "eq 1 0", parsedCase.Predicate)
	})
	t.Run("sets the outer node", func(t *testing.T) {
		parsedCase, err := p.ParseCase(caseNode)
		assert.NoError(t, err)
		assert.Equal(t, caseNode, parsedCase.OuterNode)
	})
}

// TestInnerText ensures that we correctly extract the text from the inside of a
// node.
func TestInnerText(t *testing.T) {
	p := &control.Preprocessor{Data: []byte(sampleStatement)}
	nodes, err := parse.Parse([]byte(sampleStatement))
	require.NoError(t, err)
	assert.True(t, len(nodes) > 0)
	t.Run("gets the full inner text of a switch statement", func(t *testing.T) {
		expected := `
	case "eq 1 0" "a" {
		task.query "foo" {
			query = "echo foo"
		}
	}
	case "eq ` + "`foo` `foo`" + ` " "b" {
		task.query "bar" {
			query = "echo bar"
		}
	}
	case "eq 0 1" "c" {
		task.query "baz" {
			query = "echo baz"
		}
	}`
		actual, err := p.InnerText(nodes[0])
		assert.NoError(t, err)
		assert.Equal(t, expected, string(actual))
	})
	t.Run("gets the full inner text of a case statement", func(t *testing.T) {
		expected := `
		task.query "foo" {
			query = "echo foo"
		}
`
		switchNode, err := control.NewSwitch(nodes[0])
		assert.NoError(t, err)
		caseNodes, err := switchNode.Cases()
		assert.NoError(t, err)
		actual, err := p.InnerText(caseNodes[0].OuterNode)
		assert.NoError(t, err)
		assert.Equal(t, expected, string(actual))
	})
}
