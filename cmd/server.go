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
	"log"
	"os"
	"sync"
	"time"

	"github.com/asteris-llc/converge/rpc"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

		// set up our client and server security options
		maybeSetToken()

		if !usingSSL() {
			log.Println("[WARNING] no SSL config in use, server will accept HTTP connections")
		}

		sslConfig, err := getSSLConfig(getServerName())
		if err != nil {
			log.Fatalf("[FATAL] could not get SSL config: %s", err)
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
				log.Fatalf("[FATAL] could not run RPC: %s", err)
			}

			<-ctx.Done()
		}()

		// sleep here to avoid a race condition. The REST gateway can't connect
		// to the RPC server if the gateway starts first.
		time.Sleep(100 * time.Millisecond)

		// start HTTP server
		go func() {
			defer running.Done()

			server, err := rpc.NewRESTGateway(ctx, getRPCAddr(), clientOpts)
			if err != nil {
				log.Fatalf("[FATAL] failed to create server: %v", err)
			}

			if viper.GetBool("https") {
				log.Printf("[INFO] serving HTTPS on %s\n", viper.GetString("api-addr"))
				err = server.ListenAndServeTLS(
					viper.GetString("api-addr"),
					getCertFileLoc(),
					getKeyFileLoc(),
				)
			} else {
				log.Printf("[INFO] serving HTTP on %s\n", viper.GetString("api-addr"))
				err = server.ListenAndServe(
					viper.GetString("api-addr"),
				)
			}

			if err != nil {
				log.Fatalf("[FATAL] %s\n", err)
			}

			log.Println("[INFO] halted HTTP server")
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
}
