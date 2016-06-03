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

package exec

import (
	"bytes"
	"fmt"
	"sort"
	"text/template"

	"github.com/acmacalister/skittles"
)

// PlanResult contains the result of a resource check
type PlanResult struct {
	Path          string
	CurrentStatus string
	WillChange    bool
}

// Results is the type of a slice of PlanResults
type Results []*PlanResult

func (p *PlanResult) string(pretty bool) string {
	// condFmt returns a function that only formats its input if `pretty` is true.
	condFmt := func(format func(interface{}) string) func(interface{}) string {
		return func(in interface{}) string {
			if pretty {
				return format(in)
			}
			return fmt.Sprint(in)
		}
	}
	funcs := map[string]interface{}{
		"boldBlack": condFmt(skittles.BoldBlack),
		"blueOrYellow": condFmt(func(in interface{}) string {
			if p.WillChange {
				return skittles.Yellow(in)
			}
			return skittles.Blue(in)
		}),
	}
	tmplStr := "{{boldBlack .Path}}:"
	tmplStr += "\n\tCurrently: {{blueOrYellow .CurrentStatus}}"
	tmplStr += "\n\tWill Change: {{blueOrYellow .WillChange}}"
	tmpl := template.Must(template.New("").Funcs(funcs).Parse(tmplStr))

	var buf bytes.Buffer
	_ = tmpl.Execute(&buf, p)
	return buf.String()
}

// Pretty prints a PlanResult, optionally with ANSI terminal colors. It is used
// in PlanResult.String and Results.String.
func (p *PlanResult) Pretty() string {
	return p.string(true)
}

// String satisfies the Stringer interface, and is used to print a string
// representation of a PlanResult.
func (p *PlanResult) String() string {
	return p.string(false)
}

func (rs Results) string(pretty bool) (printMe string) {
	// first, collect string representations of all the PlanResults
	results := []string{}
	for _, r := range rs {
		results = append(results, r.string(pretty))
	}

	// sort them by lexical order, which ends up being module path
	sort.Strings(results)

	// join them together with newlines
	for i, r := range results {
		if i != 0 {
			printMe += "\n"
		}
		printMe += r
	}
	return printMe
}

// Pretty prints Results, optionally with ANSI terminal colors. It is used
// in Results.String.
func (rs Results) Pretty() string {
	return rs.string(true)
}

// String satisfies the Stringer interface, and implements a pretty printer with
// ANSI terminal colors for a slice of PlanResults.
func (rs Results) String() string {
	return rs.string(false)
}
