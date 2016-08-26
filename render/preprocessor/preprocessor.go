package preprocessor

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/asteris-llc/converge/graph"
)

// Preprocessor is a template preprocessor
type Preprocessor struct {
	vertices map[string]struct{}
}

// New creates a new preprocessor for the specified graph
func New(g *graph.Graph) *Preprocessor {
	m := make(map[string]struct{})
	for _, vertex := range g.Vertices() {
		m[vertex] = struct{}{}
	}
	return &Preprocessor{m}
}

// SplitTerms takes a string and splits it on '.'
func SplitTerms(in string) []string {
	return strings.Split(in, ".")
}

// Inits returns a list of heads of the string,
// e.g. [1,2,3] -> [[1,2,3],[1,2],[1]]
func Inits(in []string) [][]string {
	var results [][]string
	for i := 0; i < len(in); i++ {
		results = append([][]string{in[0 : i+1]}, results...)
	}
	return results
}

// HasField returns true if the provided struct has the defined field
func HasField(obj interface{}, fieldName string) bool {
	v := reflect.ValueOf(obj)
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return false
		}
		v = v.Elem()
	}
	_, hasField := v.Type().FieldByName(fieldName)
	return hasField
}

// HasMethod returns true if the provided struct supports the defined method
func HasMethod(obj interface{}, methodName string) bool {
	_, found := reflect.TypeOf(obj).MethodByName(methodName)
	return found
}

// EvalMember gets a member from a stuct, dereferencing pointers as necessary
func EvalMember(name string, obj interface{}) interface{} {
	v := reflect.ValueOf(obj)
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return fmt.Errorf("cannot dereference nil pointer of type %s\n", v.Type().String())
		}
		v = v.Elem()
	}

	if _, hasField := v.Type().FieldByName(name); !hasField {
		return fmt.Errorf("%s has no field named %s", v.Type().String(), name)
	}
	return v.FieldByName(name)
}
