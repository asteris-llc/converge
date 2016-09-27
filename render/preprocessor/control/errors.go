package control

import (
	"fmt"
	"reflect"
)

// TypeError represents a type error implementation of the error interface
type TypeError struct {
	expected string
	actual   reflect.Type
}

// NewTypeError returns a new TypeError with the appropriate expected and actual
// types.
func NewTypeError(expected string, actual interface{}) *TypeError {
	return &TypeError{
		expected: expected,
		actual:   reflect.TypeOf(actual),
	}
}

// Error implements the error interface
func (t *TypeError) Error() string {
	return fmt.Sprintf(
		"type error: expected %s but got %s",
		t.expected, t.actual.String(),
	)
}
