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

// AccumulatingPipeline is a list of TaggedPipelines
type AccumulatingPipeline struct {
	Pipeline
}

// TaggedPipeline is a pipeline with tags associated with subphases
type TaggedPipeline struct {
	Tag string
	Pipeline
}

// ResultAccumulator stores a list of results
type ResultAccumulator map[string]interface{}

// NewPipeline creats a new Pipeline with an empty call stack
func NewPipeline() Pipeline {
	return Pipeline{CallStack: list.Mzero()}
}

// AndThen pushes a function onto the pipeline call stack
func (p Pipeline) AndThen(f func(interface{}) monad.Monad) Pipeline {
	p.CallStack = list.Append(f, p.CallStack)
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

// NewAccumulatingPipeline returns a new accumulating pipeline
func NewAccumulatingPipeline() AccumulatingPipeline {
	return AccumulatingPipeline{Pipeline: NewPipeline()}
}

// AndThen adds a new pipeline into the accumulating pipeline with the
// associated tag
func (p AccumulatingPipeline) AndThen(tag string, pipeline Pipeline) AccumulatingPipeline {
	tagged := TaggedPipeline{Tag: tag, Pipeline: pipeline}
	p.CallStack = list.Append(tagged, p.CallStack)
	return p
}

// Exec executes
func (p AccumulatingPipeline) Exec(zeroValue interface{}) (ResultAccumulator, either.EitherM) {
	acc := make(ResultAccumulator)
	foldFunc := func(carry, elem interface{}) interface{} {
		tagged, ok := elem.(TaggedPipeline)
		if !ok {
			return either.LeftM(badTypeError("TaggedPipeline", elem))
		}
		result := tagged.Pipeline.Exec(carry)
		acc[tagged.Tag] = result
		return result
	}
	val := list.Foldl(foldFunc, zeroValue, p.Pipeline.CallStack)
	return acc, val.(either.EitherM)
}

func badTypeError(expected string, actual interface{}) error {
	return fmt.Errorf("expected type `%s' but actual value is of type %T", expected, actual)
}
