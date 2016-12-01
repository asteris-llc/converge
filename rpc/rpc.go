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
	"crypto/tls"

	"github.com/asteris-llc/converge/rpc/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// New registers all servers and handlers for the RPC server
func New(token string, secure *tls.Config, resourceRoot string, enableBinaryDownload bool) (*grpc.Server, error) {
	var opts []grpc.ServerOption

	if token != "" {
		jwt := NewJWTAuth(token)
		opts = append(opts, grpc.StreamInterceptor(jwt.StreamInterceptor))
		opts = append(opts, grpc.UnaryInterceptor(jwt.UnaryInterceptor))
	}

	if secure != nil {
		opts = append(opts, grpc.Creds(credentials.NewTLS(secure)))
	}

	server := grpc.NewServer(opts...)

	pb.RegisterExecutorServer(server, &executor{})
	pb.RegisterGrapherServer(server, &grapher{})
	pb.RegisterResourceHostServer(
		server,
		&resourceHost{
			root:                 resourceRoot,
			enableBinaryDownload: enableBinaryDownload,
		},
	)

	return server, nil
}
