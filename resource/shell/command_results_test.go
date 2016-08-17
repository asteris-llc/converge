package shell_test

import (
	"fmt"
	"testing"

	"github.com/asteris-llc/converge/resource/shell"
	"github.com/stretchr/testify/assert"
)

func Test_Eq(t *testing.T) {
	a := newResults("a", 0, "")
	b := newResults("a", 0, "")
	c := newResults("b", 0, "")
	assert.True(t, a.Eq(b))
	assert.True(t, b.Eq(a))
	assert.True(t, a.Eq(a))
	assert.False(t, a.Eq(c))
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
	fmt.Println("=============================")
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
