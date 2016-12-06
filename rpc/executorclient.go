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
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// NewExecutorClient returns a client for a server that implements Executor
func NewExecutorClient(ctx context.Context, addr string, security *Security) (pb.ExecutorClient, error) {
	opts, err := security.Client()
	if err != nil {
		return nil, errors.Wrap(err, "could not get client options")
	}

	cc, err := grpc.DialContext(ctx, addr, opts...)
	if err != nil {
		return nil, err
	}

	return pb.NewExecutorClient(cc), nil
}
