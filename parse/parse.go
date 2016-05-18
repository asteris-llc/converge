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
	"fmt"

	"github.com/asteris-llc/converge/resource"
	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
)

// New parses a module and returns it
func New(content []byte) (*resource.Module, error) {
	f, err := hcl.ParseBytes(content)
	if err != nil {
		return nil, err
	}

	return parseModule(f.Node)
}

func parseModule(node ast.Node) (*resource.Module, error) {
	module := &resource.Module{
		Params: map[string]*resource.Param{},
	}
	var errs MultiError

	ast.Walk(node, func(n ast.Node) (ast.Node, bool) {
		if item, ok := n.(*ast.ObjectItem); ok {
			token := item.Keys[0].Token.Text
			if token == "param" {
				id, param, err := parseParam(item)
				if err != nil {
					errs = append(errs, err)
				} else if _, present := module.Params[id]; !present {
					module.Params[id] = param
				} else {
					errs = append(errs, &ParseError{item.Pos(), fmt.Sprintf("duplicate param %q", id)})
				}
			} else {
				var (
					resource resource.Resource
					err      error
				)

				switch token {
				case "task":
					resource, err = parseTask(item)

				case "template":
					resource, err = parseTemplate(item)

				case "module":
					resource, err = parseModuleCall(item)

				default:
					err = &ParseError{item.Pos(), fmt.Sprintf("unknown resource type %q", item.Keys[0].Token.Value())}
				}

				if err != nil {
					errs = append(errs, err)
				} else {
					module.Resources = append(module.Resources, resource)
				}
			}
			return n, false
		}

		return n, true
	})

	return module, errs
}

func parseParam(item *ast.ObjectItem) (id string, p *resource.Param, err error) {
	/*
		ideal input:

		param "x" { default = "y" }
	*/
	if len(item.Keys) < 2 {
		err = &ParseError{item.Pos(), "param has no name (expected `param \"name\"`)"}
		return
	}

	id = item.Keys[1].Token.Value().(string)

	p = new(resource.Param)
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

	t = new(resource.ShellTask)
	t.TaskName = item.Keys[1].Token.Value().(string)
	err = hcl.DecodeObject(t, item.Val)

	return
}

func parseTemplate(item *ast.ObjectItem) (t *resource.Template, err error) {
	/*
		ideal input:

		template "x" {
		  content = "y"
		  destination = "z"
		}
	*/
	if len(item.Keys) < 2 {
		err = &ParseError{item.Pos(), "template has no name (expected `template \"name\"`)"}
		return
	}

	t = new(resource.Template)
	t.TemplateName = item.Keys[1].Token.Value().(string)
	err = hcl.DecodeObject(t, item.Val)

	return
}

func parseModuleCall(item *ast.ObjectItem) (module *resource.ModuleTask, err error) {
	/*
		ideal input:

		module "source" "name" {
		  args = 1
		}
	*/
	if len(item.Keys) < 3 {
		err = &ParseError{item.Pos(), "module missing source or name (expected `module \"source\" \"name\"`)"}
		return
	}

	module = &resource.ModuleTask{
		Args:       map[string]interface{}{},
		Source:     item.Keys[1].Token.Value().(string),
		ModuleName: item.Keys[2].Token.Value().(string),
	}
	err = hcl.DecodeObject(&module.Args, item.Val)

	return
}
