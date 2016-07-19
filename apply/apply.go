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

package apply

import (
	"fmt"
	"log"

	"golang.org/x/net/context"

	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/plan"
)

// Apply the actions in a Graph of resource.Tasks
func Apply(ctx context.Context, in *graph.Graph) (*graph.Graph, error) {
	return in.Transform(func(id string, out *graph.Graph) error {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		val := out.Get(id)
		result, ok := val.(*plan.Result)
		if !ok {
			fmt.Println(val)
			return fmt.Errorf("%s: could not get *plan.Result, was %T", id, val)
		}

		var newResult *Result

		if result.WillChange {
			log.Printf("[DEBUG] applying %q\n", id)

			err := result.Task.Apply()
			if err != nil {
				return fmt.Errorf("error applying %s: %s", id, err)
			}

			status, willChange, err := result.Task.Check()
			if err != nil {
				return fmt.Errorf("error checking %s: %s", id, err)
			} else if willChange {
				return fmt.Errorf("%s still needs to be changed after application. Status: %s", id, status)
			}

			newResult = &Result{
				Ran:    true,
				Status: status,
				Plan:   result,
			}
		} else {
			newResult = &Result{
				Ran:    false,
				Status: result.Status,
				Plan:   result,
			}
		}

		out.Add(id, newResult)

		return nil
	})
}
