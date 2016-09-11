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
	fmt.Printf("checking for field %s on %T\n", fieldName, obj)
	fieldName = toPublicFieldCase(fieldName)
	var v reflect.Type
	switch oType := obj.(type) {
	case reflect.Type:
		v = oType
	case reflect.Value:
		v = oType.Type()
	default:
		v = reflect.TypeOf(obj)
	}

	fmt.Printf("iterating over kind %v\n", v.Kind())
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	fmt.Printf("finished iterating, ensuring is struct\n")
	if v.Kind() != reflect.Struct {
		fmt.Printf("is not a struct, instead %v\n", v.Kind())
		return false
	}
	fmt.Printf("is a struct, getting field by name...\n")
	_, hasField := v.FieldByName(fieldName)
	fmt.Printf("hasfield: %v\n", hasField)
	return hasField
}

// MethodType gets the type of a method based on the object and method name
func MethodType(obj interface{}, methodName string) (reflect.Type, bool) {
	methodName = toPublicFieldCase(methodName)
	var v reflect.Type
	switch oType := obj.(type) {
	case reflect.Type:
		v = oType
	case reflect.Value:
		v = oType.Type()
	default:
		v = reflect.TypeOf(obj)
	}

	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		if methodType, found := v.MethodByName(methodName); found {
			return methodType.Type, true
		}
		v = v.Elem()
	}
	methodType, found := v.MethodByName(methodName)
	return methodType.Type, found
}

// LookupMethodReturnType does a lookup for a method on an object, and if it's
// found gets it's normalized return type.  If the method doesn't exit or an
// error occurs it returns an error
func LookupMethodReturnType(obj interface{}, methodName string) (reflect.Type, error) {
	methodType, ok := MethodType(obj, methodName)
	if !ok {
		return nil, fmt.Errorf("cannot get method return type for non-existant method %s", methodName)
	}
	return NormalizedReturnType(methodType)
}

// HasMethod returns true if the provided struct supports the defined method
func HasMethod(obj interface{}, methodName string) bool {
	fmt.Println("\ttrying to see if has method is true...")
	_, found := MethodType(obj, methodName)
	return found
}

// MethodReturnType returns a slice of the return types of the method
func MethodReturnType(t reflect.Type) ([]reflect.Type, error) {
	var types []reflect.Type
	if t.Kind() != reflect.Func {
		return types, fmt.Errorf("cannot get return values for non-function type %s", t)
	}
	for idx := 0; idx < t.NumOut(); idx++ {
		types = append(types, t.Out(idx))
	}
	return types, nil
}

// NormalizedReturnType returns a type if the return type is a type or (type,
// error), and an error otherwise.
func NormalizedReturnType(t reflect.Type) (reflect.Type, error) {
	badReturnTypeError := errors.New("return type should be a single value or tuple of (value, error)")
	errType := reflect.TypeOf((*error)(nil)).Elem()
	returns, err := MethodReturnType(t)
	if err != nil {
		return nil, err
	}
	if len(returns) == 1 {
		return returns[0], nil
	}
	if len(returns) == 2 {
		if returns[1].Implements(errType) {
			return returns[0], nil
		}
	}
	return nil, badReturnTypeError
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

	if reflect.Struct == e.Kind() {
		for idx := 0; idx < e.Type().NumField(); idx++ {
			field := e.Type().Field(idx)
			results = append(results, field.Name)
		}
	}
	if reflect.Struct == e.Kind() || reflect.Interface == e.Kind() {
		for idx := 0; idx < e.Type().NumMethod(); idx++ {
			method := e.Type().Method(idx)
			results = append(results, method.Name)
		}
	}
	return results, nil
}

// EvalMember gets a member from a stuct, dereferencing pointers as necessary
func EvalMember(name string, obj interface{}) (reflect.Value, error) {
	name = toPublicFieldCase(name)
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

// MethodValue returns a value for the method name on obj, or an error
func MethodValue(name string, obj interface{}) (reflect.Value, error) {
	name = toPublicFieldCase(name)

	v := reflect.ValueOf(obj)
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return reflect.Zero(reflect.TypeOf(obj)), nilPtrError(v)
		}
		// A method may be attached to a pointer type so check before dereferencing
		if _, ok := v.Type().MethodByName(name); ok {
			return v.MethodByName(name), nil
		}

		if v.IsNil() {
			return reflect.Zero(reflect.TypeOf(obj)), nilPtrError(v)
		}
		v = v.Elem()
	}

	if _, hasMethod := v.Type().MethodByName(name); !hasMethod {
		return reflect.Zero(reflect.TypeOf(obj)), missingFieldError(name, v)
	}

	return v.MethodByName(name), nil
}

