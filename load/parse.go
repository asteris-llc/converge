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
	"regexp"

	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/builtin/file"
	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
)

var (
	nameRe = regexp.MustCompile(`^[\w\-\.]+$`)
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
		names  = map[string]struct{}{}
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
				res, err = parseNamedResource(new(resource.ShellTask), item)
				if err != nil {
					break
				}

				// If no requirements are specified, it is assumed that the task before
				// the current task is the only requirement
				if previousTaskName != "" && !res.HasBaseDependencies() {
					res.SetDepends([]string{previousTaskName})
				}
				previousTaskName = res.String()

			case "module":
				res, err = parseModuleCall(item)

			case "template":
				res, err = parseNamedResource(new(resource.Template), item)

			case "param":
				res, err = parseNamedResource(new(resource.Param), item)

			case "file.mode":
				res, err = parseNamedResource(new(file.Mode), item)

			default:
				err = &ParseError{item.Pos(), fmt.Sprintf("unknown resource type %q", item.Keys[0].Token.Value())}
			}

			// check if any errors happened during parsing
			if err != nil {
				errs = append(errs, err)
				return n, false
			}

			// validate the name
			if !nameRe.MatchString(res.String()) {
				errs = append(errs, &ParseError{item.Pos(), fmt.Sprintf("invalid name %q", res.String())})
				return n, false
			}

			// check if the name is already present, error if so
			dupCheckName := res.String()
			if _, present := names[dupCheckName]; present {
				errs = append(errs, &ParseError{item.Pos(), fmt.Sprintf("duplicate %s %q", token, res.String())})
				return n, false
			}

			// Dependencies are always in the form `depends = [ "resource_type.name" ]`
			names[dupCheckName] = struct{}{}

			module.Resources = append(module.Resources, res)
			return n, false
		}

		return n, true
	})

	if len(errs) == 0 {
		return module, nil
	}
	return module, errs
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

func parseNamedResource(base resource.Resource, item *ast.ObjectItem) (resource.Resource, error) {
	if len(item.Keys) < 2 {
		return nil, &ParseError{
			item.Pos(),
			fmt.Sprintf(
				"%s has no name (expected `%s \"name\"`)",
				item.Keys[0].Token.Value(),
				item.Keys[0].Token.Value(),
			),
		}
	}

	err := hcl.DecodeObject(base, item.Val)
	base.SetName(item.Keys[1].Token.Value().(string))

	return base, err
}
