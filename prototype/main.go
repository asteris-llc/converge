package main

import (
	"bytes"
	"fmt"
	"strings"

	"io/ioutil"

	"github.com/asteris-llc/converge/render/extensions"
	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/hashicorp/hcl/hcl/printer"
)

func name(n *ast.ObjectItem) string {
	return n.Keys[len(n.Keys)-1].Token.Value().(string)
}

func tabs(i int) string {
	s := ""
	for i = i; i > 0; i-- {
		s = fmt.Sprintf("\t%s", s)
	}
	return s
}

func showObjectItem(oi *ast.ObjectType, bytes []byte) string {
	return string(bytes[oi.Lbrace.Offset:oi.Rbrace.Offset])
}

func printObjectItem(oi *ast.ObjectItem, pfx int, data []byte) {
	fmt.Printf("%sKeys: ", tabs(pfx))
	for _, k := range oi.Keys {
		fmt.Printf("%v;", k.Token.Value())
	}
	fmt.Println()
	fmt.Printf("%sValue Type: %T\n", tabs(pfx), oi.Val)

	if asObjectType, ok := oi.Val.(*ast.ObjectType); ok {
		for _, i := range asObjectType.List.Items {
			fmt.Printf("%ssub-item (%T)\n", tabs(pfx), i)
			printObjectItem(i, pfx+1, data)
		}
	}
}

type SwitchRange struct {
	Start       int
	End         int
	Data        []byte
	Replacement []byte
}

func (s SwitchRange) String() string {
	return string(s.Data)
}

func NewSwitchRange(obj *ast.ObjectType, bytes []byte, replacement []byte) SwitchRange {
	start := obj.Lbrace.Offset
	end := obj.Rbrace.Offset + 1
	return SwitchRange{
		Start:       start,
		End:         end,
		Data:        bytes[start:end],
		Replacement: replacement,
	}
}

func switches(objects *ast.ObjectList, bytes []byte) []SwitchRange {
	var switches []SwitchRange
	for _, item := range objects.Items {
		oType, ok := item.Val.(*ast.ObjectType)
		firstKey := item.Keys[0]
		if !ok || firstKey.Token.Value() != "switch" {
			continue
		}
		newData := processSwitch(oType, bytes)
		switches = append(switches, NewSwitchRange(oType, bytes, newData))
	}
	return switches
}

func processSwitch(obj *ast.ObjectType, bytes []byte) []byte {
	var key string
	for _, item := range obj.List.Items {
		var predicate string
		if isDefault(item) {
			predicate = "true"
		} else if isCase(item) {
			predicate = fmt.Sprintf("%v", item.Keys[2].Token.Value())
		} else {
			fmt.Println("not a valid case or default caluse")
			return []byte("")
		}
		if resolvePredicate(predicate) {
			key = string(bytes[item.Val.(*ast.ObjectType).Lbrace.Offset+1 : item.Val.(*ast.ObjectType).Rbrace.Offset])
			break
		}
	}
	return []byte(key)
}

func isSwitch(item *ast.ObjectItem) bool {
	return item.Keys[0].Token.Value() == "switch"
}

func isCase(item *ast.ObjectItem) bool {
	return item.Keys[0].Token.Value() == "case"
}

func isDefault(item *ast.ObjectItem) bool {
	return item.Keys[0].Token.Value() == "default"
}

func nodeIsSwitch(item ast.Node) bool {
	if asItem, ok := item.(*ast.ObjectItem); ok {
		return isSwitch(asItem)
	}
	return false
}

func parseSwitch(obj *ast.ObjectType, bytes []byte) ast.Node {
	parsed, _ := hcl.Parse(string(processSwitch(obj, bytes)))
	return parsed.Node
}

func parseSwitchItem(obj *ast.ObjectItem, bytes []byte) (ast.Node, error) {
	otype, ok := obj.Val.(*ast.ObjectType)
	if !ok {
		return nil, fmt.Errorf("value is not an ObjectType")
	}
	return parseSwitch(otype, bytes), nil
}

func showObjectListInfo(lst *ast.ObjectList) {
	fmt.Printf("[ ")
	for _, item := range lst.Items {
		fmt.Printf("%T ", item)
	}
	fmt.Printf("]\n")
}

func resolveFile(fname string) {
	data, err := ioutil.ReadFile(fname)
	if err != nil {
		fmt.Println(err)
		return
	}
	obj, err := hcl.ParseBytes(data)
	if err != nil {
		fmt.Println(err)
		return
	}
	processed := ast.Walk(obj.Node, func(n ast.Node) (ast.Node, bool) {
		asItem, ok := n.(*ast.ObjectItem)
		if !ok {
			return n, true
		}
		if isSwitch(asItem) {
			n1, err := parseSwitchItem(asItem, data)
			if err != nil {
				return n, false
			}
			l, ok := n1.(*ast.ObjectList)
			if !ok {
				return n1, false
			}
			if len(l.Items) > 0 {
				return l.Items[0], false
			}
			return &ast.ObjectList{Items: []*ast.ObjectItem{}}, false
		}
		return n, true
	})
	var output bytes.Buffer
	printer.Fprint(&output, processed)
	fmt.Println(output.String())
}

func resolvePredicate(pred string) bool {
	pred = strings.ToLower(pred)
	if pred == "true" {
		return true
	} else if pred == "false" {
		return false
	}

	unsupported := func(s ...string) (string, error) {
		return "", fmt.Errorf("unsupported call `%v` in switch template", s)
	}
	language := extensions.MinimalLanguage()
	language.On("param", unsupported)
	language.On(extensions.RefFuncName, unsupported)
	pred = fmt.Sprintf("{{ %s }}", pred)
	results, err := language.Render(struct{}{}, "predicate resolver", pred)
	if err != nil {
		fmt.Println("languag execution returned an error: ", err)
		return false
	}
	return (results.String() == "true")
}

func demoResolvePredicate() {
	fmt.Println(resolvePredicate("true"))
	fmt.Println(resolvePredicate("false"))
	fmt.Println(resolvePredicate(`eq "foo" "foo"`))
	fmt.Println(resolvePredicate(`eq "foo" "bar"`))
	fmt.Println(resolvePredicate(`or (eq "foo" "bar") (eq "foo" "foo")`))
}

func demoResolveFile() {
	resolveFile("sample.hcl")
}

func main() {
	demoResolveFile()
	return
}
