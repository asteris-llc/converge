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
	"crypto/tls"

	"github.com/asteris-llc/converge/rpc/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// ClientOpts contains options for the Converge RPC client
type ClientOpts struct {
	Token string
	SSL   *tls.Config
}

// Opts transforms the current config into options for grpc.DialContext
func (c *ClientOpts) Opts() (out []grpc.DialOption) {
	if c.SSL == nil {
		out = append(out, grpc.WithInsecure())
	} else {
		out = append(out, grpc.WithTransportCredentials(credentials.NewTLS(c.SSL)))
	}

	if c.Token != "" {
		out = append(out, grpc.WithPerRPCCredentials(NewJWTAuth(c.Token)))
	}

	return out
}

// NewExecutorClient returns a client for a server that implements Executor
func NewExecutorClient(ctx context.Context, addr string, opts *ClientOpts) (pb.ExecutorClient, error) {
	cc, err := grpc.DialContext(ctx, addr, opts.Opts()...)
	if err != nil {
		return nil, err
	}

	return pb.NewExecutorClient(cc), nil
}
