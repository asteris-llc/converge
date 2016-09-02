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

package either

import (
	"errors"
	"fmt"

	"github.com/asteris-llc/converge/executor/monad"
)

// LeftType represents the left-hand side of an Either
type LeftType struct{ Val interface{} }

// RightType represents the right-hand side of an Either
type RightType struct{ Val interface{} }

// AndThenFunc is a type alias for a function that takes a param and returns Either
type AndThenFunc func(interface{}) Either

// Either represents the Either monad. LeftType.AndThen returns identity,
// RightType.AndThen returns the result of applying the function to the element.
type Either interface {
	AndThen(AndThenFunc) Either
}

// EitherM provides a full monadic wrapper around Either with Return so that it
// implements monad.Monad. An either may be a LeftType or a RightType.  By
// convention LeftType represents some terminal or error condition, and
// RightType represents some intermediate or terminal phase of computation.  It
// is common to use Either in a way similar ot multi-return functions,
// e.g. Either error Type.
type EitherM struct {
	Either
}

// AndThen wraps the underlying Either's AndThen
func (m EitherM) AndThen(f func(interface{}) monad.Monad) monad.Monad {
	switch t := m.Either.(type) {
	case LeftType:
		return m
	case RightType:
		return f(t.Val)
	}
	return LeftM(errors.New("invalid either type"))
}

// Show shows the string
func (m EitherM) String() string {
	val, isRight := m.FromEither()
	pfx := "Left"
	if isRight {
		pfx = "Right"
	}
	return fmt.Sprintf("%s (%v)", pfx, val)
}

// Return creates a Right value
func (m EitherM) Return(i interface{}) monad.Monad {
	return EitherM{Either: Right(i)}
}

// FromLeft gets the left-hand value
func (m EitherM) FromLeft() interface{} {
	return m.Either.(LeftType).Val
}

// FromRight gets the right-hand value
func (m EitherM) FromRight() interface{} {
	return m.Either.(RightType).Val
}

// IsRight returns true if the value is Right
func (m EitherM) IsRight() bool {
	_, ok := m.Either.(RightType)
	return ok
}

// IsLeft returns true if the value is Left
func (m EitherM) IsLeft() bool {
	_, ok := m.Either.(LeftType)
	return ok
}

// FromEither returns the value and true if it was a right-hand value
func (m EitherM) FromEither() (interface{}, bool) {
	if m.IsRight() {
		return m.FromRight(), true
	}
	return m.FromLeft(), false
}

// ReturnM returns and EitherM
func ReturnM(i interface{}) EitherM {
	var e EitherM
	return e.Return(i).(EitherM)
}

// LeftM generates a left instance of EitherM
func LeftM(i interface{}) EitherM {
	return EitherM{Either: Left(i)}
}

// RightM generates a left instance of EitherM
func RightM(i interface{}) EitherM {
	return EitherM{Either: Right(i)}
}

// AndThen on LeftType returns the type without function application
func (l LeftType) AndThen(AndThenFunc) Either {
	return l
}

// AndThen on the RightType returns the result of function application
func (r RightType) AndThen(f AndThenFunc) Either {
	return f(r.Val)
}

// Left creates a new left type value
func Left(i interface{}) Either {
	return LeftType{i}
}

// Right creates a new right type value
func Right(i interface{}) Either {
	return RightType{i}
}

// FromLeft attempts to coerce the either into a left type value
func FromLeft(e Either) (interface{}, bool) {
	if left, ok := e.(LeftType); ok {
		return left.Val, true
	}
	return nil, false
}

// FromRight attempts to coerce the either into a right type value
func FromRight(e Either) (interface{}, bool) {
	if right, ok := e.(RightType); ok {
		return right.Val, true
	}
	return nil, false
}
