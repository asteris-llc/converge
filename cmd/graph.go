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

	"github.com/Sirupsen/logrus"
	"github.com/asteris-llc/converge/load"
	"github.com/spf13/cobra"
)

// graphCmd represents the check command
var graphCmd = &cobra.Command{
	Use:   "graph",
	Short: "graph the execution of a module",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("Need at least one module filename as argument, got 0")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		for _, fname := range args {
			logger := logrus.WithField("filename", fname)

			graph, err := load.Load(fname)
			if err != nil {
				logger.WithError(err).Fatal("could not parse file")
			}

			fmt.Println(graph.GraphString())
		}
	},
}

func init() {
	RootCmd.AddCommand(graphCmd)
}
