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
	"reflect"

	"github.com/pkg/errors"
)

// FieldMap represents a map of field names to interfaces
type FieldMap map[string]interface{}

/* Rules for field access in structures:

- Fields that are tagged with `export` will be exported.

- Named structs that are tagged with `export` will be exported as a struct

- Embedded structs will have their exported fields exported in the namespace of
	the containing struct

- Embedded interfaces will not be exported, nor have their fields exported

- If an embedded struct field name collides with a field from the struct that
	it's embedded in, both will be exported with the embedded struct being
	accessible with 'StructName.FieldName' */

var (
	// ErrNilStruct is returned if there is an attempt to introspect a nil value
	ErrNilStruct = errors.New("cannot inspect types with a nil value")

	// ErrNonStructIntrospection is returned when attempting to introspect fields
	// on a non-struct type.
	ErrNonStructIntrospection = errors.New("cannot extract fields for types other than struct")

	// ErrFieldIndexOutOfBounds is returned when a field index is beyond the
	// number of fields in the struct type
	ErrFieldIndexOutOfBounds = errors.New("struct field index is out of bounds")

	// ErrDuplicateFieldName is returned if there is a duplicate reference name in
	// an exported field slice during map construction
	ErrDuplicateFieldName = errors.New("detected duplicate field name")
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
		return nil, false
	}
	val, err := getStruct(reflect.ValueOf(input))
	if err != nil {
		return nil, false
	}

	if index >= val.Type().NumField() {
		return nil, false
	}
	fieldType := val.Type().Field(index)
	fieldVal := val.Field(index)
	exportedName, ok := fieldType.Tag.Lookup("export")
	if !ok {
		return nil, false
	}
	return &ExportedField{
		FieldName:     fieldType.Name,
		ReferenceName: exportedName,
		Value:         fieldVal,
	}, true
}

// ExportedFields returns a slice of fields that have been exported from a
// struct; including embedded fields
func ExportedFields(input interface{}) (exported []*ExportedField, err error) {
	nonEmbeddedFields := make(map[string]struct{})
	embeddedFields := make(map[string][]*ExportedField)
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
			thisField := asStruct.Field(i).Interface()
			fromEmbedded, err := ExportedFields(thisField)
			if err != nil {
				return exported, err
			}
			embeddedFields[asStruct.Type().Field(i).Name] = fromEmbedded
			continue
		}

		exportedAs, isReExported := asStruct.Type().Field(i).Tag.Lookup("re-export-as")
		if isReExported {
			isKind, kindErr := fieldIsKind(input, i, reflect.Struct)
			if kindErr != nil {
				return exported, kindErr
			}
			if !isKind {
				continue
			}
			thisField := asStruct.Field(i).Interface()
			fromEmbedded, err := ExportedFields(thisField)
			if err != nil {
				return exported, err
			}
			for _, f := range fromEmbedded {
				f.ReferenceName = exportedAs + "." + f.ReferenceName
				exported = append(exported, f)
			}
			continue
		}
		if field, ok := newExportedField(input, i); ok {
			nonEmbeddedFields[field.ReferenceName] = struct{}{}
			exported = append(exported, field)
		}
	}
	for embeddedStruct, fieldSet := range embeddedFields {
		exported = append(exported, disambiguateFields(nonEmbeddedFields, embeddedStruct, fieldSet)...)
	}
	return exported, nil
}

// disambiguateFields will prefix the struct name to the exported field name for
// any exported field whos name would collide with the exported fields of the
// parent struct
func disambiguateFields(
	structFields map[string]struct{},
	structName string,
	fields []*ExportedField,
) []*ExportedField {
	for _, field := range fields {
		if _, ok := structFields[field.ReferenceName]; ok {
			field.ReferenceName = structName + "." + field.ReferenceName
		}
	}
	return fields
}

// GenerateLookupMap takes an exported field list and generates a map of lookup
// names to values
func GenerateLookupMap(fields []*ExportedField) (FieldMap, error) {
	output := make(FieldMap)
	for _, field := range fields {
		_, ok := output[field.ReferenceName]
		if ok {
			return output, errors.Wrap(ErrDuplicateFieldName, field.ReferenceName)
		}
		output[field.ReferenceName] = field.Value.Interface()
	}
	return output, nil
}

// LookupMapFromStruct generates a lookup map from a struct
func LookupMapFromStruct(input interface{}) (FieldMap, error) {
	exported, err := ExportedFields(input)
	if err != nil {
		return make(FieldMap), err
	}
	return GenerateLookupMap(exported)
}

// LookupMapFromInterface gets the concrete implementation of an interface and
// then gets the struct fields from it
func LookupMapFromInterface(input interface{}) (FieldMap, error) {
	switch reflect.ValueOf(input).Kind() {
	case reflect.Ptr, reflect.Interface:
		return LookupMapFromInterface(reflect.ValueOf(input).Elem().Interface())
	case reflect.Struct:
		return LookupMapFromStruct(input)
	}
	return make(FieldMap), ErrNonStructIntrospection
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
