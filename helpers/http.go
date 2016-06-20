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

package helpers

import (
	"errors"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"os"
	"path"
	"time"
)

type Stoppable interface {
	Serve() error
	Stop() error
	URL() error
}

// StoppableHTTP is an HTTP server that can be stopped.
type StoppableHTTP struct {
	Server   *http.Server // http server to listen with
	listener net.Listener // listener to close
}

// Serve starts the HTTP server. It should generally be called in its own
// goroutine.
func (s *StoppableHTTP) Serve() (err error) {
	s.listener, err = net.Listen("tcp", s.Server.Addr)
	if err != nil {
		return
	}
	return s.Server.Serve(s.listener)
}

// Stop closes the underlying listener, effectively killing the server
func (s *StoppableHTTP) Stop() error {
	return s.listener.Close()
}

// URL returns the URL that the server would listen on/is listening on.
func (s *StoppableHTTP) URL() string {
	return s.Server.Addr
}

// SingleFileServer is a Stoppable server that serves a single static file.
type SingleFileServer struct {
	Path      string
	Port      int
	server    *http.Server
	stoppable *StoppableHTTP
}

// Serve satisfies the Stoppable interface
func (sfs *SingleFileServer) Serve() error {
	f, err := os.Open(sfs.Path)
	if err != nil {
		return err
	}
	http.HandleFunc(path.Join("/", path.Base(sfs.Path)), func(w http.ResponseWriter, r *http.Request) {
		http.ServeContent(w, r, path.Base(sfs.Path), time.Now(), f)
	})

	sfs.server = &http.Server{Addr: fmt.Sprintf(":%v", sfs.Port)}
	sfs.stoppable = &StoppableHTTP{Server: sfs.server}
	return sfs.stoppable.Serve()
}

// Stop satisfies the Stoppable interface
func (sfs *SingleFileServer) Stop() error {
	return sfs.stoppable.Stop()
}

// URL satisfies the Stoppable interface
func (sfs *SingleFileServer) URL() string {
	return fmt.Sprintf("http://localhost:%v/%v", sfs.Port, path.Base(sfs.Path))
}

// HTTPServeFile constructs a SingleFileServer on a random port, returning that
// server.
func HTTPServeFile(filePath string) (sfs *SingleFileServer, err error) {
	for tries := 5; tries > 0; tries-- {
		// get a random port in the IANA dynamic/ephemeral range
		port := rand.Intn(65535-49151) + 49151

		// start an HTTP server on that port
		errors := make(chan error)
		sfs := &SingleFileServer{Port: port, Path: filePath}
		go func(stoppable *SingleFileServer, errors chan error) {
			errors <- stoppable.Serve()
		}(sfs, errors)

		// if it hasn't terminated in .1s, assume it's listening
		dur, _ := time.ParseDuration(".1s")
		select {
		case <-errors:
		case <-time.After(dur):
			return sfs, nil
		}
	}
	return sfs, errors.New("Couldn't find port to listen on")
}