// EvalMethod applies params to the provided method and returns a tuple of value
// and error.  The function should return either a single value or a tuple of
// value and an error.  Since we are looking specifically for methods attached
// to obj, obj will be treated as an implicit first argument.
func EvalMethod(name string, obj interface{}, params ...interface{}) (reflect.Value, error) {
	fmt.Printf("calling EvalMethod (%s) on type %T", name, obj)
	nilVal := reflect.Zero(reflect.TypeOf((*error)(nil)))

	if _, ok := obj.(reflect.Type); ok {
		fmt.Println("... it's a type, not trying to do anything")
		return reflect.Zero(reflect.TypeOf(obj)), errors.New("cannot eval method on nil pointer")
	}

	if v, ok := obj.(reflect.Value); ok {
		fmt.Println("... it's a value, checking for nil-ness")
		if v.IsNil() {
			fmt.Println("it's nil, stopping")
			return reflect.Zero(reflect.TypeOf(obj)), nilPtrError(v)
		}
	}

	fmt.Println("checking for nil-ness...")
	if obj == nil {
		fmt.Println("it's nil")
		return reflect.Zero(reflect.TypeOf(obj)), errors.New("cannot eval method on nil pointer")
	}
	fmt.Println("it's not nil")

	valParams := make([]reflect.Value, len(params))
	for idx, param := range params {
		valParams[idx] = toValue(param)
	}

	fmt.Println("finished converting params")

	method, err := MethodValue(name, obj)

	fmt.Println("got method value")

	if err != nil {
		return nilVal, fmt.Errorf("unable to get method %s on type %T: %s", name, obj, err)
	}

	fmt.Println("checking field count")
	if method.Type().NumIn() != len(params) {
		return nilVal,
			fmt.Errorf(
				"%s has arity %d but received %d params",
				name,
				method.Type().NumIn(),
				len(params),
			)
	}
	fmt.Println("calling...")
	unnormalized := method.Call(valParams)
	fmt.Println("normalizing...")
	return normalizeResults(unnormalized)
}

// HasPath returns true of the set of terms can resolve to a value
func HasPath(obj interface{}, terms ...string) error {
	t := reflect.TypeOf(obj)
	for _, term := range terms {
		if returnType, err := LookupMethodReturnType(t, term); err == nil {
			t = returnType
			continue
		}

		for t.Kind() == reflect.Ptr {
			if returnType, err := LookupMethodReturnType(t, term); err == nil {
				t = returnType
				continue
			}
			t = t.Elem()
		}

		if k := t.Kind(); k == reflect.Interface {
			if returnType, err := LookupMethodReturnType(t, term); err == nil {
				t = returnType
				continue
			}
			return nil
		} else if k != reflect.Struct {
			return errors.New("cannot access non-structure field")
		}

		if returnType, err := LookupMethodReturnType(t, term); err == nil {
			t = returnType
			continue
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
	for idx, term := range terms {
		terms[idx] = toPublicFieldCase(term)
	}

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
		fmt.Println("trying to look up term ", term, "...")
		if HasField(obj, term) {
			val, err := EvalMember(term, obj)
			if err != nil {
				return nil, ErrUnresolvable
			}
			obj = val.Interface()
		} else if HasMethod(obj, term) {
			val, err := EvalMethod(term, obj)
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

// mapToLower converts a string slice to all lower case
func mapToLower(strs []string) []string {
	for idx, str := range strs {
		strs[idx] = strings.ToLower(str)
	}
	return strs
}

// toPublicFieldCase converts the first letter in the string to capital
func toPublicFieldCase(s string) string {
	return strings.ToUpper(string(s[0])) + s[1:]
}

func nilPtrError(v reflect.Value) error {
	typeStr := v.Type().String()
	return fmt.Errorf("cannot dereference nil pointer of type %s", typeStr)
}

func missingFieldError(name string, v reflect.Value) error {
	return fmt.Errorf("%s has no field named %s", v.Type().String(), name)
}

// toValue converts an interface type to a value representing the underlying
// type, if it's already a value it returns it's parameter unmodified, and if
// it's a reflect.Type it returns a zero value.
func toValue(i interface{}) reflect.Value {
	if asVal, ok := i.(reflect.Value); ok {
		return asVal
	}
	if asType, ok := i.(reflect.Type); ok {
		return reflect.Zero(asType)
	}
	return reflect.ValueOf(i)
}

// normalizeResult will convert a slice of return values from
// reflet.Value.Call() into a tuple of a value and an error.
func normalizeResults(vals []reflect.Value) (reflect.Value, error) {
	fmt.Println("calling normalize results...")
	nilVal := reflect.Zero(reflect.TypeOf((*error)(nil)))
	errType := reflect.TypeOf((*error)(nil)).Elem()
	len := len(vals)
	if len == 0 {
		return nilVal, nil
	}
	if len == 1 {
		elem := vals[0]
		if elem.Type().Implements(errType) {
			return nilVal, elem.Interface().(error)
		}
		return elem, nil
	}
	if len == 2 {
		fst := vals[0]
		snd := vals[1]
		if snd.Type().Implements(errType) {
			return fst, snd.Interface().(error)
		}
	}
	last := vals[len-1]
	if last.Type().Implements(errType) {
		var err error
		if !last.IsNil() {
			err = last.Interface().(error)
		}
		return reflect.ValueOf(valSliceToInterfaceSlice(vals[0 : len-1])), err
	}
	return reflect.ValueOf(valSliceToInterfaceSlice(vals)), nil
}

func valSliceToInterfaceSlice(vals []reflect.Value) []interface{} {
	interfaces := make([]interface{}, len(vals))
	for idx, val := range vals {
		interfaces[idx] = val.Interface()
	}
	return interfaces
}
