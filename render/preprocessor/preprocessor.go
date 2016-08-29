package preprocessor

import (
	"bytes"
	"errors"
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

// JoinTerms takes a list of terms and joins them with '.'
func JoinTerms(s []string) string {
	return strings.Join(s, ".")
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

// Prefixes returns a set of prefixes for a string, e.g. "a.b.c.d" will yield
// []string{"a.b.c.d","a.b.c","a.b.","a"}
func Prefixes(in string) (out []string) {
	for _, termSet := range Inits(SplitTerms(in)) {
		out = append(out, JoinTerms(termSet))
	}
	return out
}

// Find returns the first element of the string slice for which f returns true
func Find(slice []string, f func(string) bool) (string, bool) {
	for _, elem := range slice {
		if f(elem) {
			return elem, true
		}
	}
	return "", false
}

// MkCallPipeline transforms a term group (b.c.d) into a pipeline (b | c | d)
func MkCallPipeline(s string) string {
	return strings.Join(SplitTerms(s), " | ")
}

// DesugarCall takes a call in the form of "a.b.c.d" and returns a desugared
// string that will work with the language extension provided by calling
// .Language()
func DesugarCall(g *graph.Graph, call string) (string, error) {
	var out bytes.Buffer
	pfx, rest, found := VertexSplit(g, call)
	if !found {
		return "", errors.New("syntax error call to non-existant dependency")
	}
	out.WriteString(fmt.Sprintf("(noderef %q)", pfx))
	if rest != "" {
		out.WriteString(fmt.Sprintf("| %s", MkCallPipeline(rest)))
	}
	return out.String(), nil
}

// VertexSplit takes a graph with a set of vertexes and a string, and returns
// the longest vertex id from the graph and the remainder of the string.  If no
// matching vertex is found 'false' is returned.
func VertexSplit(g *graph.Graph, s string) (string, string, bool) {
	prefix, found := Find(Prefixes(s), g.Contains)
	if !found {
		return "", s, false
	}
	if prefix == s {
		return prefix, "", true
	}
	return prefix, s[len(prefix)+1:], true
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
func EvalMember(name string, obj interface{}) (reflect.Value, error) {
	v := reflect.ValueOf(obj)
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return reflect.Zero(reflect.TypeOf(obj)), nilPtrError(v)
		}
		v = v.Elem()
	}

	if _, hasField := v.Type().FieldByName(name); !hasField {
		return reflect.Zero(reflect.TypeOf(obj)), missingFieldError(name, v)
	}
	return v.FieldByName(name), nil
}

func nilPtrError(v reflect.Value) error {
	typeStr := v.Type().String()
	return fmt.Errorf("cannot dereference nil pointer of type %T", typeStr)
}

func missingFieldError(name string, v reflect.Value) error {
	return fmt.Errorf("%s has no field named %s", v.Type().String(), name)
}
