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

package load

import (
	"fmt"

	"github.com/asteris-llc/converge/resource"
	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
)

// Parse parses a module and returns it
func Parse(content []byte) (*resource.Module, error) {
	f, err := hcl.ParseBytes(content)
	if err != nil {
		return nil, err
	}

	return parseModule(f.Node)
}

func parseModule(node ast.Node) (*resource.Module, error) {
	// the approach were taking here is to create some state that we'll manage
	// locally, and then walk over the nodes in the AST, gathering errors as we
	// go. This is also the point at which we enforce module-level semantic
	// checks, such as erroring out on duplicate param or resource names.
	var (
		errs   MultiError
		module = new(resource.Module)
		names  = map[string]bool{}
	)
	previousTaskName := ""
	ast.Walk(node, func(n ast.Node) (ast.Node, bool) {
		// we're only interested in ObjectItems. These are a path plus a value, and
		// quite handy.
		if item, ok := n.(*ast.ObjectItem); ok {
			token := item.Keys[0].Token.Text
			var (
				res resource.Resource
				err error
			)

			switch token {
			case "task":
				res, err = parseTask(item)

			case "template":
				res, err = parseTemplate(item)

			case "module":
				res, err = parseModuleCall(item)

			case "param":
				res, err = parseParam(item)

			default:
				err = &ParseError{item.Pos(), fmt.Sprintf("unknown resource type %q", item.Keys[0].Token.Value())}
			}

			// check if any errors happened during parsing
			if err != nil {
				errs = append(errs, err)
				return n, false
			}

			// check if the name is already present, error if so
			dupCheckName := token + "." + res.Name()
			if present := names[dupCheckName]; present {
				errs = append(errs, &ParseError{item.Pos(), fmt.Sprintf("duplicate %s %q", token, res.Name())})
				return n, false
			}
			names[dupCheckName] = true
			// now that we've run the gauntlet, it's safe to add the resource to the
			// resource list.
			if token == "task" {
				if previousTaskName != "" {
					task := res.(resource.Task)
					task.AddDep(previousTaskName)
				}
				previousTaskName = dupCheckName
			}
			module.Resources = append(module.Resources, res)

			return n, false
		}
		return n, true
	})
	//Check that all dependencies were are resources in this module
	for _, r := range module.Resources {
		dependencies := r.Depends()
		for _, dep := range dependencies {
			if _, present := names[dep]; !present {
				errs = append(errs, &ParseError{node.(*ast.ObjectList).Pos(),
					fmt.Sprintf("Resource %q depends on resource, %q,  which does not exist in this module", r.Name(), dep)})
			}
		}
	}
	if len(errs) == 0 {
		return module, nil
	}
	return module, errs
}

func parseParam(item *ast.ObjectItem) (p *resource.Param, err error) {
	/*
		ideal input:

		param "x" { default = "y" }
	*/
	if len(item.Keys) < 2 {
		err = &ParseError{item.Pos(), "param has no name (expected `param \"name\"`)"}
		return
	}

	p = &resource.Param{
		ParamName: item.Keys[1].Token.Value().(string),
	}
	err = hcl.DecodeObject(p, item.Val)
	return
}

func parseTask(item *ast.ObjectItem) (t *resource.ShellTask, err error) {
	/*
		ideal input:

		task "x" {
			check = "y"
			apply = "z"
			depends = [""]
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
			depends = [""]
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
			args = {
				arg1 = 1
			}
			depends = [""]
		}
	*/
	if len(item.Keys) < 3 {
		err = &ParseError{item.Pos(), "module missing source or name (expected `module \"source\" \"name\"`)"}
		return
	}

	module = &resource.ModuleTask{
		Args: resource.Values{},
	}
	err = hcl.DecodeObject(&module, item.Val)
	module.Source = item.Keys[1].Token.Value().(string)
	module.ModuleName = item.Keys[2].Token.Value().(string)
	return module, err
}
