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
	"log"
	"os"
	"runtime"

	"github.com/asteris-llc/converge/prettyprinters"
	"github.com/asteris-llc/converge/prettyprinters/human"
	"github.com/asteris-llc/converge/render"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// bind a set of PFlags to Viper, failing and exiting on error
func viperBindPFlags(flags *pflag.FlagSet) {
	if err := viper.BindPFlags(flags); err != nil {
		log.Fatalf("[FATAL] could not bind flags: %s", err)
	}
}

func getPrinter() prettyprinters.Printer {
	filter := human.ShowEverything
	if !viper.GetBool("show-meta") {
		filter = human.HideByKind("module", "param", "root")
	}
	if viper.GetBool("only-show-changes") {
		filter = human.AndFilter(human.ShowOnlyChanged, filter)
	}

	printer := human.NewFiltered(filter)
	printer.Color = UseColor()

	return prettyprinters.New(printer)
}

// UseColor tells us whether or not to print colors using ANSI escape sequences
// based on the following: 1. If we're in a color terminal 2. If the user has
// specified the `nocolor` option (deduced via Viper) 3. If we're on Windows.
func UseColor() bool {
	isColorTerminal := isatty.IsTerminal(os.Stdout.Fd()) && (runtime.GOOS != "windows")
	return !viper.GetBool("nocolor") && isColorTerminal
}

// getParams wraps getParamsFromFlags, logging and exiting upon error
func getParams(cmd *cobra.Command) render.Values {
	if !cmd.HasPersistentFlags() {
		log.Fatalf("[FATAL] %s: can't get parameters, command doesn't have persistent flags\n", cmd.Name())
	}

	params, errors := getParamsFromFlags(cmd.PersistentFlags())
	for i, err := range errors {
		log.Printf("[ERROR] error while parsing parameters: %s\n", err)

		// after the last error is printed, exit
		if i == len(errors)-1 {
			log.Fatalf("[FATAL] errors while parsing parameters, see log above")
		}
	}
	return params
}
