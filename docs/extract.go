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

	"reflect"

	"github.com/spf13/pflag"
)

const durationComment = `
Acceptable formats are a number in seconds or a duration string. A Duration
represents the elapsed time between two instants as an int64 second count.
The representation limits the largest representable duration to approximately
290 years. A duration string is a possibly signed sequence of decimal numbers,
each with optional fraction and a unit suffix, such as "300ms", "-1.5h" or
"2h45m". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
`

var (
	typeName      string
	fPath         string
	taskPath      string
	examplePath   string
	resourceName  string
	taskName      string
	stripDocLines int

	tmpl = template.Must(template.New("").Funcs(template.FuncMap{
		"fencedCode": fencedCode,
		"code":       func(x string) string { return "`" + x + "`" },
		"codeCommaJoin": func(items []string, terminal string) string {
			var out string
			if len(items) == 2 && terminal != "" {
				return fmt.Sprintf("`%s` %s `%s`", items[0], terminal, items[1])
			}
			for i, item := range items {
				out += "`" + item + "`"
				if i+1 != len(items) {
					out += ", "
				}
				if i == len(items)-2 {
					out += terminal + " "
				}
			}

			return out
		},
	}).Parse(`
{{.TopDoc}}

## Example

{{fencedCode .ExampleSource}}

## Parameters
{{ range .Fields}}
- {{.Name}} ({{if .Required}}required {{end}}{{if ne .Base ""}}base {{.Base}} {{end}}{{.Type}})

{{ if .MutuallyExclusive}}
	Only one of {{codeCommaJoin .MutuallyExclusive "or"}} may be set.

{{end}}{{ if .ValidValues}}
	Valid values: {{codeCommaJoin .ValidValues "and"}}

{{end}}{{if ne .Doc ""}}  {{.Doc}}{{end}}
{{end}}
{{ if .HasExportedFields }}
## Exported Fields

{{- range .GetExported}}
- {{.ExportedAs}} ({{.Type}})
{{if ne .Doc ""}}  {{.Doc}}{{end}} {{end}}
{{- range .GetReExported}}
- {{.ExportedAs}} re-exports fields from {{.Type}}
{{if ne .Doc ""}}  {{.Doc}}{{end}} {{end}} {{end}}
`))
)

func init() {
	pflag.StringVar(&typeName, "type", "", "type to extract and document")
	pflag.StringVar(&taskName, "task", "", "type of the resource to extract and document")
	pflag.StringVar(&fPath, "path", "", "source of Go file for extraction")
	pflag.StringVar(&taskPath, "task-path", "", "source file for the task for extraction")
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

	if taskPath != "" && taskName != "" {
		fset = token.NewFileSet()
		file, err = parser.ParseFile(
			fset,
			taskPath,
			nil,
			parser.ParseComments,
		)
		if err != nil {
			log.Fatal(err)
		}

		fset = token.NewFileSet()
		file, err = parser.ParseFile(
			fset,
			taskPath,
			nil,
			parser.ParseComments,
		)

		if err != nil {
			log.Fatal(err)
		}

		exportedExtractor := &ExportExtractor{
			Target: taskName,
		}

		ast.Walk(exportedExtractor, file)
		extractor.ExportedFields = exportedExtractor
	}

	fmt.Println(extractor)
}

// Field represents a documentation field
type Field struct {
	Name              string
	Type              string
	Doc               string
	Required          bool
	Base              string
	MutuallyExclusive []string
	ValidValues       []string
	ExportedAs        string
}

// ExportExtractor handles exported data docs
type ExportExtractor struct {
	Target           string
	Doc              string
	ExportedFields   []*Field
	ReExportedFields []*Field
}

