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
	"context"

	"github.com/asteris-llc/converge/graph"
)

// Status represents an executor node that can provide status
type Status interface {
	Error() error
}

// Execute executes a pipeline on each node
func Execute(ctx context.Context, in *graph.Graph, pipeline Pipeline) (*graph.Graph, error) {
	out, err := in.Transform(ctx, func(id string, out *graph.Graph) error {
		return nil
	})
	return out, err
}
