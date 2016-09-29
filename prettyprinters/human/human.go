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

package human

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
	"text/tabwriter"
	"text/template"

	"github.com/asteris-llc/converge/graph"
	pp "github.com/asteris-llc/converge/prettyprinters"
	"github.com/pkg/errors"
)

// Printer for human-readable output
type Printer struct {
	Color  bool // color output
	Filter FilterFunc
}

var (
	funcs   = map[string]interface{}{}
	funcsMu sync.Mutex
)

// New returns a base version of Printer
func New() *Printer {
	return NewFiltered(ShowEverything)
}

// NewFiltered returns a version of Printer that will filter according to the
// specified func
func NewFiltered(f FilterFunc) *Printer {
	return &Printer{Filter: f}
}

// InitColors initializes the colors used by the human printer
func (p *Printer) InitColors() {
	reset := "\x1b[0m"
	p.funcsMapWrite("black", p.styled(func(in string) string { return "\x1b[30m" + in + reset }))
	p.funcsMapWrite("red", p.styled(func(in string) string { return "\x1b[31m" + in + reset }))
	p.funcsMapWrite("green", p.styled(func(in string) string { return "\x1b[32m" + in + reset }))
	p.funcsMapWrite("yellow", p.styled(func(in string) string { return "\x1b[33m" + in + reset }))
	p.funcsMapWrite("blue", p.styled(func(in string) string { return "\x1b[34m" + in + reset }))
	p.funcsMapWrite("magenta", p.styled(func(in string) string { return "\x1b[35m" + in + reset }))
	p.funcsMapWrite("cyan", p.styled(func(in string) string { return "\x1b[36m" + in + reset }))
	p.funcsMapWrite("white", p.styled(func(in string) string { return "\x1b[37m" + in + reset }))
}

// StartPP does nothing, but is required to satisfy the GraphPrinter interface
func (p *Printer) StartPP(g *graph.Graph) (pp.Renderable, error) {
	return pp.HiddenString(), nil
}

// FinishPP provides summary statistics about the printed graph
func (p *Printer) FinishPP(g *graph.Graph) (pp.Renderable, error) {
	tmpl, err := p.template("{{if gt (len .Errors) 0}}{{red \"Summary\"}}{{else}}{{green \"Summary\"}}{{end}}: {{len .Errors}} errors, {{.ChangesCount}} changes{{if .Errors}}\n{{range .Errors}}\n * {{.}}{{end}}{{end}}\n")
	if err != nil {
		return pp.HiddenString(), err
	}

	counts := struct {
		ChangesCount int
		Errors       []error
	}{}

	for _, id := range g.Vertices() {
		printable, ok := g.Get(id).(Printable)
		if !ok {
			continue
		}

		if printable.HasChanges() {
			counts.ChangesCount++
		}

		if err = printable.Error(); err != nil {
			counts.Errors = append(
				counts.Errors,
				errors.Wrap(err, id),
			)
		}
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, counts)

	return &buf, err
}

// DrawNode containing a result
func (p *Printer) DrawNode(g *graph.Graph, id string) (pp.Renderable, error) {
	printable, ok := g.Get(id).(Printable)
	if !ok {
		return pp.HiddenString(), errors.New("cannot print values that don't implement Printable")
	}

	if !p.Filter(id, printable) {
		return pp.HiddenString(), nil
	}

	tmpl, err := p.template(`{{if .Error}}{{red .ID}}{{else if .HasChanges}}{{yellow .ID}}{{else}}{{.ID}}{{end}}:
	{{- if .Error}}
	{{red "Error"}}: {{.Error}}
	{{- end}}
	Messages:
	{{- range $msg := .Messages}}
	{{indent $msg}}
	{{- end}}
	Has Changes: {{if .HasChanges}}{{yellow "yes"}}{{else}}no{{end}}
	Changes:
		{{- range $key, $values := .Changes}}
		{{cyan $key}}:	{{diff ($values.Original) ($values.Current)}}
		{{- else}} No changes {{- end}}

`)
	if err != nil {
		return pp.HiddenString(), err
	}

	var intermediate, out bytes.Buffer
	err = tmpl.Execute(&intermediate, &printerNode{ID: id, Printable: printable})
	if err != nil {
		return pp.HiddenString(), err
	}

	tabWriter := tabwriter.NewWriter(&out, 4, 4, 4, ' ', tabwriter.Debug)
	_, err = tabWriter.Write(intermediate.Bytes())

	return &out, err
}

func (p *Printer) funcsMapWrite(key string, value interface{}) {
	funcsMu.Lock()
	defer funcsMu.Unlock()
	funcs[key] = value
}

func (p *Printer) template(source string) (*template.Template, error) {
	p.funcsMapWrite("diff", p.diff)
	p.funcsMapWrite("indent", p.indent)
	p.funcsMapWrite("empty", p.empty)

	funcsMu.Lock()
	defer funcsMu.Unlock()
	return template.New("").Funcs(funcs).Parse(source)
}

func (p *Printer) styled(style func(string) string) func(string) string {
	if !p.Color {
		return func(in string) string { return in }
	}

	return style
}

func (p *Printer) diff(before, after string) (string, error) {
	// remember when modifying these that diff is responsible for leading
	// whitespace
	if !strings.Contains(strings.TrimSpace(before), "\n") && !strings.Contains(strings.TrimSpace(after), "\n") {
		return fmt.Sprintf("%q\t=>\t%q", strings.TrimSpace(before), strings.TrimSpace(after)), nil
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

func (p *Printer) indent(in string) string {
	return "\t" + strings.Replace(in, "\n", "\n\t", -1)
}
func (p *Printer) empty(s string) bool {
	return s == ""
}
