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
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
)

// Parse content into a bunch of nodes
func Parse(content []byte) (resources []*Node, err error) {
	obj, err := hcl.ParseBytes(content)
	if err != nil {
		return resources, err
	}

	ast.Walk(obj.Node, func(n ast.Node) (ast.Node, bool) {
		baseItem, ok := n.(*ast.ObjectItem)
		if !ok {
			return n, true
		}

		item := &Node{baseItem}

		if itemErr := item.Validate(); itemErr != nil {
			err = multierror.Append(err, itemErr)
			return n, false
		}

		resources = append(resources, item)

		return n, false
	})

	return resources, err
}
