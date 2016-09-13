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
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/healthcheck"
	"github.com/asteris-llc/converge/load"
	"github.com/asteris-llc/converge/plan"
	"github.com/asteris-llc/converge/render"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// healthcheckCmd represents the 'healthcheck' command
var healthcheckCmd = &cobra.Command{
	Use:   "healthcheck",
	Short: "display a system health check",
	Long: `Health checks determine the health status of your system.  Health
checks are similar to 'plan' but will not calculate potential deltas, and will
not display healthy checks.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("Need at least one module filename as argument, got 0")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		// set up execution context
		ctx, cancel := context.WithCancel(context.Background())
		GracefulExit(cancel)

		// params
		params := getParams(cmd)

		verifyModules := viper.GetBool("verify-modules")
		if !verifyModules {
			log.WithField("component", "client").Warn("skipping module verfiction")
		}

		for _, fname := range args {
			flog := log.WithField("file", fname)
			flog.Info("checking health")

			loaded, err := load.Load(ctx, fname, verifyModules)
			if err != nil {
				flog.WithError(err).Fatal("could not parse file")
			}

			rendered, err := render.Render(ctx, loaded, params)
			if err != nil {
				flog.WithError(err).Fatal("could not render")
			}

			merged, err := graph.MergeDuplicates(ctx, rendered, graph.SkipModuleAndParams)
			if err != nil {
				flog.WithError(err).Fatal("could not merge duplicates")
			}

			planned, err := plan.Plan(ctx, merged)
			if err != nil && err != plan.ErrTreeContainsErrors {
				flog.WithError(err).Fatal("planning failed")
			}

			results, err := healthcheck.CheckGraph(ctx, planned)
			if err != nil {
				flog.WithError(err).Fatal("checking failed")
			}

			out, perr := healthPrinter().Show(ctx, results)
			if perr != nil {
				flog.WithError(perr).Fatal("failed printing results")
			}

			fmt.Print("\n")
			fmt.Print(out)
			if err != nil {
				os.Exit(1)
			}
		}
	},
}

func init() {
	healthcheckCmd.Flags().Bool("quiet", false, "show only a short summary of the status")
	healthcheckCmd.Flags().Bool("verify-modules", false, "verify module signatures")
	registerParamsFlags(healthcheckCmd.Flags())

	RootCmd.AddCommand(healthcheckCmd)
}
