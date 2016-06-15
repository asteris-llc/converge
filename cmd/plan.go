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

	"golang.org/x/net/context"

	"github.com/acmacalister/skittles"
	"github.com/asteris-llc/converge/exec"
	"github.com/asteris-llc/converge/load"
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
		params := getParams(cmd)

		// set up execution context
		ctx, cancel := context.WithCancel(context.Background())
		GracefulExit(cancel)

		for _, fname := range args {
			log.Printf("[INFO] planning %s\n", fname)

			graph, err := load.Load(fname, params)
			if err != nil {
				log.Fatalf("[FATAL] %s: could not parse file: %s\n", fname, err)
			}

			results, err := exec.Plan(ctx, graph)
			if err != nil {
				log.Fatalf("[FATAL] %s: planning failed: %s\n", fname, err)
			}

			var counts struct {
				results, changes int
			}

			fmt.Print("\n")
			for _, result := range results {
				counts.results++
				if result.WillChange {
					counts.changes++
				}

				if UseColor() {
					fmt.Println(result.Pretty())
				} else {
					fmt.Println(result)
				}
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
	addParamsArguments(planCmd.PersistentFlags())
	RootCmd.AddCommand(planCmd)
}
