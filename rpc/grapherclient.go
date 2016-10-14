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
	"context"
	"io"

	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/graph/node"
	"github.com/asteris-llc/converge/rpc/pb"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

// NewGrapherClient returns a client for a server that implements Executor
func NewGrapherClient(ctx context.Context, addr string, opts *ClientOpts) (*GrapherClient, error) {
	cc, err := grpc.DialContext(ctx, addr, opts.Opts()...)
	if err != nil {
		return nil, err
	}

	return &GrapherClient{pb.NewGrapherClient(cc)}, nil
}

// GrapherClient is a wrapper around a pb.GrapherClient
type GrapherClient struct {
	client pb.GrapherClient
}

// Graph gets the graph from the remote side
func (gc *GrapherClient) Graph(ctx context.Context, loc *pb.LoadRequest, opts ...grpc.CallOption) (*graph.Graph, error) {
	stream, err := gc.client.Graph(ctx, loc, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "could not open stream")
	}

	g := graph.New()

	for {
		container, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, errors.Wrap(err, "error getting graph component")
		}

		if container.Component == nil {
			continue
		}

		if vertex := container.GetVertex(); vertex != nil {
			g.Add(node.New(vertex.Id, vertex))
		} else if edge := container.GetEdge(); edge != nil {
			var parent bool
			for _, attr := range edge.Attributes {
				if attr == "parent" {
					parent = true
				}
			}

			if parent {
				g.ConnectParent(edge.Source, edge.Dest)
			} else {
				g.Connect(edge.Source, edge.Dest)
			}
		}
	}

	return g, nil
}
