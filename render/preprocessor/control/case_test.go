package control_test

import (
	"testing"

	"github.com/asteris-llc/converge/parse"

	"github.com/asteris-llc/converge/render/preprocessor/control"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCaseNode tests the generation of a *parse.Node with the correct metadata
// about the case node
func TestCaseNode(t *testing.T) {
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

	nodes, err := parse.Parse([]byte(sampleStatement))
	require.NoError(t, err)
	switchObj, err := control.NewSwitch(nodes[0], []byte(sampleStatement))
	require.NoError(t, err)
	require.True(t, len(switchObj.Branches) > 0)
	caseObj := switchObj.Branches[0]
	caseNode, err := caseObj.GenerateNode()
	require.NoError(t, err)
	t.Run("sets the node key to macro.case", func(t *testing.T) {
		assert.Equal(t, "macro.case", caseNode.Kind())
	})
}

// TestCaseIsTrue ensures truthiness
func TestCaseIsTrue(t *testing.T) {
	c := &control.Case{}
	t.Run("true when predicate is 'true'", func(t *testing.T) {
		c.Predicate = "true"
		actual, err := c.IsTrue()
		assert.NoError(t, err)
		assert.True(t, actual)
	})
	t.Run("false when predicate is 'false'", func(t *testing.T) {
		c.Predicate = "false"
		actual, err := c.IsTrue()
		assert.NoError(t, err)
		assert.False(t, actual)
	})
	t.Run("true when predicate is true template expression", func(t *testing.T) {
		c.Predicate = "eq 1 1"
		actual, err := c.IsTrue()
		assert.NoError(t, err)
		assert.True(t, actual)
	})
	t.Run("false when predicate is false template expression", func(t *testing.T) {
		c.Predicate = "eq 1 0"
		actual, err := c.IsTrue()
		assert.NoError(t, err)
		assert.False(t, actual)
	})
	t.Run("error when predicate is empty", func(t *testing.T) {
		c.Predicate = ""
		_, err := c.IsTrue()
		assert.Error(t, err)
	})
	t.Run("error when predicate is non-truthy value", func(t *testing.T) {
		c.Predicate = "sven"
		_, err := c.IsTrue()
		assert.Error(t, err)
	})
	t.Run("error when template returns error", func(t *testing.T) {
		c.Predicate = "eq 1"
		_, err := c.IsTrue()
		assert.Error(t, err)
	})
}
