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
	"bytes"
	"errors"
	"io/ioutil"
	"log"
	"os"

	"github.com/hashicorp/hcl/hcl/printer"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// fmtCmd represents the fmt command
var fmtCmd = &cobra.Command{
	Use:   "fmt",
	Short: "format a source file",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("Need at least one module filename as argument, got 0")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		for _, fname := range args {
			content, err := ioutil.ReadFile(fname)
			if err != nil {
				log.Fatalf("[FATAL] %s: could not read %s\n", fname, err)
			}

			formatted, err := printer.Format(content)
			if err != nil {
				log.Fatalf("[FATAL] %s: could not format content\n", fname)
			}

			if viper.GetBool("check") {
				if !bytes.Equal(content, formatted) {
					log.Fatalf("[FATAL] %s: needs formatting\n", fname)
				}
			} else {
				stat, err := os.Stat(fname)
				if err != nil {
					log.Fatalf("[FATAL] %s: could not stat: %s\n", fname, err)
				}

				err = ioutil.WriteFile(fname, formatted, stat.Mode())
				if err != nil {
					log.Fatalf("[FATAL] %s: could not write content: %s\n", fname, err)
				}
			}
		}
	},
}

func init() {
	fmtCmd.Flags().Bool("check", false, "only check, no writing")

	RootCmd.AddCommand(fmtCmd)
}
