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

package tmpltools

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/asteris-llc/converge/resource"
)

// ExtendedTemplate provides information about extended template rendering
type ExtendedTemplate struct {
	Color bool
}

// New generates a new template.Template with the functions from the extended template
func (tmpl *ExtendedTemplate) New(source string) (*template.Template, error) {
	reset := "\x1b[0m"
	funcs := map[string]interface{}{
		// colors
		"black":   tmpl.styled(func(in string) string { return "\x1b[30m" + in + reset }),
		"red":     tmpl.styled(func(in string) string { return "\x1b[31m" + in + reset }),
		"green":   tmpl.styled(func(in string) string { return "\x1b[32m" + in + reset }),
		"yellow":  tmpl.styled(func(in string) string { return "\x1b[33m" + in + reset }),
		"blue":    tmpl.styled(func(in string) string { return "\x1b[34m" + in + reset }),
		"magenta": tmpl.styled(func(in string) string { return "\x1b[35m" + in + reset }),
		"cyan":    tmpl.styled(func(in string) string { return "\x1b[36m" + in + reset }),
		"white":   tmpl.styled(func(in string) string { return "\x1b[37m" + in + reset }),

		// utils
		"indent":      tmpl.indent,
		"diff":        tmpl.diff,
		"showWarning": showWarning,
	}
	return template.New("").Funcs(funcs).Parse(source)
}

func (tmpl *ExtendedTemplate) styled(style func(string) string) func(string) string {
	if !tmpl.Color {
		return func(in string) string { return in }
	}

	return style
}

func (tmpl *ExtendedTemplate) diff(before, after string) (string, error) {
	// remember when modifying these that diff is responsible for leading
	// whitespace
	if !strings.Contains(strings.TrimSpace(before), "\n") && !strings.Contains(strings.TrimSpace(after), "\n") {
		return fmt.Sprintf(" %q => %q", strings.TrimSpace(before), strings.TrimSpace(after)), nil
	}

	newTmpl, err := tmpl.New(`before:
{{indent .Before}}
after:
{{indent .After}}`)
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)
	err = newTmpl.Execute(buf, struct{ Before, After string }{before, after})

	return "\n" + tmpl.indent(tmpl.indent(buf.String())), err
}

func (tmpl *ExtendedTemplate) indent(in string) string {
	return "\t" + strings.Replace(in, "\n", "\n\t", -1)
}

func (tmpl *ExtendedTemplate) empty(s string) bool {
	return s == ""
}

func showWarning(c resource.HealthStatusCode) string {
	switch c {
	case resource.StatusHealthy:
		return "Healthy"
	case resource.StatusWarning:
		return "Warning"
	case resource.StatusError:
		return "Error"
	default:
		return "Fatal: Unkown Error"
	}
}

// Run creates a template extension with the provided color setting and then
// runs it with the provided template source
func Run(color bool, source string) (*template.Template, error) {
	return (&ExtendedTemplate{Color: color}).New(source)
}
