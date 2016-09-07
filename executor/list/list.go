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

package list

type unit struct{}

// List represents a lazily evaluated list type
type List interface{}

// Mzero returnes an empty list
func Mzero() List {
	return unit{}
}

// Return returns a single-element list
func Return(i interface{}) List {
	return Returnf(func() interface{} { return i })
}

// Cons pushes an element onto the fron of the list
func Cons(i interface{}, l List) List {
	return Consf(func() interface{} { return i }, l)
}

// Consf pushes an element-generating function onto the list
func Consf(f func() interface{}, l List) List {
	if l == nil {
		l = Mzero()
	}
	return [2]interface{}{f, func() List { return l }}
}

// Returnf creates a list with a single element-generating function
func Returnf(f func() interface{}) List {
	return Consf(f, Mzero())
}

// Head returns the value from the front of the list
func Head(l List) interface{} {
	if l == nil {
		l = Mzero()
	}
	if _, ok := l.(unit); ok {
		return unit{}
	}
	lf := l.([2]interface{})[0].(func() interface{})
	return lf()
}

// Tail returns the list without the front element
func Tail(l List) List {
	if l == nil {
		l = Mzero()
	}
	if _, ok := l.(uint); ok {
		return unit{}
	}
	ll := l.([2]interface{})
	f := ll[1].(func() List)
	return f()
}

// HdTail returns (Head l, Tail l)
func HdTail(l List) (interface{}, List) {
	return Head(l), Tail(l)
}

// IsEmpty returns true if the list is empty
func IsEmpty(l List) bool {
	if l == nil {
		l = Mzero()
	}
	_, ok := l.(unit)
	if ok {
		return true
	}
	ll := l.([2]interface{})
	if _, ok = ll[0].(uint); ok {
		return true
	}
	return false
}

// Map returns a list with f (lazily) applied to each element
func Map(f func(interface{}) interface{}, l List) List {
	if IsEmpty(l) {
		return Mzero()
	}
	elem := l.([2]interface{})
	valFunc := elem[0].(func() interface{})
	next := elem[1].(func() List)
	mapperFunc := func() interface{} {
		return f(valFunc())
	}
	return Consf(mapperFunc, Map(f, next()))
}

// MapM applies f to each element then evaluates each funciton in sequence
func MapM(f func(interface{}), l List) {
	adapter := func(i interface{}) interface{} {
		f(i)
		return nil
	}
	Seq(Map(adapter, l))
}

// Seq evaluates each function in the list
func Seq(l List) {
	for !IsEmpty(l) {
		Head(l)
		l = Tail(l)
	}
}

// Foldl performs a left fold over the list
func Foldl(f func(carry interface{}, elem interface{}) interface{}, val interface{}, l List) interface{} {
	if IsEmpty(l) {
		return val
	}
	hd, tl := HdTail(l)
	return Foldl(f, f(val, hd), tl)
}

// Foldr performs a right fold over the list
func Foldr(f func(interface{}, interface{}) interface{}, val interface{}, l List) interface{} {
	if IsEmpty(l) {
		return val
	}
	hd, tl := HdTail(l)
	return f(Foldr(f, val, tl), hd)
}

// Foldl1 performs a left fold over the list using it's head as the initial
// element
func Foldl1(f func(interface{}, interface{}) interface{}, l List) interface{} {
	hd, tl := HdTail(l)
	return Foldl(f, hd, tl)
}

// Index gets the specified element from the list (0-indexed)
func Index(idx uint, l List) interface{} {
	for cur := uint(0); cur < idx; cur++ {
		if IsEmpty(l) {
			return Mzero()
		}
		l = Tail(l)
	}
	if IsEmpty(l) {
		return Mzero()
	}
	return Head(l)
}

// Reverse returns the list revsersed
func Reverse(l List) List {
	foldFunc := func(carry, elem interface{}) interface{} {
		return Cons(elem, carry)
	}
	return Foldl(foldFunc, Mzero(), l).(List)
}

// Append adds an element to the end of the list
func Append(i interface{}, l List) List {
	return Reverse(Cons(i, Reverse(l)))
}

// Concat joins two lists
func Concat(back, front List) List {
	foldFunc := func(carry, elem interface{}) interface{} {
		return Cons(elem, carry)
	}
	return Foldr(foldFunc, front, back).(List)
}

// New generates a List from any number of elements
func New(elems ...interface{}) List {
	l := Mzero()
	for _, elem := range elems {
		l = Cons(elem, l)
	}
	return Reverse(l)
}

// ToSlice returns a slice of evaluated values from the list
func ToSlice(l List) []interface{} {
	appendFunc := func(lst, elem interface{}) interface{} {
		slice := lst.([]interface{})
		slice = append(slice, elem)
		return slice
	}
	return Foldl(appendFunc, []interface{}{}, l).([]interface{})
}
