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

package load

import (
	"context"

	"github.com/asteris-llc/converge/graph"
	"github.com/pkg/errors"
)

// Load produces a fully-formed graph from the given root
func Load(ctx context.Context, root string, verify bool) (*graph.Graph, error) {
	base, err := Nodes(ctx, root, verify)
	if err != nil {
		return nil, errors.Wrap(err, "loading failed")
	}
	resolved, err := ResolveDependencies(ctx, base)
	if err != nil {
		return nil, errors.Wrap(err, "could not resolve dependencies")
	}
	resourced, err := SetResources(ctx, resolved)
	if err != nil {
		return nil, errors.Wrap(err, "could not resolve resources")
	}
	predicated, err := ResolveConditionals(ctx, resourced)
	return predicated, nil
}
