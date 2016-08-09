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

package jsonl_test

import (
	"fmt"
	"testing"

	"github.com/asteris-llc/converge/graph"
	pp "github.com/asteris-llc/converge/prettyprinters"
	"github.com/asteris-llc/converge/prettyprinters/jsonl"
	"github.com/stretchr/testify/assert"
)

func TestSatisfiesInterface(t *testing.T) {
	t.Parallel()

	assert.Implements(t, (*pp.BasePrinter)(nil), new(jsonl.JSONPrinter))
	assert.Implements(t, (*pp.NodePrinter)(nil), new(jsonl.JSONPrinter))
	assert.Implements(t, (*pp.EdgePrinter)(nil), new(jsonl.JSONPrinter))
}

func TestDrawNode(t *testing.T) {
	g := graph.New()
	g.Add("x", 1)

	printer := new(jsonl.JSONPrinter)
	out, err := printer.DrawNode(g, "x")

	assert.NoError(t, err)
	assert.Equal(t, `{"id":"x","value":1}`+"\n", fmt.Sprint(out))
}

func TestDrawEdge(t *testing.T) {
	g := graph.New()

	printer := new(jsonl.JSONPrinter)
	out, err := printer.DrawEdge(g, "x", "x/y")

	assert.NoError(t, err)
	assert.Equal(t, `{"source":"x","destination":"x/y"}`+"\n", fmt.Sprint(out))
}
