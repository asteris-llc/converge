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
	"errors"
	"os"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/asteris-llc/converge/rpc"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc/grpclog"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:     "server",
	Short:   "serve content over HTTP(S)",
	Aliases: []string{"serve"},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// check HTTPS
		if err := validateSSL(); err != nil {
			return err
		}

		// check module serving
		stat, err := os.Stat(viper.GetString("root"))
		if err != nil {
			return err
		}
		if !stat.IsDir() {
			return errors.New("root should be a directory")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithCancel(context.Background())
		GracefulExit(cancel)

		var running sync.WaitGroup
		running.Add(2)

		setLocal(true) // so we generate a token

		// set up our client and server security options
		maybeSetToken()

		if !usingSSL() {
			log.Warning("no SSL config in use, server will accept HTTP connections")
		}

		sslConfig, err := getSSLConfig(getServerName())
		if err != nil {
			log.WithError(err).Fatal("could not get SSL config")
		}

		clientOpts := &rpc.ClientOpts{
			Token: viper.GetString("auth-token"),
			SSL:   sslConfig,
		}

		// start RPC server
		go func() {
			defer running.Done()

			err := startRPC(
				ctx,
				getRPCAddr(),
				sslConfig,
				viper.GetString("root"),
				viper.GetBool("self-serve"),
			)
			if err != nil {
				log.WithError(err).Fatal("could not run RPC")
			}

			<-ctx.Done()
		}()

		// sleep here to avoid a race condition. The REST gateway can't connect
		// to the RPC server if the gateway starts first.
		time.Sleep(100 * time.Millisecond)

		// start HTTP server
		go func() {
			defer running.Done()

			httpLog := log.WithFields(log.Fields{
				"addr":    viper.GetString("api-addr"),
				"service": "API",
			})

			server, err := rpc.NewRESTGateway(ctx, getRPCAddr(), clientOpts)
			if err != nil {
				httpLog.WithError(err).Fatal("failed to create server")
			}

			if viper.GetBool("https") {
				httpLog.WithField("protocol", "HTTPS").Info("serving")
				err = server.ListenAndServeTLS(
					viper.GetString("api-addr"),
					getCertFileLoc(),
					getKeyFileLoc(),
				)
			} else {
				httpLog.WithField("protocol", "HTTP").Info("serving")
				err = server.ListenAndServe(
					viper.GetString("api-addr"),
				)
			}

			if err != nil {
				httpLog.WithError(err).Fatal("failed to serve")
			}

			httpLog.Info("halted")
		}()

		running.Wait()
	},
}

func init() {
	RootCmd.AddCommand(serverCmd)

	// common
	registerSSLFlags(serverCmd.Flags())
	registerRPCFlags(serverCmd.Flags())

	// API
	serverCmd.Flags().String("api-addr", addrServerHTTP, "address to serve API")
	serverCmd.Flags().String("root", ".", "location of modules to serve")
	serverCmd.Flags().Bool("self-serve", false, "serve own binary for bootstrapping")

	// set RPC logging to use logrus
	grpclog.SetLogger(log.WithField("component", "grpc"))
}
