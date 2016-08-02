package extensions

import (
	"errors"
	"fmt"
	"strings"
	"text/template"

	"github.com/asteris-llc/converge/graph"
)

type TemplateExtension struct {
	GetDependencyFunc  func(*[]string) func(...string) (string, error)
	GetApplicationFunc func(...interface{}) func(...string) (interface{}, error)
}

var TemplateExtensions = map[string]TemplateExtension{
	"split": TemplateExtension{
		GetDependencyFunc: doesNotGenerateDependencies,
		GetApplicationFunc: contextFreeApplication(func(params ...string) (interface{}, error) {
			sep := params[0]
			str := params[1]
			return strings.Split(str, sep), nil
		}),
	},
	"param": TemplateExtension{
		GetDependencyFunc: func(out *[]string) func(...string) (string, error) {
			return func(params ...string) (string, error) {
				name := params[0]
				fmt.Println("Adding param dependency on param." + name)
				*out = append(*out, "param."+params[0])
				return params[0], nil
			}
		},
		GetApplicationFunc: func(contexts ...interface{}) func(...string) (interface{}, error) {
			renderGraph := contexts[0].(*graph.Graph)
			id := contexts[1].(string)
			return func(params ...string) (interface{}, error) {
				name := params[0]
				val := renderGraph.GetSibling(id, "param."+name)
				if val == nil {
					return "", errors.New("param not found")
				}
				return fmt.Sprintf("%+v", val), nil
			}
		},
	},
}

func doesNotGenerateDependencies(*[]string) func(...string) (string, error) {
	return func(params ...string) (string, error) {
		if len(params) > 0 {
			return "", nil
		}
		return "", nil
	}
}

func contextFreeApplication(f func(...string) (interface{}, error)) func(...interface{}) func(...string) (interface{}, error) {
	return func(...interface{}) func(...string) (interface{}, error) {
		return f
	}
}

func GetDependencyFuncMap(context *[]string) template.FuncMap {
	funcMap := template.FuncMap{}
	for k, v := range TemplateExtensions {
		funcMap[k] = v.GetDependencyFunc(context)
	}
	return funcMap
}

func GetApplicationFuncMap(contexts map[string][]interface{}) template.FuncMap {
	funcMap := template.FuncMap{}
	for k, v := range TemplateExtensions {
		context, found := contexts[k]
		if !found {
			context = []interface{}{}
		}
		funcMap[k] = v.GetApplicationFunc(context...)
	}
	return funcMap
}
