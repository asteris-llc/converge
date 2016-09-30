package control

import (
	"fmt"
	"reflect"
)

// NewTypeError returns a new TypeError with the appropriate expected and actual
// types.
func NewTypeError(expected string, actual interface{}) error {
	return fmt.Errorf(
		"type error: expected %s but got %s",
		expected, reflect.TypeOf(actual).String(),
	)
}

// BadPredicate returns a new error for an invalid predicate
func BadPredicate(p string) error {
	return fmt.Errorf("invalid predicate: %q", p)
}
