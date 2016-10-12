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

package healthcheck

import (
	"context"
	"errors"

	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/graph/node"
	"github.com/asteris-llc/converge/resource"
)

// Check defines the interface for a health check
type Check interface {
	FailingDep(string, resource.TaskStatus)
	HealthCheck() (*resource.HealthStatus, error)
}

// CheckGraph walks a graph and runs health checks on each health-checkable node
func CheckGraph(ctx context.Context, in *graph.Graph) (*graph.Graph, error) {
	return WithNotify(ctx, in, nil)
}

// WithNotify is CheckGraph, but with notification features
func WithNotify(ctx context.Context, in *graph.Graph, notify *graph.Notifier) (*graph.Graph, error) {

	return in.Transform(
		ctx,
		notify.Transform(func(id string, out *graph.Graph) error {
			var val interface{}
			if meta, ok := out.Get(id); ok {
				val = meta.Value()
			}

			task, err := unboxNode(val)
			if err != nil {
				return err
			}

			asCheck, ok := task.(Check)
			if !ok {
				return nil
			}

			for _, dep := range out.Dependencies(id) {
				meta, ok := out.Get(dep)
				if !ok {
					continue
				}

				depStatus, ok := meta.Value().(resource.TaskStatus)
				if !ok {
					continue
				}

				if isFailure, failErr := isFailingStatus(depStatus); failErr != nil {
					return failErr
				} else if isFailure {
					asCheck.FailingDep(dep, depStatus)
				}
			}

			status, err := asCheck.HealthCheck()
			if err != nil {
				return err
			}

			out.Add(node.New(id, status))

			return nil
		}),
	)
}

// unboxNode will remove a resource.TaskStatus from a plan.Result or apply.Result
func unboxNode(i interface{}) (resource.TaskStatus, error) {
	type statusWrapper interface {
		GetStatus() resource.TaskStatus
	}
	switch result := i.(type) {
	case statusWrapper:
		return result.GetStatus(), nil
	case resource.TaskStatus:
		return result, nil
	default:
		return nil, errors.New("cannot get task status from node")
	}
}

func isFailingStatus(stat resource.TaskStatus) (bool, error) {
	if check, ok := stat.(Check); ok {
		checkStatus, err := check.HealthCheck()
		if err != nil {
			return true, err
		}
		return checkStatus.ShouldDisplay(), nil
	}
	return stat.HasChanges(), nil
}
