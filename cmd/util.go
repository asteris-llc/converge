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
	"runtime"

	"github.com/asteris-llc/converge/prettyprinters"
	"github.com/asteris-llc/converge/prettyprinters/health"
	"github.com/asteris-llc/converge/prettyprinters/human"
	"github.com/asteris-llc/converge/resource"
	"github.com/mattn/go-isatty"
	"github.com/spf13/viper"
)

func humanProvider(filter human.FilterFunc) *human.Printer {
	if !viper.GetBool("show-meta") {
		filter = human.HideByKind("module", "param", "root")
	}
	if viper.GetBool("only-show-changes") {
		filter = human.AndFilter(human.ShowOnlyChanged, filter)
	}

	printer := human.NewFiltered(filter)
	printer.Color = UseColor()
	printer.InitColors()
	return printer
}

func getPrinter() prettyprinters.Printer {
	return prettyprinters.New(humanProvider(human.ShowEverything))
}

func healthPrinter() prettyprinters.Printer {
	showHealthNodes := func(id string, value human.Printable) bool {
		_, ok := value.(*resource.HealthStatus)
		return ok
	}
	provider := humanProvider(showHealthNodes)
	health := health.NewWithPrinter(provider)
	health.Summary = viper.GetBool("quiet")
	return prettyprinters.New(health)
}

// UseColor tells us whether or not to print colors using ANSI escape sequences
// based on the following: 1. If we're in a color terminal 2. If the user has
// specified the `nocolor` option (deduced via Viper) 3. If we're on Windows.
func UseColor() bool {
	isColorTerminal := isatty.IsTerminal(os.Stdout.Fd()) && (runtime.GOOS != "windows")
	return !viper.GetBool("nocolor") && isColorTerminal
}
