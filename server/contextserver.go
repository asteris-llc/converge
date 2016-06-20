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

package server

import (
	"log"
	"net/http"

	"github.com/braintree/manners"

	"golang.org/x/net/context"
)

// ContextServer is a Server controlled by a context for stopping gracefully
type ContextServer struct {
	ctx context.Context
}

// NewContextServer constructs and returns a server. Mux has the same meaning as
// in the standard HTTP package (that is, if nil it will use the globally
// registered handlers)
func NewContextServer(ctx context.Context) *ContextServer {
	return &ContextServer{ctx}
}

// ListenAndServe does the same thing as the net/http equivalent, except
// using the context.
func (s *ContextServer) ListenAndServe(addr string, handler http.Handler) error {
	server := manners.NewWithServer(&http.Server{Addr: addr, Handler: handler})
	go s.close(server)

	return server.ListenAndServe()
}

// ListenAndServeTLS does the same thing as the net/http equivalent, except
// using the context.
func (s *ContextServer) ListenAndServeTLS(addr, certFile, keyFile string, handler http.Handler) error {
	server := manners.NewWithServer(&http.Server{Addr: addr, Handler: handler})
	go s.close(server)

	return server.ListenAndServeTLS(certFile, keyFile)
}

func (s *ContextServer) close(server *manners.GracefulServer) {
	<-s.ctx.Done()
	log.Println("[INFO] gracefully stopping server")
	server.BlockingClose()
	log.Println("[INFO] server stopped")
}
