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

package shell_test

import (
	"testing"

	"github.com/asteris-llc/converge/resource/shell"
	"github.com/stretchr/testify/assert"
)

func Test_OutputMap_SetsOnlyPresentFields(t *testing.T) {
	var c *shell.CommandResults
	tmpMap := c.OutputMap()
	assert.Equal(t, 0, len(tmpMap))
	c = &shell.CommandResults{}
	assert.Equal(t, 0, len(c.OutputMap()))
	c.Stdout = "stdout"
	_, ok := c.OutputMap()["stdout"]
	assert.True(t, ok)
	_, ok = c.OutputMap()["stderr"]
	assert.False(t, ok)
	c.Stderr = "stderr"
	_, ok = c.OutputMap()["stderr"]
	assert.True(t, ok)
}

func Test_Cons_WhenItemToConsIsNil_ReturnsPreviousList(t *testing.T) {
	c := mkCommandResults(0)
	newList := c.Cons("", nil)
	assert.Equal(t, c, newList)
}

func Test_Cons_PrependsElementToList(t *testing.T) {
	c := mkCommandResults(0)
	expected := mkCommandResults(1)
	assert.Equal(t, uint32(0), c.ExitStatus)
	c = c.Cons("", expected)
	assert.Equal(t, expected.ExitStatus, c.ExitStatus)
}

func Test_Append_WhenToAppendIsNil_DoesNotAppend(t *testing.T) {
	c := mkCommandResults(0)
	c = c.Append("", mkCommandResults(1))
	last := c.Last()
	c = c.Append("", nil)
	assert.Equal(t, last, c.Last())
}

func Test_Eq(t *testing.T) {
	a := newResults("a", 0, "")
	b := newResults("a", 0, "")
	c := newResults("b", 0, "")
	d := newResults("b", 1, "")
	assert.True(t, a.Eq(b))
	assert.True(t, b.Eq(a))
	assert.True(t, a.Eq(a))
	assert.False(t, a.Eq(c))
	assert.False(t, c.Eq(d))
	b.Stdout = "foo"
	assert.False(t, a.Eq(b))
	a.Stdout = "foo"
	b.Stdout = "bar"
	assert.False(t, a.Eq(b))
	assert.False(t, a.Eq(nil))
}

func Test_ExitCodes(t *testing.T) {
	expected := []uint32{0, 1, 2, 3}
	c := newResults("", 0, "")
	c = c.Append("", &shell.CommandResults{ExitStatus: 1})
	c = c.Append("", &shell.CommandResults{ExitStatus: 2})
	c = c.Append("", &shell.CommandResults{ExitStatus: 3})
	exitCodes := c.ExitStatuses()
	assert.Equal(t, expected, exitCodes)
}

func Test_Reverse(t *testing.T) {
	expected := []uint32{0, 1, 2, 3}
	c := newResults("", 0, "")
	c = c.Append("", &shell.CommandResults{ExitStatus: 1})
	c = c.Append("", &shell.CommandResults{ExitStatus: 2})
	c = c.Append("", &shell.CommandResults{ExitStatus: 3})
	exitCodes := c.ExitStatuses()
	assert.Equal(t, expected, exitCodes)
}

func Test_Unlink_RemovesResult(t *testing.T) {
	first := newResults("a", 0, "")
	toRemove := newResults("b", 0, "")
	last := newResults("c", 0, "")
	c := first
	c = c.Append("b", toRemove)
	c = c.Append("c", last)
	removed, c := c.Unlink(toRemove)
	assert.Equal(t, toRemove, removed)
	assert.True(t, c.Eq(first))
	assert.True(t, last.Eq(c.ResultsContext.Next))
	assert.True(t, c.Eq(last.ResultsContext.Prev))
}

func Test_Unlink_WhenFirstElement_ReturnsNextElement(t *testing.T) {
	first := mkCommandResults(0)
	c := first.Append("", mkCommandResults(1))
	_, removed := c.Unlink(first)
	assert.Equal(t, removed.ExitStatus, uint32(1))
}

func Test_UnlinkWhen(t *testing.T) {
	expectedBefore := []uint32{2, 0, 1, 2, 3}
	expectedAfter := []uint32{0, 1, 3}
	c := newResults("a", 2, "")
	c = c.Append("", mkCommandResults(0))
	c = c.Append("", mkCommandResults(1))
	c = c.Append("", mkCommandResults(2))
	c = c.Append("", mkCommandResults(3))
	actualBefore := c.ExitStatuses()
	assert.Equal(t, expectedBefore, actualBefore)
	c = c.UnlinkWhen(func(cmd *shell.CommandResults) bool {
		return cmd.ExitStatus == 2
	})
	actualAfter := c.ExitStatuses()
	assert.Equal(t, expectedAfter, actualAfter)
}

func Test_Uniq(t *testing.T) {
	expectedBefore := []uint32{2, 0, 1, 0}
	expectedAfter := []uint32{2, 0, 1}
	c := mkCommandResults(2)
	c = c.Append("", mkCommandResults(0))
	c = c.Append("a", mkCommandResults(1))
	c = c.Append("", mkCommandResults(0))
	before := c.ExitStatuses()
	assert.Equal(t, expectedBefore, before)
	after := c.Uniq().ExitStatuses()
	assert.Equal(t, expectedAfter, after)
}

func newResults(op string, status uint32, stdout string) *shell.CommandResults {
	return &shell.CommandResults{
		ResultsContext: shell.ResultsContext{Operation: op},
		ExitStatus:     status,
		Stdout:         stdout,
	}
}
func mkCommandResults(status uint32) *shell.CommandResults {
	return &shell.CommandResults{ExitStatus: status}
}
