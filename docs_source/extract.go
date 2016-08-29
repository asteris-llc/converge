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

package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"strings"
	"text/template"

	"github.com/spf13/pflag"
)

var (
	typeName      string
	fPath         string
	examplePath   string
	resourceName  string
	stripDocLines int

	tmpl = template.Must(template.New("").Funcs(template.FuncMap{"fencedCode": fencedCode}).Parse(`
{{.TopDoc}}

## Example

{{fencedCode .ExampleSource}}

## Parameters
{{ range .Fields}}
- {{.Name}} ({{.Type}})

  {{.Doc}}{{end}}
`))
)

func init() {
	pflag.StringVar(&typeName, "type", "", "type to extract and document")
	pflag.StringVar(&fPath, "path", "", "source of Go file for extraction")
	pflag.StringVar(&resourceName, "resource-name", "", "name to import resource in HCL source")
	pflag.StringVar(&examplePath, "example", "", "name of example file to include")
	pflag.IntVar(&stripDocLines, "strip-doc-lines", 0, "strip this many lines of docs from the type - so it doesn't all have to start with \"ModuleName blah blah...\"")

	pflag.Parse()
}

func main() {
	// read example file
	out, err := ioutil.ReadFile(examplePath)
	if err != nil {
		log.Fatal(err)
	}

	// read source file
	fset := token.NewFileSet()
	file, err := parser.ParseFile(
		fset,
		fPath,
		nil,
		parser.ParseComments,
	)
	if err != nil {
		log.Fatal(err)
	}

	extractor := &TypeExtractor{
		Target:        typeName,
		ExampleSource: string(out),
		ResourceName:  resourceName,
	}
	ast.Walk(extractor, file)

	// print example + info from parsed source
	fmt.Println(extractor)
}

// Field represents a documentation field
type Field struct {
	Name, Type, Doc string
}

// TypeExtractor extracts documentation information
type TypeExtractor struct {
	Target string

	// information we get from the source
	TopDoc string
	Fields []*Field

	// information we get externally (from flags)
	ExampleSource string
	ResourceName  string
}

// Visit inspects each node in the Ast and returns a Visitor
func (te *TypeExtractor) Visit(node ast.Node) (w ast.Visitor) {
	if node == nil {
		return nil
	}

	switch n := node.(type) {

	case *ast.File:
		// don't care, but recurse
		return te

	case *ast.GenDecl:
		spec, ok := n.Specs[0].(*ast.TypeSpec)
		if !ok {
			return nil
		}

		if spec.Name.Name != te.Target {
			return nil
		}

		te.TopDoc = stripLines(te.Docs(n.Doc, spec.Doc, spec.Comment))

		return te

	case *ast.TypeSpec:
		// we've taken care of this as a field in the *ast.GenDecl case, just recurse here
		return te

	case *ast.StructType:
		if n.Fields == nil && n.Incomplete {
			return nil
		}

		return te

	case *ast.FieldList:
		// recurse to walk over the fields
		return te

	case *ast.Field:
		field := &Field{
			Name: n.Names[0].String(),
			Type: stringify(n.Type),
			Doc:  te.Docs(n.Doc, n.Comment),
		}

		if n.Tag != nil {
			for k, v := range parseTag(strings.Trim(n.Tag.Value, "`")) {
				switch k {
				case "hcl":
					field.Name = fmt.Sprintf("`%s`", v[0])

				case "doc_type":
					field.Type = v[0]
				}
			}
		}

		te.Fields = append(te.Fields, field)

		return nil

	default:
		return nil
	}
}

// Docs generates a documentation string
func (*TypeExtractor) Docs(gs ...*ast.CommentGroup) string {
	var out []string
	for _, g := range gs {
		if g != nil {
			out = append(out, g.Text())
		}
	}

	return strings.Join(out, "\n\n")
}

func (te *TypeExtractor) String() string {
	var out bytes.Buffer

	err := tmpl.Execute(&out, te)
	if err != nil {
		log.Fatal(err)
	}

	return out.String()
}

func stripLines(doc string) string {
	lines := strings.Split(doc, "\n")
	return strings.Join(lines[stripDocLines:], "\n")
}

func parseTag(tag string) map[string][]string {
	out := map[string][]string{}

	for len(tag) > 0 {
		var key bytes.Buffer

		// consume whitespace
		tag = strings.TrimLeft(tag, " ")

		// parse tag
		for _, b := range tag {
			tag = tag[1:]

			if b == ':' {
				break
			}
			key.WriteRune(b)
		}

		// consume quote
		tag = strings.TrimLeft(tag, "\"")

		// start consuming keys
		var value bytes.Buffer
		for _, b := range tag {
			tag = tag[1:]

			if b == '"' || b == ',' {
				out[key.String()] = append(out[key.String()], value.String())
				value.Reset()
			}

			if b == '"' {
				break
			} else if b == ',' {
				continue
			}

			value.WriteRune(b)
		}
	}

	return out
}

func fencedCode(in string) string {
	return fmt.Sprintf("```hcl\n%s\n```\n", in)
}

func stringify(node ast.Expr) string {
	switch n := node.(type) {
	case *ast.Ident:
		return n.Name

	case *ast.ArrayType:
		return fmt.Sprintf("list of %ss", stringify(n.Elt))

	case *ast.MapType:
		return fmt.Sprintf("map of %s to %s", stringify(n.Key), stringify(n.Value))

	case *ast.InterfaceType:
		return "anything"

	case *ast.StarExpr:
		return fmt.Sprintf("optional %s", stringify(n.X))

	default:
		return fmt.Sprintf("%T", n)
	}
}
