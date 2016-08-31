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
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
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

	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// set log level
		level, err := cmd.Flags().GetString("log-level")
		if err != nil {
			return err
		}

		parsedLevel, err := log.ParseLevel(level)
		if err != nil {
			return err
		}

		log.SetLevel(parsedLevel)

		// bind pflags for active commands
		sub := cmd
		subFlags := args

		for {
			log.WithField("command", sub.Name()).Debug("registering flags")

			if err := viper.BindPFlags(sub.Flags()); err != nil {
				return errors.Wrapf(err, "failed to bind flags for %s", sub.Name())
			}
			if err := viper.BindPFlags(sub.PersistentFlags()); err != nil {
				return errors.Wrapf(err, "failed to bind persistent flags for %s", sub.Name())
			}

			potentialSub, potentialSubFlags, err := sub.Find(subFlags)
			if err != nil {
				return errors.Wrapf(err, "failed to get child for %s", sub.Name())
			}

			if sub == potentialSub {
				break
			}

			sub = potentialSub
			subFlags = potentialSubFlags
		}

		return nil
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

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is /etc/converge/config.yaml)")
	RootCmd.PersistentFlags().BoolP("nocolor", "n", false, "force colorless output")
	RootCmd.PersistentFlags().StringP("log-level", "l", "INFO", "log level, one of debug, info, warning, error, or fatal")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	viper.SetConfigName("config")        // name of config file (without extension)
	viper.AddConfigPath("/etc/converge") // adding home directory as first search path
	viper.SetEnvPrefix("CONVERGE")       // so our environment variables are unambiguous
	viper.AutomaticEnv()                 // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
