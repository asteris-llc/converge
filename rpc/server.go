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
)

// Server represents the configuration for a Converge RPC server
type Server struct {
	// Security
	Security *Security

	// Serving
	ResourceRoot         string
	EnableBinaryDownload bool
}

// newGRPC constructs all GRPC servers and handlers
func (s *Server) newGRPC() (*grpc.Server, error) {
	server := grpc.NewServer(s.Security.Server()...)

	pb.RegisterExecutorServer(server, &executor{})
	pb.RegisterGrapherServer(server, &grapher{})
	pb.RegisterResourceHostServer(
		server,
		&resourceHost{
			root:                 s.ResourceRoot,
			enableBinaryDownload: s.EnableBinaryDownload,
		},
	)
	pb.RegisterInfoServer(server, &infoServer{})

	return server, nil
}

// NewREST constructs a new REST gateway
func (s *Server) newREST(ctx context.Context, addr *url.URL) (*http.Server, error) {
	opts, err := s.Security.Client()
	if err != nil {
		return nil, errors.Wrap(err, "could not generate REST gateway security options")
	}
	mux := runtime.NewServeMux(
		runtime.WithMarshalerOption("text/plain", newContentMarshaler()),
	)

	if err := pb.RegisterExecutorHandlerFromEndpoint(ctx, mux, addr.Host, opts); err != nil {
		return nil, errors.Wrap(err, "could not register executor")
	}

	if err := pb.RegisterResourceHostHandlerFromEndpoint(ctx, mux, addr.Host, opts); err != nil {
		return nil, errors.Wrap(err, "could not register resource host")
	}

	if err := pb.RegisterGrapherHandlerFromEndpoint(ctx, mux, addr.Host, opts); err != nil {
		return nil, errors.Wrap(err, "could not register grapher")
	}

	if err := pb.RegisterInfoHandlerFromEndpoint(ctx, mux, addr.Host, opts); err != nil {
		return nil, errors.Wrap(err, "could not register info server")
	}

	handler := http.Handler(mux)

	if s.Security.Token != "" {
		handler = NewJWTAuth(s.Security.Token).Protect(handler)
	}

	return &http.Server{
		Handler: handler,
	}, nil
}

// Listen on the given address for all server-related duties
func (s *Server) Listen(ctx context.Context, addr *url.URL) error {
	logger := logging.GetLogger(ctx).WithField("addr", addr)

	// set up a context within the waitgroup
	wg, ctx := errgroup.WithContext(ctx)

	// set up listeners
	//
	// We'll start with a regular net.Listener. This is going to be our entry
	// point into the whole system. If we're using TLS/SSL, we'll wrap the
	// original listener in a tls listener implementing the same interface.
	// s.Security takes care of this.
	//
	// We need to care about it here because wrapping the listener here means
	// that we can terminate SSL at a single point.
	//
	// One caveat: the REST interface is actually an automatically-generated
	// client of the GRPC interface. This means that we have to require both
	// server and client configuration to use the server. On the other hand, it
	// means that *all* communication is secured when any of it is. Anything
	// that talks to either server component will be encrypted over the wire.
	lis, err := net.Listen("tcp", addr.Host)
	if err != nil {
		return errors.Wrap(err, "failed to listen")
	}
	wg.Go(func() error {
		logger.Debug("waiting to close listener")
		<-ctx.Done()
		logger.Info("closing listener")

		return lis.Close()
	})

	if s.Security.UseSSL {
		logger.Debug("wrapping insecure listener in secure listener")
		lis, err = s.Security.WrapListener(lis)
		if err != nil {
			return errors.Wrap(err, "could not initialize secure listener")
		}
	}

	mux := cmux.New(lis)

	// start the GRPC listener and server
	//
	// Each of the tasks in the workers must handle if the errors they received
	// are any form of use-after-close error. This happens on shutdown for
	// cleanup purposes. In most of these cases, receiving an error means we're
	// already cleaned up so we just need to check which error it is.
	wg.Go(func() error {
		grpcSrv, err := s.newGRPC()
		if err != nil {
			return errors.Wrap(err, "failed to create grpc server")
		}
		lis := mux.Match(cmux.HTTP2HeaderField("content-type", "application/grpc"))

		logger.Info("serving GRPC")
		err = grpcSrv.Serve(lis)
		logger.Debug("finished serving GRPC")

		if err != nil && err != cmux.ErrListenerClosed {
			return errors.Wrap(err, "failed to serve GRPC")
		}
		return nil
	})

	// start the REST gateway listener and server
	//
	// Same cancellation semantics as GRPC listeners.
	wg.Go(func() error {
		restSrv, err := s.newREST(ctx, addr)
		if err != nil {
			return errors.Wrap(err, "failed to create REST server")
		}

		logger.Info("serving REST")
		err = restSrv.Serve(mux.Match(cmux.HTTP1()))
		logger.Debug("finished serving REST")

		if err != nil && err != cmux.ErrListenerClosed {
			return errors.Wrap(err, "failed to serve REST")
		}
		return nil
	})

	// start our cmux listener
	//
	// This is the "master start" switch. If our mux isn't serving, no traffic
	// will flow to either GRPC or the REST gateway.
	wg.Go(func() error {
		logger.Debug("multiplexing")
		err := mux.Serve()
		if err != nil && !IsClosedNetworkConnErr(err) {
			return errors.Wrap(err, "failed to multiplex")
		}
		return nil
	})

	// wait for all listeners to return. Reminder: the semantics of errgroup
	// mean that the *first* error that returns will be returned here. As of
	// the time of this comment, the server will immediately log a fatal line
	// and exit if it receives an error here, so all the error handling possible
	// should be done in this method.
	return wg.Wait()
}

// IsClosedNetworkConnErr detects if an error is the use of a close network connection
func IsClosedNetworkConnErr(err error) bool {
	opErr, ok := err.(*net.OpError)
	// TODO: this feels brittle, but it seems to be the only way since net
	// doesn't export a similar method.
	return ok && opErr.Err.Error() == "use of closed network connection"
}
