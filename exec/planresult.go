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
	"html/template"
	"strings"

	"github.com/acmacalister/skittles"
)

// PlanResult contains the result of a resource check
type PlanResult struct {
	Path          string
	CurrentStatus string
	WillChange    bool
}

func (p *PlanResult) string(pretty bool) string {
	funcs := map[string]interface{}{
		"boldBlack": condFmt(pretty, skittles.BoldBlack),
		"blueOrYellow": condFmt(pretty, func(in interface{}) string {
			if p.WillChange {
				return skittles.Yellow(in)
			}
			return skittles.Blue(in)
		}),
		"trimNewline": func(in string) string { return strings.TrimSuffix(in, "\n") },
	}
	tmplStr := "{{boldBlack (trimNewline .Path)}}:"
	tmplStr += "\n\tCurrently: {{blueOrYellow (trimNewline .CurrentStatus)}}"
	tmplStr += "\n\tWill Change: {{blueOrYellow (trimNewline .WillChange)}}"
	tmpl := template.Must(template.New("").Funcs(funcs).Parse(tmplStr))

	var buf bytes.Buffer
	_ = tmpl.Execute(&buf, p)
	return buf.String()
}

// Pretty pretty-prints a PlanResult with ANSI terminal colors.
func (p *PlanResult) Pretty() string {
	return p.string(true)
}

// String satisfies the Stringer interface, and is used to print a string
// representation of a PlanResult.
func (p *PlanResult) String() string {
	return p.string(false)
}
