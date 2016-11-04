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
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/arbovm/levenshtein"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

var durationType = reflect.TypeOf(time.Duration(0))

// Preparer wraps and implements resource.Resource in order to deserialize into
// regular Preparers
type Preparer struct {
	Source      map[string]interface{}
	Destination Resource
}

// NewPreparer wraps a given resource in this preparer
func NewPreparer(r Resource) *Preparer {
	return &Preparer{
		Source:      make(map[string]interface{}),
		Destination: r,
	}
}

// NewPreparerWithSource creates a new preparer with the source included
func NewPreparerWithSource(r Resource, source map[string]interface{}) *Preparer {
	prep := NewPreparer(r)
	prep.Source = source

	return prep
}

// Prepare the destination to prepare itself.
func (p *Preparer) Prepare(ctx context.Context, r Renderer) (Task, error) {
	value := reflect.ValueOf(p.Destination)
	typ := value.Type()
	wasPtr := false // so we can re-wrap later if we need to

	if typ.Kind() == reflect.Ptr {
		wasPtr = true
		typ = typ.Elem()
		value = value.Elem()
	}
	if typ.Kind() != reflect.Struct {
		return nil, errors.New("Preparer can only wrap structs")
	}

	if err := p.validateExtra(typ); err != nil {
		return nil, err
	}

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if field.Anonymous {
			continue
		}

		val, err := p.getValueForField(r, field)
		if err != nil {
			return nil, err
		}

		fieldValue := value.Field(i)
		if fieldValue.CanSet() {
			fieldValue.Set(val)
		}
	}

	if wasPtr && value.CanAddr() {
		value = value.Addr()
	}

	unwrapped := value.Interface()
	resource, ok := unwrapped.(Resource)
	if !ok {
		return nil, errors.New("unwrapped was not a Resource")
	}

	return resource.Prepare(ctx, r)
}

func (p *Preparer) validateExtra(typ reflect.Type) error {
	if typ.Kind() != reflect.Struct {
		return errors.New("can't validate extra on a non-struct type")
	}

	fieldNames := map[string]struct{}{}
	for i := 0; i < typ.NumField(); i++ {
		fieldNames[p.getFieldName(typ.Field(i))] = struct{}{}
	}

	// add special fields
	fieldNames["depends"] = struct{}{}
	fieldNames["group"] = struct{}{}

	var err error
	for key := range p.Source {
		if _, ok := fieldNames[key]; ok {
			continue
		}

		// check for spelling errors. Deploy the Levenshtein distance algorithm!
		var candidates []string
		for candidate := range fieldNames {
			if levenshtein.Distance(key, candidate) <= 5 {
				candidates = append(candidates, candidate)
			}
		}

		var msg string
		if len(candidates) > 0 {
			msg = " Maybe you meant: " + strings.Join(candidates, ", ")
		}

		err = multierror.Append(
			err,
			fmt.Errorf("I don't have a field named %q.%s", key, msg),
		)
	}

	return err
}

// getValueForField retrieves and converts the value for a given field
func (p *Preparer) getValueForField(r Renderer, field reflect.StructField) (reflect.Value, error) {
	// get the field name for use in future lookups
	name := p.getFieldName(field)
	raw, isSet := p.Source[name]

	// validate that the param is present, if required
	if err := p.validateRequired(field, raw); err != nil {
		return reflect.Zero(field.Type), err
	}

	// return a default type if nothing is set. No need to do any conversions or
	// anything in this case, we're simply returning the zero value of the field.
	if !isSet {
		return reflect.Zero(field.Type), nil
	}

	// now that we know the field is present, we can make sure it's not
	// violating any mutual exclusion constraints
	if err := p.validateMutuallyExclusive(field); err != nil {
		return reflect.Zero(field.Type), err
	}

	// get the base for numeric conversion, if present
	base, err := p.getBase(field)
	if err != nil {
		return reflect.Zero(field.Type), err
	}

	// finally after all those checks we can deserialize the value of the field
	// from the interface{}!
	value, err := p.convertValue(field.Type, r, name, raw, base)
	if err != nil {
		return value, err
	}

	// validate results
	if err := p.validateValidValues(field, r, base, value); err != nil {
		return reflect.Zero(field.Type), err
	}

	return value, nil
}

