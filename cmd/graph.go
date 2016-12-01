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
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/asteris-llc/converge/prettyprinters"
	"github.com/asteris-llc/converge/prettyprinters/graphviz"
	"github.com/asteris-llc/converge/prettyprinters/graphviz/providers"
	"github.com/asteris-llc/converge/rpc"
	"github.com/asteris-llc/converge/rpc/pb"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
)

// graphCmd represents the graph command
var graphCmd = &cobra.Command{
	Use:   "graph",
	Short: "graph the execution of a module",
	Long: `graphing is a convenient way to visualize how your graph and
dependencies are structured.

You can pipe the output directly to the 'dot' command, for example:

		converge graph myFile.hcl | dot -Tpng -o myFile.png`,

	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("Need one module filename as argument, got %d", len(args))
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		fname := args[0]

		// set up execution context
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		GracefulExit(cancel)

		// logging
		flog := log.WithField("file", fname).WithField("component", "client")

		maybeSetToken()

		ssl, err := getSSLConfig(getServerURL().Host)
		if err != nil {
			flog.WithError(err).Fatal("could not get SSL config")
		}

		if err := maybeStartSelfHostedRPC(ctx); err != nil {
			flog.WithError(err).Fatal("could not start RPC")
		}

		client, err := getRPCGrapherClient(
			ctx,
			&rpc.ClientOpts{
				Token: getToken(),
				SSL:   ssl,
			},
		)
		if err != nil {
			flog.WithError(err).Fatal("could not get client")
		}

		// load the graph
		graph, err := client.Graph(
			ctx,
			&pb.LoadRequest{
				Location:   fname,
				Parameters: getParamsRPC(cmd),
			},
		)
		if err != nil {
			flog.WithError(err).Fatal("could not get graph")
		}

		printer := prettyprinters.New(
			graphviz.New(
				graphviz.DefaultOptions(),
				providers.RPCProvider{
					ShowParams: viper.GetBool("show-params"),
				},
			),
		)

		dotCode, err := printer.Show(ctx, graph)
		if err != nil {
			flog.WithError(err).Fatal("could not generate dot output")
		}

		fmt.Println(dotCode)
	},
}

func init() {
	graphCmd.Flags().Bool("show-params", false, "also graph param dependencies")
	registerParamsFlags(graphCmd.Flags())
	registerSSLFlags(graphCmd.Flags())
	registerRPCFlags(graphCmd.Flags())
	registerLocalRPCFlags(graphCmd.Flags())

	RootCmd.AddCommand(graphCmd)
}
