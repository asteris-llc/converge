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

	"github.com/kardianos/osext"

	"golang.org/x/net/context"
)

// Server is the root of all public functionality
type Server struct {
	mux    *http.ServeMux
	server *ContextServer
}

// New creates and returns a new Server
func New(ctx context.Context, moduleDir string, selfServe bool) *Server {
	server := &Server{
		mux:    http.NewServeMux(),
		server: NewContextServer(ctx),
	}

	server.mux.Handle("/modules/", http.StripPrefix("/modules/", http.FileServer(http.Dir(moduleDir))))

	if selfServe {
		server.mux.HandleFunc("/bootstrap/binary", server.handleSelfServe)
	}

	return server
}

// ListenAndServe serves the current mux
func (s *Server) ListenAndServe(addr string) error {
	return s.server.ListenAndServe(addr, s.mux)
}

// ListenAndServeTLS serves the current mux over TLS
func (s *Server) ListenAndServeTLS(addr, certFile, keyFile string) error {
	return s.server.ListenAndServeTLS(addr, certFile, keyFile, s.mux)
}

// handleSelfServe serves the converge binary for bootstrapping in a cluster
func (s *Server) handleSelfServe(w http.ResponseWriter, r *http.Request) {
	name, err := osext.Executable()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("[ERROR]: %s", err)
		return
	}

	http.ServeFile(w, r, name)
}
