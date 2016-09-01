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
	"github.com/asteris-llc/converge/executor/either"
	"github.com/asteris-llc/converge/executor/list"
	"github.com/asteris-llc/converge/graph"
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
func (p Pipeline) AndThen(f func(interface{}) either.EitherM) Pipeline {
	p.CallStack = list.Cons(f, p.CallStack)
	return p
}

// PipelineCall represents a single element in the pipeline
type PipelineCall struct {
	Description    string
	Transformation func(interface{}, *graph.Graph) either.EitherM
}

// ResultList is a slice of result values
type ResultList struct {
	Results []Result
}

// Result defines the result of single call in a call stack
type Result struct {
	Node interface{}
}

// Execute performs a left-fold over the call stack accumulating results in the
// results structure.  Intermediate values are available for successful calls.
func (p *Pipeline) Execute() *ResultList {
	result := &ResultList{}
	return result
}
