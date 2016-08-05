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

package graph_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/asteris-llc/converge/graph"
	"github.com/asteris-llc/converge/helpers"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

func TestSprint(t *testing.T) {
	defer helpers.HideLogs(t)()

	g := sprintTestGraph()
	out, err := g.Sprint(context.Background(), sprintJustPrint, sprintNeverFilter)

	assert.NoError(t, err)
	assert.Equal(t, "root/child: 2\nroot: 1\n", out)
}

func TestSprintFilter(t *testing.T) {
	defer helpers.HideLogs(t)()

	g := sprintTestGraph()
	out, err := g.Sprint(
		context.Background(),
		sprintJustPrint,
		func(id string, val interface{}) bool { return id == "root" },
	)

	assert.NoError(t, err)
	assert.Equal(t, "root: 1\n", out)
}

func TestSprintError(t *testing.T) {
	defer helpers.HideLogs(t)()

	g := sprintTestGraph()
	out, err := g.Sprint(
		context.Background(),
		func(id string, val interface{}) (string, error) { return "", errors.New(id) },
		sprintNeverFilter,
	)

	assert.Equal(t, "", out)
	if assert.Error(t, err) {
		assert.EqualError(t, err, "1 error(s) occurred:\n\n* root/child")
	}
}

func sprintTestGraph() *graph.Graph {
	g := graph.New()

	g.Add("root", 1)
	g.Add("root/child", 2)

	g.Connect("root", "root/child")

	return g
}

func sprintJustPrint(id string, val interface{}) (string, error) {
	return fmt.Sprintf("%s: %v", id, val), nil
}

func sprintNeverFilter(string, interface{}) bool {
	return true
}
