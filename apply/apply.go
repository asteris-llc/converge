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
	"context"
	"fmt"

	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/helpers/logging"
	"github.com/asteris-llc/converge/plan"
	"github.com/asteris-llc/converge/resource"
	"github.com/pkg/errors"
)

// ErrTreeContainsErrors is a signal value to indicate errors in the graph
var ErrTreeContainsErrors = errors.New("apply had errors, check graph")

// Apply the actions in a Graph of resource.Tasks
func Apply(ctx context.Context, in *graph.Graph) (*graph.Graph, error) {
	return WithNotify(ctx, in, nil)
}

// WithNotify is Apply, but with notification functions
func WithNotify(ctx context.Context, in *graph.Graph, notify *graph.Notifier) (*graph.Graph, error) {
	var hasErrors error

	logger := logging.GetLogger(ctx).WithField("function", "Apply")

	out, err := in.Transform(
		ctx,
		notify.Transform(func(id string, out *graph.Graph) error {
			val := out.Get(id)
			result, ok := val.(*plan.Result)
			if !ok {
				return fmt.Errorf("%s: could not get *plan.Result, was %T", id, val)
			}

			for _, depID := range graph.Targets(out.DownEdges(id)) {
				dep, ok := out.Get(depID).(*Result)
				if !ok {
					return fmt.Errorf("graph walked out of order: %q before dependency %q", id, depID)
				}

				if err := dep.Error(); err != nil {
					out.Add(
						id,
						&Result{
							Ran:    false,
							Status: &resource.Status{},
							Plan:   result,
							Err:    fmt.Errorf("error in dependency %q", depID),
						},
					)
					// early return here after we set the signal error
					hasErrors = ErrTreeContainsErrors
					return nil
				}
			}

			var newResult *Result

			if result.Status.HasChanges() {
				logger.WithField("id", id).Debug("applying")

				err := result.Task.Apply()
				if err != nil {
					err = errors.Wrapf(err, "error applying %s", id)
				}

				var status resource.TaskStatus

				if err == nil {
					status, err = result.Task.Check()
					if err != nil {
						err = errors.Wrapf(err, "error checking %s", id)
					} else if status.HasChanges() {
						err = fmt.Errorf("%s still needs to be changed after application", id)
					}
				}

				if err != nil {
					hasErrors = ErrTreeContainsErrors
				}

				newResult = &Result{
					Ran:    true,
					Status: status,
					Plan:   result,
					Err:    err,
				}
			} else {
				newResult = &Result{
					Ran:    false,
					Status: result.Status,
					Plan:   result,
					Err:    nil,
				}
			}

			out.Add(id, newResult)

			return nil
		}),
	)

	if err != nil {
		return out, err
	}

	return out, hasErrors
}
