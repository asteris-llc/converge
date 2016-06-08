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

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	"github.com/acmacalister/skittles"
	"github.com/asteris-llc/converge/exec"
	"github.com/asteris-llc/converge/load"
	"github.com/spf13/cobra"
)

// planCmd represents the plan command
var planCmd = &cobra.Command{
	Use:   "plan",
	Short: "plan what needs to change in the system",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("Need at least one module filename as argument, got 0")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		params, err := getParamsFromFlags()
		if err != nil {
			logrus.WithError(err).Fatal("could not load params")
		}

		// set up execution context
		ctx, cancel := context.WithCancel(context.Background())
		GracefulExit(cancel)

		for _, fname := range args {
			logger := logrus.WithField("filename", fname)

			graph, err := load.Load(fname, params)
			if err != nil {
				logger.WithError(err).Fatal("could not parse file")
			}

			status := make(chan *exec.StatusMessage, 1)
			go func() {
				for msg := range status {
					logger.WithField("path", msg.Path).Info(msg.Status)
				}
				close(status)
			}()

			results, err := exec.PlanWithStatus(ctx, graph, status)
			if err != nil {
				logger.WithError(err).Fatal("planning failed")
			}

			// giving the results the type exec.Results allows us to pretty-print them
			var typedResults exec.Results
			for _, result := range results {
				typedResults = append(typedResults, result)
			}
			fmt.Println(typedResults.Print(UseColor()))

			// count results and changes to print summary
			var counts struct {
				results, changes int
			}

			for _, result := range results {
				counts.results++
				if result.WillChange {
					counts.changes++
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
	addParamsArguments(planCmd.Flags())
	viperBindPFlags(planCmd.Flags())

	RootCmd.AddCommand(planCmd)
}
