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

package executor

import (
	"fmt"

	"github.com/asteris-llc/converge/executor/either"
	"github.com/asteris-llc/converge/executor/list"
	"github.com/asteris-llc/converge/executor/monad"
)

// Pipeline is a type alias for a lazy list of pipeline functions
type Pipeline struct {
	CallStack list.List
}

// NewPipeline creats a new Pipeline with an empty call stack
func NewPipeline() Pipeline {
	return Pipeline{CallStack: list.Mzero()}
}

// AndThen pushes a function onto the pipeline call stack
func (p Pipeline) AndThen(f func(interface{}) monad.Monad) Pipeline {
	p.CallStack = list.Append(f, p.CallStack)
	return p
}

// LogAndThen pushes a function onto the pipeline call stack
func (p Pipeline) LogAndThen(f func(interface{}) monad.Monad, log func(interface{})) Pipeline {
	logged := func(i interface{}) monad.Monad {
		log(i)
		return f(i)
	}
	p.CallStack = list.Append(logged, p.CallStack)
	return p
}

// Connect adds a pipeline to the end of the current pipeline.
// E.g. {a,b,c}.Connect({d,e,f}) = {a,b,c,d,e.f}
func (p Pipeline) Connect(end Pipeline) Pipeline {
	p.CallStack = list.Concat(p.CallStack, end.CallStack)
	return p
}

// Exec executes the pipeline
func (p Pipeline) Exec(zeroValue interface{}) either.EitherM {
	foldFunc := func(carry, elem interface{}) interface{} {
		f, ok := elem.(func(interface{}) monad.Monad)
		if !ok {
			return either.LeftM(badTypeError("func(interface{}) monad.Monad", elem))
		}
		e, ok := carry.(either.EitherM)
		if !ok {
			return either.LeftM(badTypeError("EitherM", carry))
		}
		return e.AndThen(f)
	}
	return list.Foldl(foldFunc, zeroValue, p.CallStack).(either.EitherM)
}

func badTypeError(expected string, actual interface{}) error {
	return fmt.Errorf("expected type `%s' but actual value is of type %T", expected, actual)
}
