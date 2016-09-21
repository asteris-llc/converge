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
	"context"
	"crypto/tls"
	"encoding/json"
	"io"
	"net"
	"strings"

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

func maybeStartSelfHostedRPC(ctx context.Context, secure *tls.Config) error {
	if viper.GetBool(rpcEnableLocalName) {
		return startRPC(ctx, getLocalAddr(), secure, "", false)
	}

	return nil
}

func startRPC(ctx context.Context, addr string, secure *tls.Config, resourceRoot string, enableBinaryDownload bool) error {
	// set context for logging
	logger := logging.GetLogger(ctx).WithField("component", "rpc")
	ctx = logging.WithLogger(ctx, logger)

	// listen and start server
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return errors.Wrap(err, "could not open RPC listener connection")
	}

	server, err := rpc.New(getToken(), secure, resourceRoot, enableBinaryDownload)
	if err != nil {
		return errors.Wrap(err, "could not create RPC server")
	}

	go func() {
		<-ctx.Done()
		server.GracefulStop()
	}()

	rpcLog := logger.WithField("addr", addr)

	rpcLog.Info("serving")
	go func() {
		if err := server.Serve(lis); err != nil {
			rpcLog.WithError(err).Fatal("failed to serve")
		}

		rpcLog.Info("halted")
	}()

	return nil
}

func getRPCExecutorClient(ctx context.Context, opts *rpc.ClientOpts) (pb.ExecutorClient, error) {
	var addr string
	if viper.GetBool(rpcEnableLocalName) {
		addr = viper.GetString(rpcLocalAddrName)
	} else {
		addr = viper.GetString(rpcAddrFlagName)
	}

	return rpc.NewExecutorClient(ctx, addr, opts)
}

func getRPCGrapherClient(ctx context.Context, opts *rpc.ClientOpts) (*rpc.GrapherClient, error) {
	var addr string
	if viper.GetBool(rpcEnableLocalName) {
		addr = viper.GetString(rpcLocalAddrName)
	} else {
		addr = viper.GetString(rpcAddrFlagName)
	}

	return rpc.NewGrapherClient(ctx, addr, opts)
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

func setLocal(local bool)  { viper.Set(rpcLocalAddrName, local) }
func getLocal() bool       { return viper.GetBool(rpcLocalAddrName) }
func getRPCAddr() string   { return viper.GetString(rpcAddrFlagName) }
func getLocalAddr() string { return viper.GetString(rpcLocalAddrName) }

func getServerName() string {
	var addr string
	if getLocal() {
		addr = getLocalAddr()
	} else {
		addr = getRPCAddr()
	}

	parts := strings.SplitN(addr, ":", 1)
	if len(parts) < 2 || parts[0] == "" {
		return "127.0.0.1"
	}
	return parts[0]
}
