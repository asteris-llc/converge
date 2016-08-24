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
	"github.com/ttacon/chalk"
)

// ExtendedTemplate provides information about extended template rendering
type ExtendedTemplate struct {
	Color bool
}

// New generates a new template.Template with the functions from the extended template
func (tmpl *ExtendedTemplate) New(source string) (*template.Template, error) {
	funcs := map[string]interface{}{
		// colors
		"black":       tmpl.styled(chalk.Black.NewStyle().WithBackground(chalk.ResetColor)),
		"red":         tmpl.styled(chalk.Red.NewStyle().WithBackground(chalk.ResetColor)),
		"green":       tmpl.styled(chalk.Green.NewStyle().WithBackground(chalk.ResetColor)),
		"yellow":      tmpl.styled(chalk.Yellow.NewStyle().WithBackground(chalk.ResetColor)),
		"magenta":     tmpl.styled(chalk.Magenta.NewStyle().WithBackground(chalk.ResetColor)),
		"cyan":        tmpl.styled(chalk.Cyan.NewStyle().WithBackground(chalk.ResetColor)),
		"white":       tmpl.styled(chalk.White.NewStyle().WithBackground(chalk.ResetColor)),
		"indent":      tmpl.indent,
		"diff":        tmpl.diff,
		"showWarning": showWarning,
	}
	return template.New("").Funcs(funcs).Parse(source)
}

func (tmpl *ExtendedTemplate) styled(style chalk.Style) func(string) string {
	if !tmpl.Color {
		return func(in string) string { return in }
	}

	return style.Style
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
