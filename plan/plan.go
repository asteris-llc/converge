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

package plan

import (
	"errors"
	"fmt"
	"log"

	"golang.org/x/net/context"

	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/resource"
)

// ErrTreeContainsErrors is a signal value to indicate errors in the graph
var ErrTreeContainsErrors = errors.New("plan has errors, check graph")

// Plan the execution of a Graph of resource.Tasks
func Plan(ctx context.Context, in *graph.Graph) (*graph.Graph, error) {
	var hasErrors error

	out, err := in.Transform(ctx, func(id string, out *graph.Graph) error {
		val := out.Get(id)
		task, ok := val.(resource.Task)
		if !ok {
			fmt.Println(val)
			return fmt.Errorf("%s: could not get resource.Task, was %T", id, val)
		}

		log.Printf("[DEBUG] checking dependencies for %q\n", id)
		for _, depID := range out.DownEdges(id) {
			dep, ok := out.Get(depID).(*Result)
			if !ok {
				return fmt.Errorf("graph walked out of order: %q before dependency %q", id, depID)
			}

			if err := dep.Error(); err != nil {
				out.Add(
					id,
					&Result{
						Status:     "<unknown>",
						WillChange: true,
						Task:       task,
						Err:        fmt.Errorf("error in dependency %q", depID),
					},
				)

				// early return here after we set the signal error
				hasErrors = ErrTreeContainsErrors
				return nil
			}
		}

		log.Printf("[DEBUG] checking %q\n", id)

		status, willChange, err := task.Check()
		out.Add(
			id,
			&Result{
				Status:     status,
				WillChange: willChange,
				Task:       task,
				Err:        err,
			},
		)

		return nil
	})

	if err != nil {
		return out, err
	}

	return out, hasErrors
}
