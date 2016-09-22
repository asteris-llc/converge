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
	"github.com/asteris-llc/converge/executor/monad"
)

// MonadicPipelineFunc represents a monadic pipeline function
type MonadicPipelineFunc func(i interface{}) monad.Monad

// PipelineFunc represents a pipelined function that uses multi-return instead
// of either.
type PipelineFunc func(interface{}) (interface{}, error)

// Pipeline is a type alias for a lazy list of pipeline functions
type Pipeline struct {
	CallStack []PipelineFunc
}

// NewPipeline creats a new Pipeline with an empty call stack
func NewPipeline() Pipeline {
	return Pipeline{[]PipelineFunc{}}
}

// AndThen is a utility function that converts a PipelineFunc into a
// MonadicPipelineFunc before adding it to the execution list as part of the
// refactor to remove Either from pipeline processing.
func (p Pipeline) AndThen(f PipelineFunc) Pipeline {
	p.CallStack = append(p.CallStack, f)
	return p
}

// Connect adds a pipeline to the end of the current pipeline.
// E.g. {a,b,c}.Connect({d,e,f}) = {a,b,c,d,e.f}
func (p Pipeline) Connect(end Pipeline) Pipeline {
	p.CallStack = append(p.CallStack, end.CallStack...)
	return p
}

// Exec executes the pipeline
func (p Pipeline) Exec(zeroValue interface{}) (interface{}, error) {
	var err error
	var val = zeroValue
	for _, f := range p.CallStack {
		val, err = f(val)
		if err != nil {
			return nil, err
		}
	}
	return val, nil
}

// MultiReturnToEither adapts a PipelineFunc to a MonadicPipelineFunc.  It's use
// is limited to intermediate refactoring to remove EitherM from the pipeline
// execution code.
func MultiReturnToEither(f PipelineFunc) MonadicPipelineFunc {
	return func(i interface{}) monad.Monad {
		val, err := f(i)
		if err == nil {
			return either.RightM(val)
		}
		return either.LeftM(err)
	}
}

func badTypeError(expected string, actual interface{}) error {
	return fmt.Errorf("expected type `%s' but actual value is of type %T", expected, actual)
}
