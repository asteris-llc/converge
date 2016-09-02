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
	"net/http"

	"github.com/asteris-llc/converge/rpc/pb"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
)

func restGatewayMux(ctx context.Context, addr string, opts *ClientOpts) (http.Handler, error) {
	mux := runtime.NewServeMux(
		runtime.WithMarshalerOption("text/plain", newContentMarshaler()),
	)

	err := pb.RegisterExecutorHandlerFromEndpoint(ctx, mux, addr, opts.Opts())
	if err != nil {
		return nil, err
	}

	err = pb.RegisterResourceHostHandlerFromEndpoint(ctx, mux, addr, opts.Opts())
	if err != nil {
		return nil, err
	}

	err = pb.RegisterGrapherHandlerFromEndpoint(ctx, mux, addr, opts.Opts())
	if err != nil {
		return nil, err
	}

	return mux, nil
}

// NewRESTGateway constructs a REST gateway with the given options
func NewRESTGateway(ctx context.Context, addr string, opts *ClientOpts) (*ContextServer, error) {
	mux, err := restGatewayMux(ctx, addr, opts)
	if err != nil {
		return nil, err
	}

	// set up auth
	if opts.Token != "" {
		mux = NewJWTAuth(opts.Token).Protect(mux)
	}

	// create and return a context server
	return NewContextServer(ctx, mux), nil
}
