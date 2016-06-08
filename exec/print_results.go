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
	"strings"
)

// Results is the type of a slice of PlanResults or ApplyResults (both of which
// statisfy the prettyPrinter interface.)
type Results []prettyPrinter

type prettyPrinter interface {
	fmt.Stringer
	Pretty() string
}

// Print implements a pretty printer that uses ANSI terminal colors when a color
// terminal is available.
func (rs Results) Print(color bool) string {
	// first, collect string representations of all the PlanResults
	results := []string{}
	for _, r := range rs {
		if color {
			results = append(results, r.Pretty())
		} else {
			results = append(results, r.String())
		}
	}

	// sort them by lexical order, which ends up being module path
	sort.Strings(results)

	return strings.Join(results, "\n\n")
}
