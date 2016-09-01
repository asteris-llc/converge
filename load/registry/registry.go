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

package registry

import (
	"errors"
	"reflect"
)

// Registry for importable types
type Registry struct {
	forward map[string]reflect.Type
	reverse map[reflect.Type]string
}

// New creates a new Registry
func New() *Registry {
	return &Registry{
		map[string]reflect.Type{},
		map[reflect.Type]string{},
	}
}

// Register a new type by import name
func (r *Registry) Register(name string, i interface{}, reverse ...interface{}) error {
	if _, present := r.forward[name]; present {
		return errors.New("name already registered")
	}

	r.forward[name] = reflect.TypeOf(i)

	var err error
	for _, rev := range append(reverse, i) {
		if err = r.RegisterReverse(rev, name); err != nil {
			break
		}
	}
	return err
}

// RegisterReverse registers a name in reverse
func (r *Registry) RegisterReverse(i interface{}, name string) error {
	t := reflect.TypeOf(i)

	if _, present := r.reverse[t]; present {
		return errors.New("type already registered")
	}

	r.reverse[t] = name
	return nil
}

// NewByName creates a new value by the name it was registered under. If no
// type was registered at the given name, the second value will be false
func (r *Registry) NewByName(name string) (interface{}, bool) {
	t, present := r.forward[name]
	if !present {
		return nil, false
	}

	val := reflect.New(t)
	if val.CanInterface() {
		return reflect.Indirect(val).Interface(), true
	}

	return nil, false
}

// NameForType retrieves the name registered for a type. If no name was
// registered for the given type, the second value will be false
func (r *Registry) NameForType(i interface{}) (string, bool) {
	name, present := r.reverse[reflect.TypeOf(i)]
	return name, present
}

// package-global API
var registry *Registry

// Register a type in the global registry
func Register(name string, i interface{}, reverse ...interface{}) {
	if err := registry.Register(name, i, reverse...); err != nil {
		panic(err)
	}
}

// RegisterReverse registers a name in the global registry
func RegisterReverse(i interface{}, name string) {
	if err := registry.RegisterReverse(i, name); err != nil {
		panic(err)
	}
}

// NewByName creates a new value by the name it was registered under. If no
// type was registered at the given name, the second value will be false
func NewByName(name string) (interface{}, bool) {
	return registry.NewByName(name)
}

// NameForType retrieves the name registered for a type. If no name was
// registered for the given type, the second value will be false
func NameForType(i interface{}) (string, bool) {
	return registry.NameForType(i)
}

func init() {
	registry = New()
}
