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

package health

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"text/template"

	"github.com/asteris-llc/converge/graph"
	pp "github.com/asteris-llc/converge/prettyprinters"
	"github.com/asteris-llc/converge/prettyprinters/human"
	"github.com/asteris-llc/converge/resource"
	"github.com/ttacon/chalk"
)

// Printer for health checks
type Printer struct {
	*human.Printer
	Summary bool
}

// healthWrapper wraps a HealthStatus with ID context
type healthWrapper struct {
	*resource.HealthStatus
	ID string
}

// New returns a new Printer with an embedded human printer that hides
// non-healthcheck nodes
func New() *Printer {
	return NewWithPrinter(human.New())
}

// NewWithPrinter uses the provided human printer
func NewWithPrinter(h *human.Printer) *Printer {
	return &Printer{Printer: h}
}

// FinishPP sumarizes the results of the health check
func (p *Printer) FinishPP(g *graph.Graph) (pp.Renderable, error) {
	var warnings int
	var errors int
	var deps int
	type summaryObj struct {
		Warnings int
		Errors   int
		Deps     int
	}
	root, err := g.Root()
	if err != nil {
		return pp.HiddenString(), err
	}
	for _, vertex := range g.Vertices() {
		if vertex == root {
			continue
		}
		status, ok := g.Get(vertex).(*resource.HealthStatus)
		if !ok {
			continue
		}
		if status.IsError() {
			errors++
		}
		if status.IsWarning() {
			warnings++
		}
		if len(status.FailingDeps) > 0 {
			deps++
		}
	}
	tmpl, err := p.template(`{{if (gt .Errors 0)}}{{red "Summary"}}{{else if (gt .Warnings 0)}}{{yellow "Summary"}}{{else}}Summary{{end}}: {{.Errors}} errors, {{.Warnings}} warnings
{{.Deps}} checks will fail due to failing dependencies
`)
	if err != nil {
		fmt.Println("failed to render template")
		return pp.HiddenString(), err
	}
	var out bytes.Buffer
	err = tmpl.Execute(&out, &summaryObj{Warnings: warnings, Errors: errors, Deps: deps})
	return &out, err
}

func (p *Printer) getDrawTemplate() (*template.Template, error) {
	if p.Summary {
		return p.template(`{{if .IsError}}{{red .ID}}{{else if .IsWarning}}{{yellow .ID}}{{else}}{{.ID}}{{end}}: Status: {{showWarning .WarningLevel}}; {{len .FailingDeps}} failing dependencies

`)
	}
	return p.template(`{{if .IsError}}{{red .ID}}{{else if .IsWarning}}{{yellow .ID}}{{else}}{{.ID}}{{end}}: {{showWarning .WarningLevel}}
	Messages:
	{{- range $msg := .Messages}}
	{{indent $msg}}
	{{- end}}
	{{- if .HasChanges}}
	{{- range $key, $values := .Changes}}
	{{red $key}}: {{diff ($values.Original) ($values.Current)}}
	{{- end}}
	{{- end}}
	{{- if .HasFailingDeps}}
	Dependencies Have Failed:
	{{- range $dep, $val := .FailingDeps}}
	{{indent $dep}}: {{$val}}
	{{- end}}
	{{- end}}

`)
}

// DrawNode draws a single health check
func (p *Printer) DrawNode(g *graph.Graph, id string) (pp.Renderable, error) {
	type printerNode struct {
		ID string
		*resource.HealthStatus
	}

	if root, err := g.Root(); root == id || err != nil {
		return pp.HiddenString(), err
	}

	node := g.Get(id)
	healthStatus, ok := node.(*resource.HealthStatus)
	if !ok {
		fmt.Printf("%s is not a health node, deferring to the human printer\n", id)
		return p.Printer.DrawNode(g, id)
	}

	if !healthStatus.ShouldDisplay() {
		return pp.HiddenString(), nil
	}

	tmpl, err := p.getDrawTemplate()

	if err != nil {
		fmt.Println("template generation error")
		return pp.HiddenString(), err
	}

	var out bytes.Buffer
	err = tmpl.Execute(&out, &printerNode{ID: id, HealthStatus: healthStatus})

	return &out, err
}

func mkError(msg string) (pp.Renderable, error) {
	return nil, errors.New(msg)
}

func (p *Printer) template(source string) (*template.Template, error) {
	funcs := map[string]interface{}{
		// colors
		"black":       p.styled(chalk.Black.NewStyle().WithBackground(chalk.ResetColor)),
		"red":         p.styled(chalk.Red.NewStyle().WithBackground(chalk.ResetColor)),
		"green":       p.styled(chalk.Green.NewStyle().WithBackground(chalk.ResetColor)),
		"yellow":      p.styled(chalk.Yellow.NewStyle().WithBackground(chalk.ResetColor)),
		"magenta":     p.styled(chalk.Magenta.NewStyle().WithBackground(chalk.ResetColor)),
		"cyan":        p.styled(chalk.Cyan.NewStyle().WithBackground(chalk.ResetColor)),
		"white":       p.styled(chalk.White.NewStyle().WithBackground(chalk.ResetColor)),
		"indent":      p.indent,
		"diff":        p.diff,
		"showWarning": p.showWarning,
	}
	return template.New("").Funcs(funcs).Parse(source)
}

func (p *Printer) diff(before, after string) (string, error) {
	// remember when modifying these that diff is responsible for leading
	// whitespace
	if !strings.Contains(strings.TrimSpace(before), "\n") && !strings.Contains(strings.TrimSpace(after), "\n") {
		return fmt.Sprintf(" %q => %q", strings.TrimSpace(before), strings.TrimSpace(after)), nil
	}

	tmpl, err := p.template(`before:
{{indent .Before}}
after:
{{indent .After}}`)
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, struct{ Before, After string }{before, after})

	return "\n" + p.indent(p.indent(buf.String())), err
}

func (p *Printer) showWarning(c resource.HealthStatusCode) string {
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

func (p *Printer) indent(in string) string {
	return "\t" + strings.Replace(in, "\n", "\n\t", -1)
}

func (p *Printer) styled(style chalk.Style) func(string) string {
	if !p.Printer.Color {
		return func(in string) string { return in }
	}
	return style.Style
}
