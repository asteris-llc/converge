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
	"fmt"
	"log"

	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/resource"
)

// Plan the execution of a Graph of resource.Tasks
func Plan(in *graph.Graph) (*graph.Graph, error) {
	return in.Transform(func(id string, out *graph.Graph) error {
		task, ok := out.Get(id).(resource.Task)
		if !ok {
			return fmt.Errorf("could not get task, was %T", out.Get(id))
		}

		log.Printf("[DEBUG] checking %q\n", id)

		status, willChange, err := task.Check()
		if err != nil {
			return fmt.Errorf("error checking %s: %s", id, err)
		}

		out.Add(
			id,
			&Result{
				Status:     status,
				WillChange: willChange,
				Task:       task,
			},
		)

		return nil
	})
}
