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

	"github.com/Sirupsen/logrus"
	"github.com/asteris-llc/converge/exec"
	"github.com/asteris-llc/converge/load"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// applyCmd represents the plan command
var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "apply what needs to change in the system",
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

		for _, fname := range args {
			logger := logrus.WithField("filename", fname)

			graph, err := load.Load(fname, params)
			if err != nil {
				logger.WithError(err).Fatal("could not parse file")
			}

			plan, err := exec.Plan(graph)
			if err != nil {
				logger.WithError(err).Fatal("planning failed")
			}

			results, err := exec.Apply(graph, plan)
			if err != nil {
				logger.WithError(err).Fatal("applying failed")
			}

			var failed bool

			for _, result := range results {
				if !result.Success {
					failed = true
				}
				fmt.Println(result)
			}

			if failed {
				os.Exit(1)
			}
		}
	},
}

func init() {
	addParamsArguments(applyCmd.Flags())
	viper.BindPFlags(applyCmd.Flags())

	RootCmd.AddCommand(applyCmd)
}
