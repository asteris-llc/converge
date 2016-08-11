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
	pp "github.com/asteris-llc/converge/prettyprinters"
	"github.com/asteris-llc/converge/prettyprinters/human"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSatisfiesInterface(t *testing.T) {
	t.Parallel()

	assert.Implements(t, (*pp.BasePrinter)(nil), new(human.Printer))
	assert.Implements(t, (*pp.NodePrinter)(nil), new(human.Printer))
	assert.Implements(t, (*pp.GraphPrinter)(nil), new(human.Printer))
}

func testFinishPP(t *testing.T, in Printable, out string) {
	g := graph.New()
	g.Add("root", in)

	printer := human.New()
	str, err := printer.FinishPP(g)

	require.Nil(t, err)
	assert.Equal(t, out, str.String())
}

func TestFinishPPSuccess(t *testing.T) {
	t.Parallel()

	testFinishPP(
		t,
		Printable{"a": "b"},
		"Summary: 0 errors, 1 changes",
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
	testDrawNodesCustomPrinter(
		t,
		human.New(),
		"root",
		in,
		out,
	)
}

func testDrawNodesCustomPrinter(t *testing.T, h *human.Printer, id string, in Printable, out string) {
	g := graph.New()
	g.Add(id, in)

	str, err := h.DrawNode(g, id)

	require.Nil(t, err)
	assert.Equal(t, out, str.String())
}

func TestDrawNodeNoChanges(t *testing.T) {
	t.Parallel()

	testDrawNodes(
		t,
		Printable{},
		"root:\n  Has Changes: no\n  Fields:\n    No changes\n\n",
	)
}

func TestDrawNodeNoChangesFiltered(t *testing.T) {
	t.Parallel()

	testDrawNodesCustomPrinter(
		t,
		human.NewFiltered(human.ShowOnlyChanged),
		"root",
		Printable{},
		"",
	)
}

func TestDrawNodeMetaFiltered(t *testing.T) {
	t.Parallel()

	testDrawNodesCustomPrinter(
		t,
		human.NewFiltered(human.HideIDTypes("param")),
		"param.test",
		Printable{},
		"",
	)
}

func TestDrawNodeChanges(t *testing.T) {
	t.Parallel()

	testDrawNodes(
		t,
		Printable{"a": "b"},
		"root:\n  Has Changes: yes\n  Fields:\n    a: \"\" => \"b\"\n\n",
	)
}

func TestDrawNodeError(t *testing.T) {
	t.Parallel()

	testDrawNodes(
		t,
		Printable{"error": "x"},
		"root:\n  Error: x\n  Has Changes: yes\n  Fields:\n    error: \"\" => \"x\"\n\n",
	)
}

// printable stub

type Printable map[string]string

func (p Printable) Fields() map[string][2]string {
	out := map[string][2]string{}

	for key, value := range p {
		out[key] = [2]string{"", value}
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
