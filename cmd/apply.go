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
	"log"
	"os"

	"golang.org/x/net/context"

	"github.com/asteris-llc/converge/apply"
	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/load"
	"github.com/asteris-llc/converge/plan"
	"github.com/asteris-llc/converge/render"
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
		params := getParams(cmd)

		// set up execution context
		ctx, cancel := context.WithCancel(context.Background())
		GracefulExit(cancel)

		// iterate over modules
		for _, fname := range args {
			log.Printf("[INFO] applying %s\n", fname)

			loaded, err := load.Load(ctx, fname)
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

			// prep done! Time to run some commands!
			printer := getPrinter()

			planned, err := plan.Plan(ctx, merged)
			if err != nil {
				if err == plan.ErrTreeContainsErrors {
					out, perr := printer.Show(ctx, planned)
					if perr != nil {
						log.Printf("[ERROR] %s: printing failed plan failed: %s\n", fname, perr)
					} else {
						fmt.Print("\n")
						fmt.Print(out)
					}
					log.Fatalf("[FATAL] %s: planning failed: check output\n", fname)
				}

				log.Fatalf("[FATAL] %s: planning failed: %s\n", fname, err)
			}

			results, err := apply.Apply(ctx, planned)
			if err != nil && err != apply.ErrTreeContainsErrors {
				log.Fatalf("[FATAL] %s: applying failed: %s\n", fname, err)
			}

			// print results
			out, perr := printer.Show(ctx, results)
			if perr != nil {
				log.Fatalf("[FATAL] %s: failed printing results: %s\n", fname, err)
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
	applyCmd.Flags().Bool("show-meta", false, "show metadata (params and modules)")
	applyCmd.Flags().Bool("only-show-changes", false, "only show changes")
	addParamsArguments(applyCmd.PersistentFlags())
	viperBindPFlags(applyCmd.Flags())

	RootCmd.AddCommand(applyCmd)
}
