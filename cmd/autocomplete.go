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
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// autocompleteCmd represents the autocomplete command
var autocompleteCmd = &cobra.Command{
	Use:   "autocomplete",
	Short: "generate bash autocompletion script for Converge",
	Long: `By default, completion file is written to ./converge.bash. Use
"--out=/path/to/file" to override the file location.

Note that for the generated file to work on OS X/macOS, you'll need to install
bash-completion (or equivalent) from Homebrew (or your package manager of
choice.)`,
	Run: func(cmd *cobra.Command, args []string) {
		RootCmd.GenBashCompletionFile(viper.GetString("out"))
	},
}

func init() {
	autocompleteCmd.Flags().String("out", "./converge.bash", "path to generated autocomplete file")

	genCmd.AddCommand(autocompleteCmd)
}
