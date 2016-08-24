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

	"github.com/spf13/pflag"
)

var (
	typeName     string
	fPath        string
	examplePath  string
	resourceName string
)

func init() {
	pflag.StringVar(&typeName, "type", "", "type to extract and document")
	pflag.StringVar(&fPath, "path", "", "source of Go file for extraction")
	pflag.StringVar(&resourceName, "resource-name", "", "name to import resource in HCL source")
	pflag.StringVar(&examplePath, "example", "", "name of example file to include")

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
		ExampleSource: out,
		ResourceName:  resourceName,
	}
	ast.Walk(extractor, file)

	// print example + info from parsed source
	fmt.Println(extractor)
}

type Field struct {
	Name, Type, Doc string
}

type TypeExtractor struct {
	Target string

	// information we get from the source
	TopDoc string
	Fields []*Field

	// information we get externally (from flags)
	ExampleSource []byte
	ResourceName  string
}

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

		te.TopDoc = te.Docs(n.Doc, spec.Doc, spec.Comment)

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
		te.Fields = append(
			te.Fields,
			&Field{
				Name: n.Names[0].String(),
				Type: fmt.Sprint(n.Type),
				Doc:  te.Docs(n.Doc, n.Comment),
			},
		)
		return nil

	default:
		return nil
	}
}

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
	// example
	out.WriteString("```hcl\n")
	out.Write(te.ExampleSource)
	out.WriteString("```\n\n")

	// docs
	out.WriteString(te.TopDoc)
	out.WriteString("\n")

	for _, field := range te.Fields {
		out.WriteString("- " + field.Name + " (`" + field.Type + "`)\n\n")
		out.WriteString("  " + strings.Replace(field.Doc, "\n", "   \n", -1))
		out.WriteString("\n")
	}

	return out.String()
}
