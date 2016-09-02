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

package pb

import (
	"context"

	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/helpers/logging"
	"github.com/asteris-llc/converge/load"
	"github.com/asteris-llc/converge/render"
	"github.com/pkg/errors"
)

// Load gets a graph from a LocationRequest
func (lr *LoadRequest) Load(ctx context.Context) (*graph.Graph, error) {
	logger := logging.GetLogger(ctx).WithField("location", lr.Location)

	loaded, err := load.Load(ctx, lr.Location)
	if err != nil {
		logger.WithError(err).Error("could not load")
		return nil, errors.Wrapf(err, "loading %s", lr.Location)
	}

	values := render.Values{}
	for k, v := range lr.Parameters {
		values[k] = v
	}
	rendered, err := render.Render(ctx, loaded, values)
	if err != nil {
		logger.WithError(err).Error("could not render")
		return nil, errors.Wrapf(err, "rendering %s", lr.Location)
	}

	merged, err := graph.MergeDuplicates(ctx, rendered, graph.SkipModuleAndParams)
	if err != nil {
		logger.WithError(err).Error("could not merge")
		return nil, errors.Wrapf(err, "merging %s", lr.Location)
	}

	return merged, nil
}
