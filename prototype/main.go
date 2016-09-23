package main

import (
	"fmt"
	"io/ioutil"

	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
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
		newData := processSwitch(oType)
		switches = append(switches, NewSwitchRange(oType, bytes, newData))
	}
	return switches
}

func processSwitch(obj *ast.ObjectType) []byte {
	var key string
	for _, item := range obj.List.Items {
		thisKey := item.Keys[0].Token.Value()
		if thisKey != "2 == 2" {
			continue
		}

	}
	return []byte(key)
}

func main() {
	data, err := ioutil.ReadFile("sample.hcl")
	if err != nil {
		fmt.Println(err)
		return
	}
	obj, err := hcl.ParseBytes(data)
	if err != nil {
		fmt.Println(err)
		return
	}

	switchStatements := switches(obj.Node.(*ast.ObjectList), data)

	for _, s := range switchStatements {
		fmt.Println("<<<<< preprocessed")
		fmt.Println(string(s.Data))
		fmt.Println(">>>>> substitution")
		fmt.Println(string(s.Replacement))
		fmt.Println("==================")
	}

	return
}