// getFieldName extracts a field name from either the "hcl" tag or the field
// name itself.
func (p *Preparer) getFieldName(field reflect.StructField) string {
	if raw, ok := field.Tag.Lookup("hcl"); ok {
		return strings.SplitN(raw, ",", 1)[0]
	}

	return field.Name
}

// getBase returns the base for the conversion of strings to numbers, defaulting
// to base 10.
func (p *Preparer) getBase(field reflect.StructField) (int, error) {
	if raw, ok := field.Tag.Lookup("base"); ok {
		base, err := strconv.Atoi(raw)
		if err != nil {
			return 0, errors.Wrap(err, "could not convert base tag to int")
		}

		return base, nil
	}

	return 10, nil
}

// validateRequired detects if the value is required but not provided
func (p *Preparer) validateRequired(field reflect.StructField, val interface{}) error {
	if required, ok := field.Tag.Lookup("required"); ok && required == "true" && val == nil {
		return fmt.Errorf("%q is required", p.getFieldName(field))
	}

	return nil
}

// validateMutuallyExclusive detects if multiple mutually exclusive fields are
// set
func (p *Preparer) validateMutuallyExclusive(field reflect.StructField) error {
	if mutuallyexclusives, ok := field.Tag.Lookup("mutually_exclusive"); ok {
		name := p.getFieldName(field)

		exclusives := strings.Split(mutuallyexclusives, ",")
		for _, mutuallyexclusive := range exclusives {
			if mutuallyexclusive == name {
				continue
			}

			if _, ok := p.Source[mutuallyexclusive]; ok {
				err := "only one of "
				if len(exclusives) == 2 {
					err += `"` + exclusives[0] + `" or "` + exclusives[1] + `"`
				} else {
					for i, exclusive := range exclusives {
						err += `"` + exclusive + `"`
						if i+1 != len(exclusives) {
							err += ", "
						}
						if i+1 == len(exclusives)-1 {
							err += "or "
						}
					}
				}
				err += " can be set"
				return errors.New(err)
			}
		}
	}

	return nil
}

// validateValidValues detects if the provided value is within an acceptable set
// of values.
func (p *Preparer) validateValidValues(field reflect.StructField, r Renderer, base int, value reflect.Value) error {
	if valids, ok := field.Tag.Lookup("valid_values"); ok {
		name := p.getFieldName(field)

		for _, valid := range strings.Split(valids, ",") {
			parsed, err := p.convertValue(field.Type, r, name, valid, base)
			if err != nil {
				return errors.Wrapf(err, "invalid value for %s: %s", field.Type.Kind(), valid)
			}

			if parsed.Interface() == value.Interface() {
				return nil
			}
		}

		return fmt.Errorf("value did not pass validation. Must be one of %q, was %q", valids, value)
	}

	return nil
}

// convertValue converts and returns the value of an individual element
func (p *Preparer) convertValue(typ reflect.Type, r Renderer, name string, val interface{}, base int) (out reflect.Value, err error) {

	switch typ {
	case durationType:
		out, err = p.convertDuration(typ, r, name, val, base)
	default:
		switch typ.Kind() {
		case reflect.Bool:
			out, err = p.convertBool(r, name, val)

		case reflect.String:
			out, err = p.convertString(r, name, val)

		case reflect.Interface:
			out, err = p.convertInterface(typ, r, name, val, base)

		case reflect.Map:
			out, err = p.convertMap(typ, r, name, val, base)

		case reflect.Slice:
			out, err = p.convertSlice(typ, r, name, val, base)

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
			out, err = p.convertNumber(typ, r, name, val, base)

		case reflect.Ptr:
			out, err = p.convertPointer(typ, r, name, val, base)

		default:
			logrus.WithFields(logrus.Fields{
				"field": name,
				"type":  typ.Kind(),
			}).Warn("could not render field type, using zero value")

			out = reflect.Zero(typ)
		}
	}

	if err != nil {
		return out, err
	}

	return p.realias(out, typ)
}

// realias restores type information lost when converting. Since we convert
// based on the kind of the type, that information gets lost in the case of
// alias types (e.g. `type State string`.) Fortunately, we can just add this
// type information back in by converting, so that's what we do.
func (p *Preparer) realias(val reflect.Value, typ reflect.Type) (reflect.Value, error) {
	if val.Type() != typ {
		if !val.Type().ConvertibleTo(typ) {
			return val, fmt.Errorf("cannot re-alias %s to %s", val.Type(), typ)
		}

		return val.Convert(typ), nil
	}

	return val, nil
}

