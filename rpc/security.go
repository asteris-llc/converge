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
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net"

	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Security configuration for
type Security struct {
	Token string

	UseSSL   bool
	CAFile   string
	CertFile string // server only
	KeyFile  string // server only
}

// Server return a server option with the certificate credentials
func (s *Security) Server() (out []grpc.ServerOption) {
	if s.Token != "" {
		jwt := NewJWTAuth(s.Token)
		out = append(out, grpc.UnaryInterceptor(jwt.UnaryInterceptor))
		out = append(out, grpc.StreamInterceptor(jwt.StreamInterceptor))
	}

	return out
}

// WrapListener wraps a listener in a tls.Listener
func (s *Security) WrapListener(lis net.Listener) (net.Listener, error) {
	if s.CertFile == "" || s.KeyFile == "" {
		return nil, errors.New("need both certificate and key file")
	}

	cert, err := tls.LoadX509KeyPair(s.CertFile, s.KeyFile)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load certificates")
	}

	// TODO: add cipher suites, etc?
	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
		Rand:         rand.Reader,
	}

	return tls.NewListener(lis, config), nil
}

// Client returns a dial option for clients
func (s *Security) Client() (out []grpc.DialOption, err error) {
	if s.Token != "" {
		out = append(out, grpc.WithPerRPCCredentials(NewJWTAuth(s.Token)))
	}

	if s.UseSSL {
		logrus.Debug("setting up SSL")

		config, err := s.TLSConfig()
		if err != nil {
			return nil, err
		}

		out = append(out, grpc.WithTransportCredentials(credentials.NewTLS(config)))
	} else {
		logrus.Debug("not using SSL for client")

		out = append(out, grpc.WithInsecure())
	}

	return out, nil
}

// TLSConfig gets a TLS Config from this Security
func (s *Security) TLSConfig() (*tls.Config, error) {
	config := new(tls.Config)
	if s.CAFile != "" {
		certBytes, err := ioutil.ReadFile(s.CAFile)
		if err != nil {
			return nil, errors.Wrap(err, "could not load CA certificate")
		}
		logrus.WithField("cafile", s.CAFile).Debug("read CA certificate")

		roots := x509.NewCertPool()
		if !roots.AppendCertsFromPEM(certBytes) {
			return nil, errors.New("could not append CA certificate as PEM")
		}
		logrus.WithField("cafile", s.CAFile).Debug("loaded CA certificate as PEM")

		config.RootCAs = roots
	}

	return config, nil
}
