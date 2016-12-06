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
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
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
		defer cancel()
		GracefulExit(cancel)

		setLocal(true)  // so we generate a token
		maybeSetToken() // set the token, if it's not set
		setLocal(false) // unset local so we get the right flag addresses

		// start RPC server
		if err := startRPC(ctx); err != nil {
			log.WithError(err).Fatal("serving failed")
		}
	},
}

func init() {
	RootCmd.AddCommand(serverCmd)

	// common
	registerSSLFlags(serverCmd.Flags())
	registerRPCFlags(serverCmd.Flags())

	// API
	serverCmd.Flags().String("root", ".", "location of modules to serve")
	serverCmd.Flags().Bool("self-serve", false, "serve own binary for bootstrapping")

	// set RPC logging to use logrus
	grpclog.SetLogger(log.WithField("component", "grpc"))
}
