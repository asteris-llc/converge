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

package control_test

import (
	"reflect"
	"testing"

	"github.com/asteris-llc/converge/parse"
	"github.com/asteris-llc/converge/parse/preprocessor/switch"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var sampleInnerHCL = `
		task.query "foo" {
			query = "echo foo"
		}
`

// TestIsSwitch tests code that checks a parse node to see if it's as switch
func TestIsSwitch(t *testing.T) {
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
		switchStatement, err := control.NewSwitch(nodes[0], []byte(sampleStatement))
		assert.NoError(t, err)
		assert.Equal(t, "named-switch", switchStatement.Name)
	})
	t.Run("returns a switch with the inner node set", func(t *testing.T) {
		switchStatement, err := control.NewSwitch(nodes[0], []byte(sampleStatement))
		assert.NoError(t, err)
		assert.Equal(t, nodes[0], switchStatement.Node)
	})
	t.Run("returns a switch with a list of Cases", func(t *testing.T) {
		switchStatement, err := control.NewSwitch(nodes[0], []byte(sampleStatement))
		assert.NoError(t, err)
		assert.Equal(t, len(sampleCaseSlice), len(switchStatement.Branches))
		for idx, branch := range switchStatement.Branches {
			c := sampleCaseSlice[idx]
			c.InnerNodes = branch.InnerNodes
			assert.Equal(t, c, branch)
		}
	})
}

// TestSwitchNode tests the generation of a *parse.Node with the correct
// metadata about the switch node
func TestSwitchNode(t *testing.T) {
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
	switchNode, err := switchObj.GenerateNode()
	require.NoError(t, err)
	t.Run("sets the node key to macro.switch", func(t *testing.T) {
		assert.Equal(t, "macro.switch", switchNode.Kind())
	})
}

// TestParseSwitchConditionalWithCase contains tests for parsing a case
// statement
func TestParseSwitchConditionalWithCase(t *testing.T) {
	var sampleCase = `
switch "named-switch" {
	case "eq 1 0" "a" {
		task.query "foo" {
			query = "echo foo"
		}
	}
}
`
	nodes, err := parse.Parse([]byte(sampleCase))
	require.NoError(t, err)
	assert.True(t, len(nodes) > 0)
	switchNode, err := control.NewSwitch(nodes[0], []byte(sampleCase))
	require.NoError(t, err)
	require.True(t, len(switchNode.Branches) > 0)
	parsedCase := switchNode.Branches[0]
	t.Run("sets the name", func(t *testing.T) {
		assert.Equal(t, "a", parsedCase.Name)
	})
	t.Run("sets the predicate", func(t *testing.T) {
		assert.Equal(t, "eq 1 0", parsedCase.Predicate)
	})
	t.Run("sets the inner node to the parsed inner node", func(t *testing.T) {
		expected, err := parse.Parse([]byte(sampleInnerHCL))
		require.NoError(t, err)
		assert.Equal(t, expected, parsedCase.InnerNodes)
		assert.True(t, reflect.DeepEqual(expected, parsedCase.InnerNodes))
	})
}

// TestParseSwitchConditionalWithDefault contains tests for parsing a default
// case statement
func TestParseSwitchConditionalWithDefault(t *testing.T) {
	var sampleDefault = `
switch "named-switch" {
	default {
		task.query "foo" {
			query = "echo foo"
		}
	}
}
`
	nodes, err := parse.Parse([]byte(sampleDefault))
	require.NoError(t, err)
	assert.True(t, len(nodes) > 0)
	switchNode, err := control.NewSwitch(nodes[0], []byte(sampleDefault))
	require.NoError(t, err)
	require.True(t, len(switchNode.Branches) > 0)
	parsedCase := switchNode.Branches[0]

	t.Run("sets the name", func(t *testing.T) {
		assert.Equal(t, "default", parsedCase.Name)
	})
	t.Run("sets the predicate", func(t *testing.T) {
		assert.Equal(t, "true", parsedCase.Predicate)
	})
	t.Run("sets the inner node to the parsed inner node", func(t *testing.T) {
		expected, err := parse.Parse([]byte(sampleInnerHCL))
		require.NoError(t, err)
		assert.Equal(t, expected, parsedCase.InnerNodes)
		assert.True(t, reflect.DeepEqual(expected, parsedCase.InnerNodes))
	})
}

// TestInnerText ensures that we correctly extract the text from the inside of a
// node.
func TestInnerText(t *testing.T) {
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
		actual, err := control.InnerText(nodes[0], []byte(sampleStatement))
		assert.NoError(t, err)
		assert.Equal(t, expected, string(actual))
	})
}
