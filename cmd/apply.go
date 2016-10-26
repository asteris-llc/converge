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
	"errors"
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/graph/node"
	"github.com/asteris-llc/converge/helpers/logging"
	"github.com/asteris-llc/converge/rpc"
	"github.com/asteris-llc/converge/rpc/pb"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
)

// applyCmd represents the plan command
var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "apply what needs to change in the system",
	Long: `application is where the actual work of making your execution graph
real happens.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("Need at least one module filename as argument, got 0")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		// set up execution context
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		GracefulExit(cancel)

		// logging
		clog := log.WithField("component", "client")
		ctx = logging.WithLogger(ctx, clog)

		maybeSetToken()

		ssl, err := getSSLConfig(getServerName())
		if err != nil {
			clog.WithError(err).Fatal("could not get SSL config")
		}

		if err = maybeStartSelfHostedRPC(ctx, ssl); err != nil {
			clog.WithError(err).Fatal("could not start RPC")
		}

		client, err := getRPCExecutorClient(
			ctx,
			&rpc.ClientOpts{
				Token: getToken(),
				SSL:   ssl,
			},
		)
		if err != nil {
			clog.WithError(err).Fatal("could not get client")
		}

		rpcParams := getParamsRPC(cmd)

		verifyModules := viper.GetBool("verify-modules")
		if !verifyModules {
			clog.Warn("skipping module verification")
		}

		// execute files
		for _, fname := range args {
			flog := clog.WithField("file", fname)

			flog.Debug("applying")

			stream, err := client.Apply(
				ctx,
				&pb.LoadRequest{
					Location:   fname,
					Parameters: rpcParams,
					Verify:     verifyModules,
				},
			)
			if err != nil {
				flog.WithError(err).Fatal("error getting RPC stream")
			}

			g := graph.New()

			// get edges
			edges, err := getMeta(stream)
			if err != nil {
				flog.WithError(err).Fatal("error getting RPC metadata")
			}
			for _, edge := range edges {
				g.Connect(edge.Source, edge.Dest)
			}

			// get vertices
			err = iterateOverStream(
				stream,
				func(resp *pb.StatusResponse) {
					slog := flog.WithFields(log.Fields{
						"stage": resp.Stage,
						"run":   resp.Run,
						"id":    resp.Meta.Id,
					})
					if resp.Run == pb.StatusResponse_STARTED {
						slog.Info("got status")
					} else {
						slog.Debug("got status")
					}

					if resp.Stage == pb.StatusResponse_APPLY && resp.Run == pb.StatusResponse_FINISHED {
						details := resp.GetDetails()
						if details != nil {
							g.Add(node.New(resp.Id, details.ToPrintable()))
						}
					}
				},
			)
			if err != nil {
				flog.WithError(err).Fatal("could not get responses")
			}

			// validate resulting graph
			if err = g.Validate(); err != nil {
				flog.WithError(err).Warning("graph is not valid")
			}

			// print results
			out, err := getPrinter().Show(ctx, g)
			if err != nil {
				flog.WithError(err).Fatal("failed to print results")
			}

			fmt.Print("\n")
			fmt.Print(out)
		}
	},
}

func init() {
	applyCmd.Flags().Bool("show-meta", false, "show metadata (params and modules)")
	applyCmd.Flags().Bool("only-show-changes", false, "only show changes")
	applyCmd.Flags().Bool("verify-modules", false, "verify module signatures")
	registerRPCFlags(applyCmd.Flags())
	registerLocalRPCFlags(applyCmd.Flags())
	registerSSLFlags(applyCmd.Flags())
	registerParamsFlags(applyCmd.Flags())

	RootCmd.AddCommand(applyCmd)
}
