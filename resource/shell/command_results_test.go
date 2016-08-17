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
	fmt.Println("in test exit codes")
	expected := []uint32{0, 1, 2, 3}
	c := newResults("", 0, "")
	c = c.Cons("", &shell.CommandResults{ExitStatus: 1})
	c = c.Cons("", &shell.CommandResults{ExitStatus: 2})
	c = c.Cons("", &shell.CommandResults{ExitStatus: 3})
	exitCodes, _ := c.ExitStatuses()
	assert.Equal(t, expected, exitCodes)
}

func Test_Reverse(t *testing.T) {
	fmt.Println("In test reverse")
	expected := []uint32{0, 1, 2, 3}
	c := newResults("", 0, "")
	c = c.Append("", &shell.CommandResults{ExitStatus: 1})
	c = c.Append("", &shell.CommandResults{ExitStatus: 2})
	c = c.Append("", &shell.CommandResults{ExitStatus: 3})
	c = c.Reverse()
	exitCodes, _ := c.ExitStatuses()
	assert.Equal(t, expected, exitCodes)
}

func Test_Unlink_RemovesResult(t *testing.T) {
	fmt.Println("in Test_UnlinkRemovesResults")
	first := newResults("a", 0, "")
	toRemove := newResults("b", 0, "")
	last := newResults("c", 0, "")
	c := first
	c = c.Append("b", toRemove)
	c = c.Append("c", last)
	removed, c := c.Unlink(toRemove)
	assert.Equal(t, toRemove, removed)
	fmt.Println(c.Summarize())
	fmt.Println(c.ResultsContext.Next.Summarize())
	assert.True(t, c.Eq(first))
	assert.True(t, last.Eq(c.ResultsContext.Next))
	assert.True(t, c.Eq(last.ResultsContext.Prev))
}

func Test_Uniq(t *testing.T) {
	expectedBefore := []uint32{0, 1, 0, 2}
	expectedAfter := []uint32{0, 1, 2}
	c := newResults("a", 0, "")
	c = c.Append("", mkCommandResults(1))
	c = c.Append("", mkCommandResults(0))
	c = c.Append("", mkCommandResults(2))
	before, _ := c.ExitStatuses()
	assert.Equal(t, expectedBefore, before)
	after, _ := c.Uniq().ExitStatuses()
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
