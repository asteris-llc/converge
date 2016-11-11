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
	"reflect"
)

var (
	// ErrNilStruct is returned if there is an attempt to introspect a nil value
	ErrNilStruct = errors.New("cannot inspect types with a nil value")

	// ErrNonStructIntrospection is returned when attempting to introspect fields
	// on a non-struct type.
	ErrNonStructIntrospection = errors.New("cannot inspect non-struct fields")
)

// ExportedField represents an exported field, including the containing struct,
// offset, field name, and lookup name
type ExportedField struct {
	FieldName     string
	ReferenceName string
	StructField   *reflect.StructField
	Value         reflect.Value
}

// ExportedFields returns a slice of fields that have been exported from a
// struct, along with the name.
func ExportedFields(input interface{}) (exported []*ExportedField, err error) {
	if nil == input {
		return exported, ErrNilStruct
	}
	val, err := getStruct(reflect.ValueOf(input))
	if err != nil {
		return exported, err
	}

	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		refName, ok := field.Tag.Lookup("export")
		if !ok {
			continue
		}
		exportedField := &ExportedField{
			FieldName:     field.Name,
			ReferenceName: refName,
			Value:         val.Field(i),
		}
		exported = append(exported, exportedField)
	}
	return exported, nil
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
