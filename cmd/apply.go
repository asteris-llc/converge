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

			trimmed, err := graph.TrimDuplicates(ctx, rendered, graph.SkipModuleAndParams)
			if err != nil {
				log.Fatalf("[FATAL] %s: could not trim duplicates: %s\n", fname, err)
			}

			planned, err := plan.Plan(ctx, trimmed)
			if err != nil {
				log.Fatalf("[FATAL] %s: planning failed: %s\n", fname, err)
			}

			results, err := apply.Apply(ctx, planned)
			if err != nil {
				log.Fatalf("[FATAL] %s: applying failed: %s\n", fname, err)
			}

			fmt.Print("\n")

			// count successes and failures to print summary
			var (
				cLock  = new(sync.Mutex)
				counts struct {
					results, ran int
				}
			)

			err = results.Walk(ctx, func(id string, val interface{}) error {
				result, ok := val.(*apply.Result)
				if !ok {
					return fmt.Errorf("expected %T at %q, but got %T", result, id, val)
				}

				cLock.Lock()
				defer cLock.Unlock()

				counts.results++
				if result.Ran {
					counts.ran++
				}

				fmt.Printf(
					"%s:\n\tRan: %t\n\tOld Status:\n\t\t%s\n\tNew Status:\n\t\t%s\n\n",
					id,
					result.Ran,
					strings.Replace(result.Plan.Status, "\n", "\n\t\t", -1),
					strings.Replace(result.Status, "\n", "\n\t\t", -1),
				)

				return nil
			})
			if err != nil {
				log.Fatalf("[FATAL] %s: printing failed: %s\n", fname, err)
			}

			// summarize the changes for the user
			summary := fmt.Sprintf("Apply complete. %d resources, %d applied\n", counts.results, counts.ran)
			if UseColor() {
				summary = skittles.Green(summary)
			}
			fmt.Print(summary)
		}
	},
}

func init() {
	addParamsArguments(applyCmd.PersistentFlags())
	viperBindPFlags(applyCmd.Flags())

	RootCmd.AddCommand(applyCmd)
}
