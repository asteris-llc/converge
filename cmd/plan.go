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
	"os"

	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/load"
	"github.com/asteris-llc/converge/plan"
	"github.com/asteris-llc/converge/render"
	"github.com/spf13/cobra"
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
		GracefulExit(cancel)

		// params
		params, err := getParamsFromFlags(cmd.Flags())
		if err != nil {
			log.Fatalf("[FATAL] could not read params: %s\n", err)
		}

		for _, fname := range args {
			log.Printf("[INFO] planning %s\n", fname)

			loaded, err := load.Load(ctx, fname, false)
			if err != nil {
				log.Fatalf("[FATAL] %s: could not parse file: %s\n", fname, err)
			}

			rendered, err := render.Render(ctx, loaded, params)
			if err != nil {
				log.Fatalf("[FATAL] %s: could not render: %s\n", fname, err)
			}

			merged, err := graph.MergeDuplicates(ctx, rendered, graph.SkipModuleAndParams)
			if err != nil {
				log.Fatalf("[FATAL] %s: could not merge duplicates: %s\n", fname, err)
			}

			results, err := plan.Plan(ctx, merged)
			if err != nil && err != plan.ErrTreeContainsErrors {
				log.Fatalf("[FATAL] %s: planning failed: %s\n", fname, err)
			}

			// print results
			out, perr := getPrinter().Show(ctx, results)
			if perr != nil {
				log.Fatalf("[FATAL] %s: failed printing results: %s\n", fname, perr)
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
	planCmd.Flags().Bool("show-meta", false, "show metadata (params and modules)")
	planCmd.Flags().Bool("only-show-changes", false, "only show changes")
	addParamsArguments(planCmd.PersistentFlags())
	RootCmd.AddCommand(planCmd)
}