// Visit visits a node and logs exported fields
func (e *ExportExtractor) Visit(node ast.Node) (w ast.Visitor) {
	switch n := node.(type) {
	case *ast.File, *ast.TypeSpec, *ast.FieldList:
		return e
	case *ast.StructType:
		if n.Fields == nil && n.Incomplete {
			return nil
		}
		return e
	case *ast.GenDecl:
		spec, ok := n.Specs[0].(*ast.TypeSpec)
		if !ok {
			return nil
		}

		if spec.Name.Name != e.Target {
			return nil
		}
		return e
	case *ast.Field:
		typ := stringify(n.Type, "")
		doc := (&TypeExtractor{}).Docs(n.Doc, n.Comment)
		field := &Field{
			Name: n.Names[0].String(),
			Type: typ,
			Doc:  doc,
		}
		fmt.Println("Got a field: ", n.Names[0].String())
		if n.Tag != nil {
			tag := reflect.StructTag(strings.Trim(n.Tag.Value, "`"))
			fmt.Println("\t tag: ", tag)
			if export, ok := tag.Lookup("export"); ok {
				fmt.Println("\t\t Got an 'export' tag")
				field.ExportedAs = export
				field.ExportedAs = fmt.Sprintf("`%s`", strings.SplitN(export, ",", 1)[0])
				e.ExportedFields = append(e.ExportedFields, field)
				fmt.Println("\t\t exported fields now: ", e.ExportedFields)
			} else if export, ok := tag.Lookup("re-export-as"); ok {
				fmt.Println("\t\t Got a 're-export-as' tag")
				field.ExportedAs = fmt.Sprintf("`%s`", strings.SplitN(export, ",", 1)[0])
				fmt.Println("\t\t exported fields now: ", e.ExportedFields)
				e.ReExportedFields = append(e.ReExportedFields, field)
			} else {
				fmt.Println("\t\t No export tag")
			}
		}
		return e
	default:
		return nil
	}
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

	// information about exported fields
	ExportedFields *ExportExtractor
}

// HasExportedFields returns true if any fields are exported
func (te *TypeExtractor) HasExportedFields() bool {
	return te.ExportedFields != nil
}

// GetExported gets the exported fields
func (te *TypeExtractor) GetExported() []*Field {
	return te.ExportedFields.ExportedFields
}

// GetReExported gets the re-exported fields
func (te *TypeExtractor) GetReExported() []*Field {
	return te.ExportedFields.ReExportedFields
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
		typ := stringify(n.Type, "optional")
		doc := te.Docs(n.Doc, n.Comment)
		if strings.Contains(typ, "duration") {
			doc += durationComment
		}
		field := &Field{
			Name: n.Names[0].String(),
			Type: typ,
			Doc:  doc,
		}

		if n.Tag != nil {
			tag := reflect.StructTag(strings.Trim(n.Tag.Value, "`"))
			if hcl, ok := tag.Lookup("hcl"); ok {
				field.Name = fmt.Sprintf("`%s`", strings.SplitN(hcl, ",", 1)[0])
			}

			if docType, ok := tag.Lookup("doc_type"); ok {
				field.Type = docType
			}

			if base, ok := tag.Lookup("base"); ok {
				field.Base = base
			}

			if required, ok := tag.Lookup("required"); ok && required == "true" {
				field.Required = true
			}

			if mutuallyexclusive, ok := tag.Lookup("mutually_exclusive"); ok {
				field.MutuallyExclusive = strings.Split(mutuallyexclusive, ",")
			}

			if validvalues, ok := tag.Lookup("valid_values"); ok {
				field.ValidValues = strings.Split(validvalues, ",")
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

func fencedCode(in string) string {
	return fmt.Sprintf("```hcl\n%s\n```\n", in)
}

// stringify a type expression to a human-readable format.  The pointersAs
// string will be used to prefix a pointer type.
func stringify(node ast.Expr, pointersAs string) string {
	switch n := node.(type) {
	case *ast.Ident:
		return n.Name

	case *ast.ArrayType:
		return fmt.Sprintf("list of %ss", stringify(n.Elt, pointersAs))

	case *ast.MapType:
		return fmt.Sprintf("map of %s to %s", stringify(n.Key, pointersAs), stringify(n.Value, pointersAs))

	case *ast.InterfaceType:
		return "anything"

	case *ast.StarExpr:
		return fmt.Sprintf("%s %s", pointersAs, stringify(n.X, pointersAs))

	case *ast.SelectorExpr:
		selExp := fmt.Sprintf("%s.%s", stringify(n.X, pointersAs), stringify(n.Sel, pointersAs))
		switch selExp {
		case "time.Duration":
			return "duration"
		case "pkg.State":
			return "State"
		case "resource.Value":
			return "anything"
		default:
			return selExp
		}

	default:
		return fmt.Sprintf("%T", n)
	}
}
