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
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"github.com/spf13/viper"
)

const manSection = "8" // System Administration tools and daemons

// manCmd represents the man command
var manCmd = &cobra.Command{
	Use:   "man",
	Short: "generate man pages for Converge",
	Long: `Generate man pages for Converge

By default, this places man pages into the "man/man` + manSection + `" directory under
the current directory. Use "--path=PATH" to override the output directory. For
example, to instally man pages globally on many Unix-like systems, use
"--path=/usr/local/share/man/man` + manSection + `".`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := os.MkdirAll(viper.GetString("path"), os.FileMode(0755)); err != nil {
			logrus.WithError(err).Fatal("could not create man tree path")
		}

		header := &doc.GenManHeader{
			Title:   "Converge",
			Section: manSection,
		}
		if err := doc.GenManTree(RootCmd, header, viper.GetString("path")); err != nil {
			logrus.WithError(err).Fatal("could not generate man tree")
		}
	},
}

func init() {
	manCmd.Flags().String("path", "man/man"+manSection, "path to generated man pages")

	genCmd.AddCommand(manCmd)
}
