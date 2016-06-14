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

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	"github.com/acmacalister/skittles"
	"github.com/asteris-llc/converge/exec"
	"github.com/asteris-llc/converge/load"
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
			logger := logrus.WithField("filename", fname)

			graph, err := load.Load(fname, params)
			if err != nil {
				logger.WithError(err).Fatal("could not parse file")
			}

			planStatus := make(chan *exec.StatusMessage, 1)
			go func() {
				for msg := range planStatus {
					logger.WithField("path", msg.Path).Info(msg.Status)
				}
				close(planStatus)
			}()

			plan, err := exec.PlanWithStatus(ctx, graph, planStatus)
			if err != nil {
				logger.WithError(err).Fatal("planning failed")
			}

			status := make(chan *exec.StatusMessage, 1)
			go func() {
				for msg := range status {
					logger.WithField("path", msg.Path).Info(msg.Status)
				}
				close(status)
			}()

			results, err := exec.ApplyWithStatus(ctx, graph, plan, status)
			if err != nil {
				logger.WithError(err).Fatal("applying failed")
			}

			// count successes and failures to print summary
			var counts struct {
				results, success, failures int
			}

			for _, result := range results {
				counts.results++
				if result.Success {
					counts.success++
				} else {
					counts.failures++
				}

				if UseColor() {
					fmt.Println(result.Pretty())
				} else {
					fmt.Println(result)
				}
			}

			// summarize the changes for the user
			summary := fmt.Sprintf("\nApply complete. %d changes, %d successful, %d failed\n", counts.results, counts.success, counts.failures)
			if UseColor() {
				if counts.failures > 0 {
					summary = skittles.Red(summary)
				} else {
					summary = skittles.Green(summary)
				}
			}
			fmt.Print(summary)

			if counts.failures > 0 {
				os.Exit(1)
			}
		}
	},
}

func init() {
	addParamsArguments(applyCmd.PersistentFlags())
	viperBindPFlags(applyCmd.Flags())

	RootCmd.AddCommand(applyCmd)
}
