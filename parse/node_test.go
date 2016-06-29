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

package parse_test

import (
	"errors"
	"testing"

	"github.com/asteris-llc/converge/parse"
	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/stretchr/testify/assert"
)

func fromString(content string) (*parse.Node, error) {
	obj, err := hcl.ParseString(content)
	if err != nil {
		return nil, err
	}

	list, ok := obj.Node.(*ast.ObjectList)
	if !ok {
		return nil, errors.New("not an objectlist")
	}

	return &parse.Node{list.Items[0]}, nil
}

func validateTable(t *testing.T, input, errMsg string) {
	node, err := fromString(input)
	assert.NoError(t, err)

	err = node.Validate()
	if assert.Error(t, err) {
		assert.EqualError(t, err, errMsg)
	}
}

func TestNodeValidate(t *testing.T) {
	// everything about this should be valid
	t.Parallel()

	node, err := fromString(`task "x" {}`)
	assert.NoError(t, err)
	assert.NoError(t, node.Validate())
}

func TestNodeValidateNoName(t *testing.T) {
	// missing name, which is invalid
	t.Parallel()

	validateTable(t, `x {}`, "1:1: missing name")
}

func TestNodeValidateModuleMissingNameOrSource(t *testing.T) {
	// missing name/source in a module call, which is invalid
	t.Parallel()

	validateTable(t, `module x {}`, "1:1: missing source or name in module call")
}

func TestNodeValidateTooManyKeys(t *testing.T) {
	// too many keys is a problem
	t.Parallel()

	validateTable(t, `task x y {}`, "1:1: too many keys")
}

func TestNodeValidateTooManyKeysModule(t *testing.T) {
	// too many keys is a problem in modules too!
	t.Parallel()

	validateTable(t, `module x y z {}`, "1:1: too many keys")
}

func TestNodeKind(t *testing.T) {
	t.Parallel()

	node, err := fromString(`task "x" {}`)
	assert.NoError(t, err)
	assert.Equal(t, "task", node.Kind())
}

func TestNodeName(t *testing.T) {
	t.Parallel()

	node, err := fromString(`task "x" {}`)
	assert.NoError(t, err)
	assert.Equal(t, "x", node.Name())
}

func TestNodeIsModule(t *testing.T) {
	t.Parallel()

	node, err := fromString(`module "source" "name" {}`)
	assert.NoError(t, err)
	assert.True(t, node.IsModule())
}

func TestNodeIsntModule(t *testing.T) {
	t.Parallel()

	node, err := fromString(`task "name" {}`)
	assert.NoError(t, err)
	assert.False(t, node.IsModule())
}

func TestNodeSource(t *testing.T) {
	t.Parallel()

	node, err := fromString("module x y {}")
	assert.NoError(t, err)
	assert.Equal(t, "x", node.Source())
}
