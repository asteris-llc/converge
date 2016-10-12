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
