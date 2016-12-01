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

package rpc

import (
	"encoding/json"
	"fmt"

	"github.com/asteris-llc/converge/load/registry"
	"github.com/asteris-llc/converge/render"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/rpc/pb"
	"github.com/pkg/errors"
)

type grapher struct{}

// Graph returns the information about a graph
func (g *grapher) Graph(in *pb.LoadRequest, stream pb.Grapher_GraphServer) error {
	logger, ctx := setIDLogger(stream.Context())
	logger = logger.WithField("function", "grapher.Graph")

	loaded, err := in.Load(ctx)
	if err != nil {
		logger.WithError(err).Error("loading failed")
		return errors.Wrap(err, "loading failed")
	}

	for _, vertex := range loaded.Vertices() {
		var val interface{}
		if meta, ok := loaded.Get(vertex); ok {
			val = meta.Value()
		}

		node, err := resolveVertex(vertex, val)
		if err != nil {
			return errors.Wrapf(err, "%T is an unknown vertex type", val)
		}

		kind, ok := registry.NameForType(node)
		if !ok {
			kind = "unknown"
		}

		vbytes, err := json.Marshal(node)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("could not marshal vertex for type: %T ", node))
		}

		err = stream.Send(
			pb.NewGraphComponent(&pb.GraphComponent_Vertex{
				Id:      vertex,
				Kind:    kind,
				Details: vbytes,
			}),
		)
		if err != nil {
			logger.WithError(err).WithField("id", vertex).Error("failed to send vertex")
			return errors.Wrapf(err, "failed to send %s", vertex)
		}
	}

	for _, edge := range loaded.Edges() {
		err = stream.Send(
			pb.NewGraphComponent(&pb.GraphComponent_Edge{
				Source:     edge.Source,
				Dest:       edge.Dest,
				Attributes: edge.Attributes,
			}),
		)
		if err != nil {
			logger.WithError(err).WithField("edge", edge).Error("failed to send edge")
			return errors.Wrapf(err, "failed to send %s", edge)
		}
	}

	return nil
}

func resolveVertex(id string, vertex interface{}) (resource.Task, error) {
	switch v := vertex.(type) {
	case *render.PrepareThunk:
		return resource.NewThunkedTask(id, v.Task), nil
	case *resource.TaskWrapper:
		if resolved, ok := resource.ResolveTask(v); ok {
			return resolved, nil
		}
		return nil, errors.New("unable to resolve wrapped task")
	case resource.Task:
		return v, nil
	}
	return nil, fmt.Errorf("%T cannot be resolved into a Task", vertex)
}
