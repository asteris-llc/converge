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
	human.Printer
}

// healthWrapper wraps a HealthStatus with ID context
type healthWrapper struct {
	*resource.HealthStatus
	ID string
}

// New returns a new Printer with an embedded human printer that hides
// non-healthcheck nodes
func New() *Printer {
	return &Printer{*human.New()}
}

// FinishPP sumarizes the results of the health check
func (p *Printer) FinishPP(g *graph.Graph) (pp.Renderable, error) {
	return pp.VisibleString("==End of Graph==\n"), nil
}

// DrawNode draws a single health check
func (p *Printer) DrawNode(g *graph.Graph, id string) (pp.Renderable, error) {
	check, ok := g.Get(id).(resource.Check)
	if !ok {
		return pp.HiddenString(), nil
	}
	status, err := check.HealthCheck()
	if err != nil {
		return pp.HiddenString(), err
	}

	if !status.ShouldDisplay() {
		return pp.HiddenString(), nil
	}

	fmt.Printf("status: %v\n", status)

	tmpl, err := p.template(`{{if .IsError}}{{red .ID}}{{else if .IsWarning}}{{yellow .ID}}{{else}}{{.ID}}{{end}}:
	Messages
`)
	if err != nil {
		return nil, err
	}

	var out bytes.Buffer
	wrapper := &healthWrapper{HealthStatus: status, ID: id}
	err = tmpl.Execute(&out, wrapper)
	return pp.VisibleString(out.String()), err
}

func mkError(msg string) (pp.Renderable, error) {
	return nil, errors.New(msg)
}

func (p *Printer) template(source string) (*template.Template, error) {
	funcs := map[string]interface{}{
		// colors
		"black":   p.styled(chalk.Black.NewStyle().WithBackground(chalk.ResetColor)),
		"red":     p.styled(chalk.Red.NewStyle().WithBackground(chalk.ResetColor)),
		"green":   p.styled(chalk.Green.NewStyle().WithBackground(chalk.ResetColor)),
		"yellow":  p.styled(chalk.Yellow.NewStyle().WithBackground(chalk.ResetColor)),
		"magenta": p.styled(chalk.Magenta.NewStyle().WithBackground(chalk.ResetColor)),
		"cyan":    p.styled(chalk.Cyan.NewStyle().WithBackground(chalk.ResetColor)),
		"white":   p.styled(chalk.White.NewStyle().WithBackground(chalk.ResetColor)),
		"indent":  p.indent,
	}
	return template.New("").Funcs(funcs).Parse(source)
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
