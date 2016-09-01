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
	"github.com/asteris-llc/converge/executor/list"
	"github.com/asteris-llc/converge/resource"
)

// Pipeline is a type alias for a lazy list of pipeline functions
type Pipeline list.List

type DependencyError struct {
	Err error
}

type EvaluatedResult struct {
	Task   resource.Task
	Result resource.TaskStatus
}

type PipelineStage interface{}

func IsLeft(p PipelineStage) bool {
	_, ok := p.(*DependencyError)
	return ok
}

func Left(p PipelineStage) *DependencyError {
	return p.(*DependencyError)
}

func IsRight(p PipelineStage) bool {
	_, ok := p.(*EvaluatedResult)
	return ok
}

func Right(p PipelineStage) *EvaluatedResult {
	return p.(*EvaluatedResult)
}

type PipelineFunc func(interface{}) EvaluatedResult

func bindPipeline(s PipelineStage, f func(EvaluatedResult) PipelineStage) x {
	if IsLeft(s) {
		return s
	}
	return
}
