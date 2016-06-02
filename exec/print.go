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
	"fmt"
	"sort"

	"github.com/acmacalister/skittles"
)

// Results is the type of a slice of PlanResults
type Results []*PlanResult

/* Printing */

func (p *PlanResult) string(pretty bool) string {
	// defaults: no colors
	printablePath := p.Path
	printableStatus := p.CurrentStatus
	printableChange := fmt.Sprint(p.WillChange)

	// add color
	if pretty {
		printablePath = skittles.BoldBlack(printablePath)
		if p.WillChange {
			printableStatus = skittles.Yellow(printableStatus)
			printableChange = skittles.Yellow(printableChange)
		} else {
			printableStatus = skittles.Blue(printableStatus)
			printableChange = skittles.Blue(printableChange)
		}
	}

	return fmt.Sprintf(
		"%s:\n\tCurrently: %s\n\tWill Change: %s",
		p.Path,
		printableStatus,
		printableChange,
	)
}

// Pretty prints a PlanResult, optionally with ANSI terminal colors. It is used
// in PlanResult.String and Results.String.
func (p *PlanResult) Pretty(color bool) string {
	return p.string(true)
}

// String satisfies the Stringer interface, and is used to print a string
// representation of a PlanResult.
func (p *PlanResult) String() string {
	return p.string(false)
}

func (rs Results) string(pretty bool) (printMe string) {
	if pretty {
		sort.Sort(rs)
	}
	for _, r := range rs {
		printMe += r.Pretty(pretty)
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

/* Sorting */

// Len is part of sort.Interface.
func (rs Results) Len() int {
	return len(rs)
}

// Swap is part of sort.Interface.
func (rs Results) Swap(i, j int) {
	rs[i], rs[j] = rs[j], rs[i]
}

// Less is part of sort.Interface. It sorts first based on Path, and second on
// WillChange.
func (rs Results) Less(i, j int) bool {
	if rs[i].Path == rs[j].Path {
		// If the result's have the same path, show the one that will change first.
		if rs[i].WillChange && !rs[j].WillChange {
			return true
		}
		return false
	}
	if rs[i].Path < rs[j].Path {
		return true
	}
	return false
}
