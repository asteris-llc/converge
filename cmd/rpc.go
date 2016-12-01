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

package cmd

import (
	"encoding/json"
	"io"
	"net"
	"net/url"
	"strings"
	"time"

	"google.golang.org/grpc/metadata"

	log "github.com/Sirupsen/logrus"
	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/helpers/logging"
	"github.com/asteris-llc/converge/rpc"
	"github.com/asteris-llc/converge/rpc/pb"
	"github.com/fgrid/uuid"
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
)

const (
	rpcNoTokenFlagName = "no-token"
	rpcTokenFlagName   = "rpc-token"
	rpcAddrFlagName    = "rpc-addr"
	rpcLocalAddrName   = "local-addr"
	rpcEnableLocalName = "local"
)

func registerRPCFlags(flags *pflag.FlagSet) {
	flags.String(rpcTokenFlagName, "", "token for RPC")
	flags.Bool(rpcNoTokenFlagName, false, "don't use or generate an RPC token")

	flags.String(rpcAddrFlagName, addrServer, "address for RPC connection")
}

func registerLocalRPCFlags(flags *pflag.FlagSet) {
	flags.String(rpcLocalAddrName, addrServerLocal, "address for local RPC connection")
	flags.Bool(rpcEnableLocalName, false, "self host RPC")
}

func maybeStartSelfHostedRPC(ctx context.Context) error {
	if getLocal() {
		go startRPC(ctx)

		var err error
		for i := 0; i < 5; i++ {
			_, err = net.Dial("tcp", getServerURL().Host)
			if err == nil {
				return nil
			}
			time.Sleep(100 * time.Millisecond)
		}

		return err
	}

	return nil
}

func startRPC(ctx context.Context) error {
	// set context for logging
	logger := logging.GetLogger(ctx).WithField("component", "rpc")
	ctx = logging.WithLogger(ctx, logger)

	loc := getServerURL()

	// set up security options
	sslConfig, err := getSSLConfig(loc.Host)
	if err != nil {
		return errors.Wrap(err, "could not get SSL config")
	}
	if !usingSSL() {
		logger.Warning("no SSL config in use, server will accept HTTP connections")
	}

	// create server
	server := &rpc.Server{
		Token:                getToken(),
		Secure:               sslConfig,
		ResourceRoot:         viper.GetString("root"),
		EnableBinaryDownload: viper.GetBool("self-serve"),
		ClientOpts: &rpc.ClientOpts{
			Token: getToken(),
			SSL:   sslConfig,
		},
	}

	return server.Listen(ctx, loc)
}

func getRPCExecutorClient(ctx context.Context, opts *rpc.ClientOpts) (pb.ExecutorClient, error) {
	return rpc.NewExecutorClient(ctx, getServerURL().Host, opts)
}

func getRPCGrapherClient(ctx context.Context, opts *rpc.ClientOpts) (*rpc.GrapherClient, error) {
	return rpc.NewGrapherClient(ctx, getServerURL().Host, opts)
}

func getInfoClient(ctx context.Context, opts *rpc.ClientOpts) (*rpc.InfoClient, error) {
	return rpc.NewInfoClient(ctx, getServerURL().Host, opts)
}

type recver interface {
	Recv() (*pb.StatusResponse, error)
}

func iterateOverStream(stream recver, cb func(*pb.StatusResponse)) error {
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return errors.Wrap(err, "error getting status response")
		}

		cb(resp)
	}

	return nil
}

type headerer interface {
	Header() (metadata.MD, error)
}

func getMeta(stream headerer) ([]*graph.Edge, error) {
	meta, err := stream.Header()
	if err != nil {
		return nil, errors.Wrap(err, "error getting RPC header")
	}

	var edges []*graph.Edge
	if blobs, ok := meta["edges"]; ok {
		for _, blob := range blobs {
			var out []*graph.Edge
			err := json.Unmarshal([]byte(blob), &out)
			if err != nil {
				return nil, errors.Wrap(err, "could not deserialize edge metadata")
			}

			edges = append(edges, out...)
		}
	}

	return edges, nil
}

// Token

func getToken() string { return viper.GetString(rpcTokenFlagName) }

func maybeSetToken() {
	if viper.GetBool(rpcNoTokenFlagName) {
		log.Warning("no token set, server is unauthenticated. This should *only* be used for development.")
		return
	}

	if getToken() == "" && getLocal() {
		viper.Set(rpcTokenFlagName, uuid.NewV4().String())
		log.WithField("token", getToken()).Warn("setting session-local token")
	}
}

// More getters

func setLocal(local bool)  { viper.Set(rpcEnableLocalName, local) }
func getLocal() bool       { return viper.GetBool(rpcEnableLocalName) }
func getLocalAddr() string { return viper.GetString(rpcLocalAddrName) }
func getRPCAddr() string   { return viper.GetString(rpcAddrFlagName) }

func getServerURL() *url.URL {
	out := new(url.URL)

	if getLocal() {
		out.Host = getLocalAddr()
	} else {
		out.Host = getRPCAddr()
	}

	// set host to localhost, if not set
	if strings.HasPrefix(out.Host, ":") {
		out.Host = "127.0.0.1" + out.Host
	}

	// set protocol
	if usingSSL() {
		out.Scheme = "https"
	} else {
		out.Scheme = "http"
	}

	return out
}
