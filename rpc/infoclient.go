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

	"google.golang.org/grpc"

	"github.com/asteris-llc/converge/rpc/pb"
	"github.com/golang/protobuf/ptypes/empty"
)

// NewInfoClient returns a client for a server that implements Info
func NewInfoClient(ctx context.Context, addr string, opts *ClientOpts) (*InfoClient, error) {
	cc, err := grpc.DialContext(ctx, addr, opts.Opts()...)
	if err != nil {
		return nil, err
	}

	return &InfoClient{pb.NewInfoClient(cc)}, nil
}

// InfoClient is a wrapper around a pb.InfoClient
type InfoClient struct {
	client pb.InfoClient
}

// Ping gets a ping response from the server
func (i *InfoClient) Ping(ctx context.Context) error {
	_, err := i.client.Ping(ctx, new(empty.Empty))
	return err
}
