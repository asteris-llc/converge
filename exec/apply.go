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

package exec

import (
	"fmt"
	"sync"

	"golang.org/x/net/context"

	"github.com/asteris-llc/converge/load"
	"github.com/asteris-llc/converge/resource"
)

// ApplyResult contains the result of a resource check
type ApplyResult struct {
	Path      string
	OldStatus string
	NewStatus string
	Success   bool
}

func (p *ApplyResult) String() string {
	return fmt.Sprintf(
		"%s:\n\tStatus: %q => %q\n\tSuccess: %t",
		p.Path,
		p.OldStatus,
		p.NewStatus,
		p.Success,
	)
}

// Apply the operations to be performed
func Apply(ctx context.Context, graph *load.Graph, plan []*PlanResult) (results []*ApplyResult, err error) {
	// transform checks into something we can look up easily
	planResults := map[string]*PlanResult{}
	for _, result := range plan {
		planResults[result.Path] = result
	}

	// iterate over the checks, looking for things that need changes
	lock := new(sync.Mutex)

	err = graph.Walk(func(path string, res resource.Resource) error {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		planResult, ok := planResults[path]
		if !ok || !planResult.WillChange {
			return nil
		}

		task, ok := res.(resource.Task)
		if !ok {
			return nil
		}

		var (
			err    error
			result = &ApplyResult{
				Path:      path,
				OldStatus: planResult.CurrentStatus,
			}
		)

		result.NewStatus, result.Success, err = task.Apply()
		if err != nil {
			return err
		}

		lock.Lock()
		defer lock.Unlock()
		results = append(results, result)

		return nil
	})

	return results, err
}
