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
	"net"
	"net/http"
	"net/url"

	"golang.org/x/sync/errgroup"

	"github.com/asteris-llc/converge/helpers/logging"
	"github.com/asteris-llc/converge/rpc/pb"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/pkg/errors"
	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Server represents the configuration for a Converge RPC server
type Server struct {
	// Security
	Token  string
	Secure *tls.Config

	// Serving
	ResourceRoot         string
	EnableBinaryDownload bool

	// Client
	ClientOpts *ClientOpts
}

// newGRPC constructs all servers and handlers
func (s *Server) newGRPC() (*grpc.Server, error) {
	var opts []grpc.ServerOption
	if s.Secure != nil {
		opts = append(opts, grpc.Creds(credentials.NewTLS(s.Secure)))
	}

	server := grpc.NewServer(opts...)

	if s.Token != "" {
		jwt := NewJWTAuth(s.Token)
		opts = append(opts, grpc.UnaryInterceptor(jwt.UnaryInterceptor))
		opts = append(opts, grpc.StreamInterceptor(jwt.StreamInterceptor))
	}

	pb.RegisterExecutorServer(server, &executor{})
	pb.RegisterGrapherServer(server, &grapher{})
	pb.RegisterResourceHostServer(
		server,
		&resourceHost{
			root:                 s.ResourceRoot,
			enableBinaryDownload: s.EnableBinaryDownload,
		},
	)

	return server, nil
}

// NewREST constructs a new REST gateway
func (s *Server) newREST(ctx context.Context, addr string) (*http.Server, error) {
	mux := runtime.NewServeMux(
		runtime.WithMarshalerOption("text/plain", newContentMarshaler()),
	)

	if err := pb.RegisterExecutorHandlerFromEndpoint(ctx, mux, addr, s.ClientOpts.Opts()); err != nil {
		return nil, errors.Wrap(err, "could not register executor")
	}

	if err := pb.RegisterResourceHostHandlerFromEndpoint(ctx, mux, addr, s.ClientOpts.Opts()); err != nil {
		return nil, errors.Wrap(err, "could not register resource host")
	}

	if err := pb.RegisterGrapherHandlerFromEndpoint(ctx, mux, addr, s.ClientOpts.Opts()); err != nil {
		return nil, errors.Wrap(err, "could not register grapher")
	}

	handler := http.Handler(mux)

	if s.Token != "" {
		handler = NewJWTAuth(s.Token).Protect(handler)
	}

	return &http.Server{
		Handler: handler,
	}, nil
}

// Listen on the given address for all server-related duties
func (s *Server) Listen(ctx context.Context, addr *url.URL) error {
	logger := logging.GetLogger(ctx).WithField("addr", addr)

	lis, err := net.Listen("tcp", addr.Host)
	if err != nil {
		return errors.Wrap(err, "failed to listen")
	}

	// set up a context for cancelling out of all of this
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	wg, ctx := errgroup.WithContext(ctx)

	mux := cmux.New(lis)

	// start the GRPC listener
	grpcSrv, err := s.newGRPC()
	if err != nil {
		return errors.Wrap(err, "failed to create grpc server")
	}
	grpcLis := mux.Match(cmux.HTTP2HeaderField("content-type", "application/grpc"))
	wg.Go(func() error {
		logger.Info("serving GRPC")
		err := grpcSrv.Serve(grpcLis)
		if err == cmux.ErrListenerClosed {
			return nil
		}
		return err
	})

	// start the REST listener
	restSrv, err := s.newREST(ctx, addr.Host)
	if err != nil {
		return errors.Wrap(err, "failed to create REST server")
	}
	restLis := mux.Match(cmux.HTTP1())
	wg.Go(func() error {
		logger.Debug("waiting to close REST listener")
		<-ctx.Done()
		logger.Info("closing REST listener")
		return restLis.Close()
	})
	wg.Go(func() error {
		logger.Info("serving REST")
		// TODO: https
		err := restSrv.Serve(restLis)
		if err == cmux.ErrListenerClosed {
			return nil
		}
		return err
	})

	// start our cmux listener
	wg.Go(func() error {
		logger.Info("multiplexing")
		err := mux.Serve()
		if opErr, ok := err.(*net.OpError); ok && opErr.Err.Error() == "use of closed network connection" {
			return nil
		}
		return err
	})

	// wait for all listeners to return
	return wg.Wait()
}
