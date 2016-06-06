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
	"strings"
	"text/template"

	"github.com/Sirupsen/logrus"
	"github.com/acmacalister/skittles"
)

// ApplyResult contains the result of a resource check
type ApplyResult struct {
	Path      string
	OldStatus string
	NewStatus string
	Success   bool
}

func (a *ApplyResult) string(pretty bool) string {
	funcs := map[string]interface{}{
		"plusOrMinus": func() string {
			if a.Success {
				if pretty {
					return skittles.Green("+")
				}
				return "+"
			}
			if pretty {
				return skittles.Red("-")
			}
			return "-"
		},
		"redOrGreen": condFmt(pretty, func(in interface{}) string {
			if a.Success {
				return skittles.Green(in)
			}
			return skittles.Red(in)
		}),
		"trimNewline": func(in string) string { return strings.TrimSuffix(in, "\n") },
	}
	tmplStr := `{{plusOrMinus}} {{redOrGreen .Path}}:
	Status: "{{trimNewline .OldStatus}}" => "{{trimNewline .NewStatus}}"
	Success: {{redOrGreen .Success}}`
	tmpl := template.Must(template.New("").Funcs(funcs).Parse(tmplStr))

	var buf bytes.Buffer
	err := tmpl.Execute(&buf, a)
	if err != nil {
		logrus.WithError(err).Warn("error while outputting the result of `apply`")
	}
	return buf.String()
}

// Pretty pretty-prints an ApplyResult with ANSI terminal colors.
func (a *ApplyResult) Pretty() string {
	return a.string(true)
}

// String satisfies the Stringer interface, and is used to print a string
// representation of a ApplyResult.
func (a *ApplyResult) String() string {
	return a.string(false)
}
