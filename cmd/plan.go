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
	"strings"
	"sync"

	"golang.org/x/net/context"

	"github.com/acmacalister/skittles"
	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/load"
	"github.com/asteris-llc/converge/plan"
	"github.com/asteris-llc/converge/render"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

			results, err := plan.Plan(ctx, merged)
			if err != nil {
				log.Fatalf("[FATAL] %s: planning failed: %s\n", fname, err)
			}

			var (
				cLock  = new(sync.Mutex)
				counts struct {
					results, changes int
				}
			)

			fmt.Print("\n")

			err = results.Walk(ctx, func(id string, val interface{}) error {
				result, ok := val.(*plan.Result)
				if !ok {
					return fmt.Errorf("expected %T at %q, but got %T", result, id, val)
				}

				cLock.Lock()
				defer cLock.Unlock()

				counts.results++
				if result.WillChange {
					counts.changes++
				}

				if !viper.GetBool("show-meta") && (strings.HasPrefix(graph.BaseID(id), "param") || strings.HasPrefix(graph.BaseID(id), "module")) {
					return nil
				}

				if viper.GetBool("only-show-changes") && !result.WillChange {
					return nil
				}

				fmt.Printf(
					"%s:\n\tWill Change: %t\n\tStatus:\n\t\t%s\n\n",
					id,
					result.WillChange,
					strings.Replace(result.Status, "\n", "\n\t\t", -1),
				)
				return nil
			})
			if err != nil {
				log.Fatalf("[FATAL] %s: printing failed: %s\n", fname, err)
			}

			// summarize the potential changes for the user
			summary := fmt.Sprintf("\nPlan complete. %d checks, %d will change\n", counts.results, counts.changes)
			if UseColor() {
				if counts.changes > 0 {
					summary = skittles.Yellow(summary)
				} else {
					summary = skittles.Green(summary)
				}
			}
			fmt.Print(summary)
		}
	},
}

func init() {
	planCmd.Flags().Bool("show-meta", false, "show metadata (params and modules)")
	planCmd.Flags().Bool("only-show-changes", true, "only show changes")
	addParamsArguments(planCmd.PersistentFlags())
	RootCmd.AddCommand(planCmd)
}
