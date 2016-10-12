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

package human_test

import (
	"errors"
	"testing"

	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/graph/node"
	pp "github.com/asteris-llc/converge/prettyprinters"
	"github.com/asteris-llc/converge/prettyprinters/human"
	"github.com/asteris-llc/converge/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	defaultPrinter     = human.New()
	printerOnlyChanged = human.NewFiltered(human.ShowOnlyChanged)
	printerHideByKind  = human.NewFiltered(human.HideByKind("param"))
)

func TestSatisfiesInterface(t *testing.T) {
	t.Parallel()

	assert.Implements(t, (*pp.BasePrinter)(nil), new(human.Printer))
	assert.Implements(t, (*pp.NodePrinter)(nil), new(human.Printer))
	assert.Implements(t, (*pp.GraphPrinter)(nil), new(human.Printer))
}

func testFinishPP(t *testing.T, in Printable, out string) {
	g := graph.New()
	g.Add(node.New("root", in))

	printer := human.New()
	printer.InitColors()
	str, err := printer.FinishPP(g)

	require.Nil(t, err)
	assert.Equal(t, out, str.String())
}

func TestFinishPPSuccess(t *testing.T) {
	t.Parallel()

	testFinishPP(
		t,
		Printable{"a": "b"},
		"Summary: 0 errors, 1 changes\n",
	)
}

func TestFinishPPError(t *testing.T) {
	t.Parallel()

	testFinishPP(
		t,
		Printable{"error": "test"},
		"Summary: 1 errors, 1 changes\n\n * root: test\n",
	)
}

func testDrawNodes(t *testing.T, in Printable, out string) {
	printer := human.New()
	printer.InitColors()
	testDrawNodesCustomPrinter(
		t,
		printer,
		"root",
		in,
		out,
	)
}

func testDrawNodesCustomPrinter(t *testing.T, h *human.Printer, id string, in Printable, out string) {
	g := graph.New()
	g.Add(node.New(id, in))

	str, err := h.DrawNode(g, id)

	require.Nil(t, err)
	assert.Equal(t, out, str.String())
}

func benchmarkDrawNodes(in Printable) {
	benchmarkDrawNodesCustomPrinter(
		defaultPrinter,
		"root",
		in,
	)
}

func benchmarkDrawNodesCustomPrinter(h *human.Printer, id string, in Printable) {
	g := graph.New()
	g.Add(node.New(id, in))

	h.DrawNode(g, id)
}

func TestDrawNodeNoChanges(t *testing.T) {
	t.Parallel()

	testDrawNodes(
		t,
		Printable{},
		"root:\n Messages:\n Has Changes: no\n Changes: No changes\n\n",
	)
}

func BenchmarkDrawNodeNoChanges(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkDrawNodes(
			Printable{},
		)
	}
}

func TestDrawNodeNoChangesFiltered(t *testing.T) {
	t.Parallel()
	printerOnlyChanged.InitColors()

	testDrawNodesCustomPrinter(
		t,
		printerOnlyChanged,
		"root",
		Printable{},
		"",
	)
}

func BenchmarkDrawNodeNoChangesFiltered(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkDrawNodesCustomPrinter(
			printerOnlyChanged,
			"root",
			Printable{},
		)
	}
}

func TestDrawNodeMetaFiltered(t *testing.T) {
	t.Parallel()
	printerHideByKind.InitColors()

	testDrawNodesCustomPrinter(
		t,
		printerHideByKind,
		"param.test",
		Printable{},
		"",
	)
}

func BenchmarkDrawNodeMetaFiltered(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkDrawNodesCustomPrinter(
			printerHideByKind,
			"param.test",
			Printable{},
		)
	}
}

func TestDrawNodeChanges(t *testing.T) {
	t.Parallel()

	testDrawNodes(
		t,
		Printable{"a": "b"},
		"root:\n Messages:\n Has Changes: yes\n Changes:\n  a: \"\" => \"b\"\n\n",
	)
}

func BenchmarkDrawNodeChanges(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkDrawNodes(
			Printable{"a": "b"},
		)
	}
}

func TestDrawNodeError(t *testing.T) {
	t.Parallel()

	testDrawNodes(
		t,
		Printable{"error": "x"},
		"root:\n Error: x\n Messages:\n Has Changes: yes\n Changes:\n  error: \"\" => \"x\"\n\n",
	)
}

func BenchmarkDrawNodeError(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkDrawNodes(
			Printable{"error": "x"},
		)
	}
}

// printable stub

type Printable map[string]string

func (p Printable) Messages() []string {
	return []string{}
}

func (p Printable) Changes() map[string]resource.Diff {
	out := map[string]resource.Diff{}

	for key, value := range p {
		out[key] = resource.TextDiff{Values: [2]string{"", value}}
	}

	return out
}

func (p Printable) HasChanges() bool {
	return len(p) > 0
}

func (p Printable) Error() error {
	err, ok := p["error"]
	if !ok {
		return nil
	}

	return errors.New(err)
}
