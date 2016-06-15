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
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// Name describes the name for packaging
const Name = "converge"

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   Name,
	Short: Name + " applies changes to systems over a graph",
	Long: Name + ` is a tool that reads modules files (see the samples directory
in the source) and applies their actions to a system.

The workflow generally looks like this:

1. write your module files
2. see what changes will happen with "converge plan yourfile.hcl"
3. apply the changes with "converge apply yourfile.hcl"

You can also visualize the execution graph with "converge graph yourfile.hcl" -
see "converge graph --help" for more details.`,

	PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
		level, err := cmd.Flags().GetString("log-level")
		if err != nil {
			return err
		}

		return SetLogLevel(level)
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.converge.yaml)")
	RootCmd.PersistentFlags().BoolP("nocolor", "n", false, "force colorless output")
	RootCmd.PersistentFlags().StringP("log-level", "l", "INFO", fmt.Sprintf("log level, one of %v", levels))

	viperBindPFlags(RootCmd.PersistentFlags())
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName(".converge") // name of config file (without extension)
	viper.AddConfigPath("$HOME")     // adding home directory as first search path
	viper.AutomaticEnv()             // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