// convertDuration converts a time.Duration
func (p *Preparer) convertDuration(typ reflect.Type, r Renderer, name string, val interface{}, base int) (reflect.Value, error) {
	if val == nil {
		return reflect.Zero(typ), nil
	}

	switch reflect.ValueOf(val).Kind() {
	case reflect.Int:
		num, err := p.convertNumber(typ, r, name, val, base)
		if err != nil {
			return reflect.Zero(typ), errors.Wrapf(err, "could not convert %v to duration", val)
		}
		dur := time.Duration(num.Int())
		return reflect.ValueOf(dur), nil

	case reflect.String:
		dur, err := time.ParseDuration(val.(string))
		if err != nil {
			return reflect.Zero(typ), errors.Wrapf(err, "could not convert %s to duration", val)
		}
		return reflect.ValueOf(dur), nil

	default:
		return reflect.Zero(typ), fmt.Errorf("cannot handle duration conversion of %v", reflect.ValueOf(val).Kind())
	}
}

// convertBool converts a value to bool using the following rules:
//
// - bool values are used without conversion
// - string values for truth are any capitalization of "t" or "true"
// - any other string value is false
func (p *Preparer) convertBool(r Renderer, name string, val interface{}) (reflect.Value, error) {
	if val == nil {
		return reflect.ValueOf(false), nil
	}

	switch t := val.(type) {
	case bool:
		return reflect.ValueOf(val), nil

	case string:
		boolish, err := r.Render(name, t)
		if err != nil {
			return reflect.ValueOf(false), errors.Wrapf(err, "error rendering field %s", name)
		}

		switch strings.ToLower(boolish) {
		case "t", "true":
			return reflect.ValueOf(true), nil

		default:
			return reflect.ValueOf(false), nil
		}

	default:
		return reflect.ValueOf(false), fmt.Errorf("don't know how to convert %T to bool", t)
	}
}

// convertNumber converts interfaces to numbers
func (p *Preparer) convertNumber(typ reflect.Type, r Renderer, name string, val interface{}, base int) (reflect.Value, error) {
	if val == nil {
		return reflect.Zero(typ), nil
	}

	var (
		num string
		err error
	)
	switch t := val.(type) {
	case string:
		num, err = r.Render(name, t)
		if err != nil {
			return reflect.Zero(typ), errors.Wrapf(err, "error rendering field %s", name)
		}

	default:
		// we're taking an odd approach here. If we have a number already we're
		// converting it to a string because JSON (and thus HCL) numbers are all
		// floating point. Therefore there's no guarantee that we've parsed
		// using the correct semantics (signed vs unsigned vs float) or bitsize.
		num = fmt.Sprintf("%v", val)
	}

	// parse the number back out, depending on the type
	switch typ.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		raw, err := strconv.ParseInt(num, base, typ.Bits())

		if err != nil {
			return reflect.Zero(typ), errors.Wrapf(err, "could not convert %s to %s", num, typ.Kind())
		}

		switch typ.Kind() {
		case reflect.Int:
			return reflect.ValueOf(int(raw)), nil

		case reflect.Int8:
			return reflect.ValueOf(int8(raw)), nil

		case reflect.Int16:
			return reflect.ValueOf(int16(raw)), nil

		case reflect.Int32:
			return reflect.ValueOf(int32(raw)), nil

		case reflect.Int64:
			return reflect.ValueOf(raw), nil
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		raw, err := strconv.ParseUint(num, base, typ.Bits())

		if err != nil {
			return reflect.Zero(typ), errors.Wrapf(err, "could not convert %s to %s", num, typ.Kind())
		}

		switch typ.Kind() {
		case reflect.Uint:
			return reflect.ValueOf(uint(raw)), nil

		case reflect.Uint8:
			return reflect.ValueOf(uint8(raw)), nil

		case reflect.Uint16:
			return reflect.ValueOf(uint16(raw)), nil

		case reflect.Uint32:
			return reflect.ValueOf(uint32(raw)), nil

		case reflect.Uint64:
			return reflect.ValueOf(raw), nil
		}

	case reflect.Float32, reflect.Float64:
		raw, err := strconv.ParseFloat(num, typ.Bits())

		if err != nil {
			return reflect.Zero(typ), errors.Wrapf(err, "could not convert %s to %s", num, typ.Kind())
		}

		switch typ.Kind() {
		case reflect.Float32:
			return reflect.ValueOf(float32(raw)), nil

		case reflect.Float64:
			return reflect.ValueOf(raw), nil
		}
	}

	return reflect.Zero(typ), fmt.Errorf("can't parse a number from %s", typ.Kind())
}

