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

package preprocessor

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/asteris-llc/converge/graph"
)

// ErrUnresolvable indicates that a field exists but is unresolvable due to nil
// references
var ErrUnresolvable = errors.New("field is unresolvable")

// fieldMapCache caches the results of field map generation to avoid
// recalculating it during execution.
var fieldMapCache = make(map[reflect.Type]map[string]string)

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
	var v reflect.Type
	switch oType := obj.(type) {
	case reflect.Type:
		v = oType
	case reflect.Value:
		v = oType.Type()
	default:
		v = reflect.TypeOf(obj)
	}
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	fieldName, err := LookupCanonicalFieldName(v, fieldName)
	if err != nil {
		return false
	}
	_, hasField := v.FieldByName(fieldName)
	return hasField
}

// ListFields returns a list of fields for the struct
func ListFields(obj interface{}) ([]string, error) {
	var results []string
	var v reflect.Type
	switch oType := obj.(type) {
	case reflect.Type:
		v = oType
	case reflect.Value:
		v = oType.Type()
	default:
		v = reflect.TypeOf(obj)
	}
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	e := reflect.Zero(v)
	if reflect.Struct != e.Kind() {
		return results, fmt.Errorf("element is %s, not a struct", e.Type())
	}
	for idx := 0; idx < e.Type().NumField(); idx++ {
		field := e.Type().Field(idx)
		results = append(results, field.Name)
	}
	return results, nil
}

// HasMethod returns true if the provided struct supports the defined method
func HasMethod(obj interface{}, methodName string) bool {
	_, found := reflect.TypeOf(obj).MethodByName(methodName)
	return found
}

// EvalMember gets a member from a stuct, dereferencing pointers as necessary
func EvalMember(name string, obj interface{}) (reflect.Value, error) {
	name, err := LookupCanonicalFieldName(interfaceToConcreteType(obj), name)

	if err != nil {
		return reflect.Zero(reflect.TypeOf(obj)), err
	}
	v := reflect.ValueOf(obj)
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
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

// HasPath returns true of the set of terms can resolve to a value
func HasPath(obj interface{}, terms ...string) error {
	t := reflect.TypeOf(obj)
	for _, term := range terms {
		for t.Kind() == reflect.Ptr {
			t = t.Elem()
		}

		if k := t.Kind(); k == reflect.Interface {
			return nil
		} else if k != reflect.Struct {
			return errors.New("cannot access non-structure field")
		}

		term, err := LookupCanonicalFieldName(t, term)
		if err != nil {
			return err
		}

		field, ok := t.FieldByName(term)
		if !ok {
			validFields, fieldErrs := ListFields(t)
			if fieldErrs != nil {
				return fieldErrs
			}
			return fmt.Errorf("term should be one of %v not %q", mapToLower(validFields), term)
		}
		t = field.Type
	}
	return nil
}

// EvalTerms acts as a left fold over a list of term accessors
func EvalTerms(obj interface{}, terms ...string) (interface{}, error) {
	if err := HasPath(obj, terms...); err != nil {
		return nil, err
	}

	v := reflect.ValueOf(obj)
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return reflect.Zero(reflect.TypeOf(obj)), ErrUnresolvable
		}
		v = v.Elem()
	}

	for _, term := range terms {
		if HasField(obj, term) {
			val, err := EvalMember(term, obj)
			if err != nil {
				return nil, ErrUnresolvable
			}
			obj = val.Interface()
		} else {
			return nil, ErrUnresolvable
		}
	}
	return obj, nil
}

// For a given interface, fieldMap returns a map with keys being the lowercase
// versions of the string, and values being the correct version.  It returns an
// error if the interface is not a struct, or a reflect.Type or reflect.Value of
// a struct.
func fieldMap(val interface{}) (map[string]string, error) {
	fieldMap := make(map[string]string)
	var t reflect.Type
	switch typed := val.(type) {
	case reflect.Type:
		t = typed
	case reflect.Value:
		t = typed.Type()
	default:
		t = reflect.TypeOf(val)
	}
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("cannot access fields of non-struct type %T", val)
	}
	return addFieldsToMap(fieldMap, t)
}

func addFieldsToMap(m map[string]string, t reflect.Type) (map[string]string, error) {
	if cached, ok := fieldMapCache[t]; ok {
		return cached, nil
	}

	for idx := 0; idx < t.NumField(); idx++ {
		field := t.Field(idx)
		if field.Anonymous {
			var err error
			anonType := interfaceToConcreteType(field.Type)
			if anonType.Kind() == reflect.Struct {
				if m, err = addFieldsToMap(m, anonType); err != nil {
					return nil, err
				}
			}
			continue
		}

		name := field.Name
		lower := strings.ToLower(name)
		if _, ok := m[lower]; ok {
			return nil, fmt.Errorf("multiple potential matches for %s", name)
		}
		m[lower] = name
	}
	fieldMapCache[t] = m
	return m, nil
}

// LookupCanonicalFieldName takes a type and an arbitrarily cased field name and
// returns the field name with a case that matches the actual field.
func LookupCanonicalFieldName(t reflect.Type, term string) (string, error) {
	term = strings.ToLower(term)
	m, err := fieldMap(t)
	if err != nil {
		return "", err
	}
	correctCase, found := m[term]
	if found {
		return correctCase, nil
	}
	var fields []string
	for key := range m {
		fields = append(fields, key)
	}
	return "", fmt.Errorf("%s has no field that matches %s, should be one of %v", t, term, fields)
}

func interfaceToConcreteType(i interface{}) reflect.Type {
	var t reflect.Type
	switch typed := i.(type) {
	case reflect.Type:
		t = typed
	case reflect.Value:
		t = typed.Type()
	default:
		t = reflect.TypeOf(i)
	}
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}

// mapToLower converts a string slice to all lower case
func mapToLower(strs []string) []string {
	for idx, str := range strs {
		strs[idx] = strings.ToLower(str)
	}
	return strs
}

func nilPtrError(v reflect.Value) error {
	typeStr := v.Type().String()
	return fmt.Errorf("cannot dereference nil pointer of type %s", typeStr)
}

func missingFieldError(name string, v reflect.Value) error {
	return fmt.Errorf("%s has no field named %s", v.Type().String(), name)
}
