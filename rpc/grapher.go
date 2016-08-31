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
	"github.com/asteris-llc/converge/rpc/pb"
	"github.com/pkg/errors"
)

type grapher struct {
	auth *authorizer
}

// Graph returns the information about a graph
func (g *grapher) Graph(in *pb.LoadRequest, stream pb.Grapher_GraphServer) error {
	logger, ctx := setIDLogger(stream.Context())
	logger = logger.WithField("function", "grapher.Graph")

	if err := g.auth.authorize(ctx); err != nil {
		logger.WithError(err).Info("authorization failed")
		return errors.Wrap(err, "authorization failed")
	}

	loaded, err := in.Load(ctx)
	if err != nil {
		logger.WithError(err).Error("loading failed")
		return errors.Wrap(err, "loading failed")
	}

	for _, vertex := range loaded.Vertices() {
		err = stream.Send(
			&pb.GraphComponent{&pb.GraphComponent_Vertex_{&pb.GraphComponent_Vertex{
				Id: vertex,
			}}},
		)
		if err != nil {
			logger.WithError(err).WithField("id", vertex).Error("failed to send vertex")
			return errors.Wrapf(err, "failed to send %s", vertex)
		}
	}

	for _, edge := range loaded.Edges() {
		err = stream.Send(
			&pb.GraphComponent{&pb.GraphComponent_Edge_{&pb.GraphComponent_Edge{
				Source:     edge.Source,
				Dest:       edge.Dest,
				Attributes: edge.Attributes,
			}}},
		)
		if err != nil {
			logger.WithError(err).WithField("edge", edge).Error("failed to send edge")
			return errors.Wrapf(err, "failed to send %s", edge)
		}
	}

	return nil
}
