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

package module

import (
	"fmt"

	"github.com/asteris-llc/converge/resource"
	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
)

// New parses a module and returns it
func New(content []byte) (*Module, error) {
	module := &Module{
		Params: map[string]*Param{},
	}
	var errs MultiError

	f, err := hcl.ParseBytes(content)
	if err != nil {
		return nil, err
	}

	ast.Walk(f.Node, func(n ast.Node) (ast.Node, bool) {
		if item, ok := n.(*ast.ObjectItem); ok {
			switch item.Keys[0].Token.Text {
			case "param":
				id, param, err := parseParam(item)
				if err != nil {
					errs = append(errs, err)
				} else if _, present := module.Params[id]; !present {
					module.Params[id] = param
				} else {
					errs = append(errs, &ParseError{item.Pos(), fmt.Sprintf("duplicate param %q", id)})
				}

			case "task":
				task, err := parseTask(item)
				if err != nil {
					errs = append(errs, err)
				} else {
					module.Resources = append(module.Resources, task)
				}

			default:
				fmt.Println(item.Keys[0].Token)
			}

			return n, false
		}

		return n, true
	})

	return module, errs
}

func parseParam(item *ast.ObjectItem) (id string, p *Param, err error) {
	/*
		ideal input:

		param "x" { default = "y" }
	*/
	if len(item.Keys) < 2 {
		err = &ParseError{item.Pos(), "param has no name (expected `param \"name\"`)"}
		return
	}

	if pID, ok := item.Keys[1].Token.Value().(string); ok {
		id = pID
	} else {
		err = &ParseError{item.Pos(), fmt.Sprintf("param needs a string name (have %s)", item.Keys[1].Token.Type)}
		return
	}

	p = new(Param)
	err = hcl.DecodeObject(p, item.Val)
	return
}

func parseTask(item *ast.ObjectItem) (t *resource.ShellTask, err error) {
	/*
		ideal input:

		task "x" {
		  check = "y"
		  apply = "z"
		}
	*/
	if len(item.Keys) < 2 {
		err = &ParseError{item.Pos(), "task has no name (expected `task \"name\"`)"}
		return
	}

	id, ok := item.Keys[1].Token.Value().(string)
	if !ok {
		err = &ParseError{item.Pos(), fmt.Sprintf("task needs a string name (have %s)", item.Keys[1].Token.Type)}
		return
	}

	t = new(resource.ShellTask)
	t.Name = id
	err = hcl.DecodeObject(t, item.Val)

	return
}
