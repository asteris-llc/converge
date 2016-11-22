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

package resource

import (
	"errors"
	"fmt"
	"reflect"
)

/* Rules for field access in structures:
- Fields that are tagged with `export` will be exported.
- Named structs that are tagged with `export` will be exported as a struct
- Embedded structs will have their exported fields exported in the namespace of the containing struct
- Embedded interfaces will not be exported, nor have their fields exported
*/

var (
	// ErrNilStruct is returned if there is an attempt to introspect a nil value
	ErrNilStruct = errors.New("cannot inspect types with a nil value")

	// ErrNonStructIntrospection is returned when attempting to introspect fields
	// on a non-struct type.
	ErrNonStructIntrospection = errors.New("cannot inspect non-struct fields")

	// ErrFieldIndexOutOfBounds is returned when a field index is beyond the
	// number of fields in the struct type
	ErrFieldIndexOutOfBounds = errors.New("struct field index is out of bounds")
)

// ExportedField represents an exported field, including the containing struct,
// offset, field name, and lookup name
type ExportedField struct {
	FieldName     string
	ReferenceName string
	StructField   *reflect.StructField
	Value         reflect.Value
}

func newExportedField(input interface{}, index int) (*ExportedField, bool) {
	if nil == input {
		fmt.Println("newExportedField: input is nil")
		return nil, false
	}
	val, err := getStruct(reflect.ValueOf(input))
	if err != nil {
		fmt.Println("newExportedField: cannot resolve input to a struct type: ", err)
		return nil, false
	}

	if index >= val.Type().NumField() {
		fmt.Println("newExportedField: index out of bounds (", index, " > ", val.Type().NumField(), ")")
		return nil, false
	}
	fieldType := val.Type().Field(index)
	fieldVal := val.Field(index)
	exportedName, ok := fieldType.Tag.Lookup("export")
	if !ok {
		fmt.Printf("newExportedField: %T.%s: field is not exported\n", input, fieldType.Name)
		return nil, false
	}
	return &ExportedField{
		FieldName:     fieldType.Name,
		ReferenceName: exportedName,
		Value:         fieldVal,
	}, true
}

// ExportedFields returns a slice of fields that have been exported from a
// struct, along with the name.
func ExportedFields(input interface{}) (exported []*ExportedField, err error) {
	var embeddedExports []*ExportedField
	if nil == input {
		return exported, ErrNilStruct
	}
	asStruct, err := getStruct(reflect.ValueOf(input))
	if err != nil {
		return exported, err
	}
	for i := 0; i < asStruct.Type().NumField(); i++ {
		isAnon, anonErr := fieldIsAnonymous(input, i)
		if anonErr != nil {
			return exported, anonErr
		}
		if isAnon {
			isKind, kindErr := fieldIsKind(input, i, reflect.Struct)
			if kindErr != nil {
				return exported, kindErr
			}
			if !isKind {
				continue
			}
			fromEmbedded, err := ExportedFields(input)
			if err != nil {
				return exported, err
			}
			embeddedExports = append(embeddedExports, fromEmbedded...)
			continue
		}
		fmt.Println("non-anonymous field at idx: ", i)
		if field, ok := newExportedField(input, i); ok {
			fmt.Println("\t generated field, appending to exported list")
			exported = append(exported, field)
		}
	}
	return exported, nil
}

func getFieldKind(input interface{}, index int) (reflect.Kind, error) {
	if input == nil {
		return reflect.Invalid, ErrNilStruct
	}
	asStruct, err := getStruct(reflect.ValueOf(input))
	if err != nil {
		return reflect.Invalid, err
	}
	if index >= asStruct.Type().NumField() {
		return reflect.Invalid, ErrFieldIndexOutOfBounds
	}
	fieldType := asStruct.Type().Field(index).Type
	for fieldType.Kind() == reflect.Ptr {
		fieldType = fieldType.Elem()
	}
	return fieldType.Kind(), nil
}

func fieldIsKind(input interface{}, index int, kinds ...reflect.Kind) (bool, error) {
	for _, kind := range kinds {
		actualKind, err := getFieldKind(input, index)
		if err != nil {
			return false, err
		}
		if kind == actualKind {
			return true, nil
		}
	}
	return false, nil
}

func fieldIsAnonymous(input interface{}, index int) (bool, error) {
	if input == nil {
		return false, ErrNilStruct
	}
	asStruct, err := getStruct(reflect.ValueOf(input))
	if err != nil {
		return false, err
	}
	if index >= asStruct.Type().NumField() {
		return false, ErrFieldIndexOutOfBounds
	}
	return asStruct.Type().Field(index).Anonymous, nil
}

func getStruct(val reflect.Value) (reflect.Value, error) {
	for val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return val, ErrNonStructIntrospection
	}
	return val, nil
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
