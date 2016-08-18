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
	"fmt"
	"log"

	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/rpc"
	"github.com/asteris-llc/converge/rpc/pb"
	"github.com/spf13/cobra"
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

		maybeSetToken()

		ssl, err := getSSLConfig(getServerName())
		if err != nil {
			log.Fatalf("[FATAL] could not get SSL config: %v", err)
		}

		if err := maybeStartSelfHostedRPC(ctx, ssl); err != nil {
			log.Fatalf("[FATAL] %s\n", err)
		}

		client, err := getRPCExecutorClient(
			ctx,
			&rpc.ClientOpts{
				Token: getToken(),
				SSL:   ssl,
			},
		)
		if err != nil {
			log.Fatalf("[FATAL] %s\n", err)
		}

		rpcParams := getParamsRPC(cmd)

		// execute files
		for _, fname := range args {
			log.Printf("[INFO] applying %s\n", fname)

			stream, err := client.Apply(
				ctx,
				&pb.ExecRequest{
					Location:   fname,
					Parameters: rpcParams,
				},
			)
			if err != nil {
				log.Fatalf("[FATAL] %s: error getting RPC stream: %s\n", fname, err)
			}

			g := graph.New()

			// get edges
			edges, err := getMeta(stream)
			if err != nil {
				log.Fatalf("[FATAL] %s: %s\n", fname, err)
			}
			for _, edge := range edges {
				g.Connect(edge.Source, edge.Dest)
			}

			// get vertices
			err = iterateOverStream(
				stream,
				func(resp *pb.StatusResponse) {
					log.Printf("[INFO] %s: %s %s %s\n", fname, resp.Stage, resp.Id, resp.Run)

					if resp.Stage == pb.StatusResponse_APPLY && resp.Run == pb.StatusResponse_FINISHED {
						details := resp.GetDetails()
						if details != nil {
							g.Add(resp.Id, details.ToPrintable())
						}
					}
				},
			)
			if err != nil {
				log.Fatalf("[FATAL] %s: %s\n", fname, err)
			}

			// validate resulting graph
			if err := g.Validate(); err != nil {
				log.Printf("[WARNING] %s: graph is not valid: %s\n", fname, err)
			}

			// print results
			out, err := getPrinter().Show(ctx, g)
			if err != nil {
				log.Fatalf("[FATAL] %s: failed printing results: %s\n", fname, err)
			}

			fmt.Print("\n")
			fmt.Print(out)
		}
	},
}

func init() {
	applyCmd.Flags().Bool("show-meta", false, "show metadata (params and modules)")
	applyCmd.Flags().Bool("only-show-changes", false, "only show changes")
	registerRPCFlags(applyCmd.Flags())
	registerLocalRPCFlags(applyCmd.Flags())
	registerSSLFlags(applyCmd.Flags())
	registerParamsFlags(applyCmd.Flags())

	RootCmd.AddCommand(applyCmd)
}
