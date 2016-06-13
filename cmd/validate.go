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
	"log"

	"github.com/asteris-llc/converge/load"
	"github.com/asteris-llc/converge/resource"
	"github.com/spf13/cobra"
)

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "validate that the syntax of a module file is valid",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("Need at least one module filename as argument, got 0")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		for _, fname := range args {
			_, err := load.Load(fname, resource.Values{})
			if err != nil {
				log.Fatalf("[FATAL] %s: could not parse file: %s\n", fname, err)
			}

			log.Printf("[INFO] %s: module valid\n", fname)
		}
	},
}

func init() {
	RootCmd.AddCommand(validateCmd)
}
