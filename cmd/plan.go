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
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/graph/node"
	"github.com/asteris-llc/converge/helpers/logging"
	"github.com/asteris-llc/converge/rpc/pb"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
)

// planCmd represents the plan command
var planCmd = &cobra.Command{
	Use:   "plan",
	Short: "plan what needs to change in the system",
	Long: `planning is the first stage in the execution of your changes, and it
can be done separately to see what needs to be changed before execution.`,
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

		if err := maybeStartSelfHostedRPC(ctx); err != nil {
			clog.WithError(err).Fatal("could not start RPC")
		}

		client, err := getRPCExecutorClient(ctx, getSecurityConfig())
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

			flog.Debug("planning")

			stream, err := client.Plan(
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

			timer := new(TimerDisplay)
			timer.Start()
			oldOut := flog.Logger.Out
			flog.Logger.Out = timer.Bypass()

			// get vertices
			var planError bool
			err = iterateOverStream(
				stream,
				func(resp *pb.StatusResponse) {
					slog := flog.WithFields(log.Fields{
						"stage": resp.Stage,
						"run":   resp.Run,
						"id":    resp.Meta.Id,
					})
					switch resp.Run {
					case pb.StatusResponse_STARTED:
						timer.AddTimer(resp.Meta.Id + ": " + resp.Stage.String())
						slog.Info("got status")

					case pb.StatusResponse_FINISHED:
						timer.RemoveTimer(resp.Meta.Id + ": " + resp.Stage.String())
						slog.Debug("got status")

						details := resp.GetDetails()
						if details != nil {
							printable := details.ToPrintable()
							if printable.Error() != nil {
								planError = true
							}
							g.Add(node.New(resp.Id, printable))
						}

					default:
						slog.Warn("got unexpected status")
					}
				},
			)

			timer.Stop()
			flog.Logger.Out = oldOut

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
			if planError {
				os.Exit(1)
			}
		}
	},
}

func init() {
	planCmd.Flags().Bool("show-meta", false, "show metadata (params and modules)")
	planCmd.Flags().Bool("only-show-changes", false, "only show changes")
	planCmd.Flags().Bool("verify-modules", false, "verify module signatures")
	registerRPCFlags(planCmd.Flags())
	registerLocalRPCFlags(planCmd.Flags())
	registerSSLFlags(planCmd.Flags())
	registerParamsFlags(planCmd.Flags())

	RootCmd.AddCommand(planCmd)
}