// convertString converts and renders any strings given it
func (p *Preparer) convertString(r Renderer, name string, val interface{}) (reflect.Value, error) {
	if val == nil {
		return reflect.ValueOf(""), nil
	}

	strVal, ok := val.(string)
	if !ok {
		return reflect.ValueOf(""), fmt.Errorf("value was not a string: %v", val)
	}

	rendered, err := r.Render(name, strVal)
	if err != nil {
		return reflect.ValueOf(""), errors.Wrapf(err, "error rendering field %s", name)
	}

	return reflect.ValueOf(rendered), nil
}

// convertInterface reflects back to convertValue in the case of non-zero
// values. There's really not much to see here.
func (p *Preparer) convertInterface(typ reflect.Type, r Renderer, name string, val interface{}, base int) (reflect.Value, error) {
	if val == nil {
		return reflect.Zero(typ), nil
	}

	raw, err := p.convertValue(reflect.TypeOf(val), r, name, val, base)
	if err != nil {
		return raw, err
	}

	return p.maybeUnwrapMap(raw), nil
}

// convertMap properly converts and renders both keys and values
func (p *Preparer) convertMap(typ reflect.Type, r Renderer, name string, val interface{}, base int) (reflect.Value, error) {
	if val == nil {
		return reflect.Zero(typ), nil
	}

	values := p.maybeUnwrapMap(reflect.ValueOf(val))

	if values.Kind() != reflect.Map {
		return reflect.Zero(typ), fmt.Errorf("expected map for %q, got %T", name, val)
	}

	acc := reflect.MakeMap(typ)
	for i, key := range values.MapKeys() {
		// key
		k, err := p.convertValue(
			typ.Key(),
			r,
			fmt.Sprintf("%s.%d.key", name, i),
			key.Interface(),
			base,
		)
		if err != nil {
			return reflect.Zero(typ), errors.Wrapf(err, "could not render %s.%d.key", name, i)
		}

		// value
		v, err := p.convertValue(
			typ.Elem(),
			r,
			fmt.Sprintf("%s.%d.value", name, i),
			values.MapIndex(key).Interface(),
			base,
		)
		if err != nil {
			return reflect.Zero(typ), errors.Wrapf(err, "could not render %s.%d.value", name, i)
		}

		acc.SetMapIndex(k, v)
	}

	return acc, nil
}

func (p *Preparer) maybeUnwrapMap(val reflect.Value) reflect.Value {
	typ := val.Type()
	// HCL does this annoying thing where it deserializes into lists by default.
	// So our val might be a list with one map at index 0. Hooray!
	if typ.Kind() == reflect.Slice && val.Len() == 1 && typ.Elem().Kind() == reflect.Map {
		val = val.Index(0)
	}

	return val
}

// convertSlice properly converts and renders all elements in a slice.
func (p *Preparer) convertSlice(typ reflect.Type, r Renderer, name string, val interface{}, base int) (reflect.Value, error) {
	if val == nil {
		return reflect.Zero(typ), nil
	}

	values := reflect.ValueOf(val)
	if values.Kind() != reflect.Slice {
		return reflect.Zero(typ), fmt.Errorf("expected slice for %q, got %T", name, val)
	}

	acc := reflect.MakeSlice(typ, values.Len(), values.Cap())
	for i := 0; i < values.Len(); i++ {
		item, err := p.convertValue(
			typ.Elem(),
			r,
			fmt.Sprintf("%s.%d", name, i),
			values.Index(i).Interface(),
			base,
		)
		if err != nil {
			return reflect.Zero(typ), errors.Wrapf(err, "could not render %s.%d", name, i)
		}

		acc.Index(i).Set(item)
	}

	return acc, nil
}

// convertPointer wraps whatever value we have in a pointer
func (p *Preparer) convertPointer(typ reflect.Type, r Renderer, name string, val interface{}, base int) (reflect.Value, error) {
	inner, err := p.convertValue(typ.Elem(), r, name, val, base)
	if err != nil {
		return inner, err
	}

	out := reflect.New(typ.Elem())
	out.Elem().Set(inner)

	return out, nil
}
