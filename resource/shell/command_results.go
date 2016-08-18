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

package shell

import (
	"fmt"
	"os"
	"strings"
)

// CommandResults hold the resulting state of command execution
type CommandResults struct {
	ResultsContext
	ExitStatus uint32
	Stdout     string
	Stderr     string
	Stdin      string
	State      *os.ProcessState
}

// ResultsContext provides a linked list of CommandResults with a operation
// context that tells us the providence of the command.
type ResultsContext struct {
	Operation string
	Next      *CommandResults
	Prev      *CommandResults
}

// OutputMap returns a map of fds by name and the output on them
func (c *CommandResults) OutputMap() map[string]string {
	m := make(map[string]string)
	if c == nil {
		return m
	}
	if c.Stdout != "" {
		m["stdout"] = c.Stdout
	}
	if c.Stderr != "" {
		m["stderr"] = c.Stderr
	}
	return m
}

// Cons a command result to another command result, allowing the capture of
// multiple runs of a task with a session.
func (c *CommandResults) Cons(op string, toAppend *CommandResults) *CommandResults {
	if c == nil {
		toAppend.Operation = op
		return toAppend
	}
	if toAppend == nil {
		return c
	}
	fmt.Printf("consing with op = %s\n", op)
	toAppend.ResultsContext = ResultsContext{Operation: op, Next: c}
	c.Prev = toAppend
	return toAppend
}

// Append adds an element to the end of the list
func (c *CommandResults) Append(op string, toAppend *CommandResults) *CommandResults {
	if toAppend == nil {
		return c
	}
	last := c.Last()
	toAppend.ResultsContext.Prev = last
	toAppend.ResultsContext.Operation = op
	last.Next = toAppend
	return c
}

// Unlink removes an element from the results list, and returns a tuple of the
// removed element and the updated list.  If the Element to remove is the head
// of the list, then it will return the next element, otherwise return the
// current head.
func (c *CommandResults) Unlink(cmd *CommandResults) (removed *CommandResults, results *CommandResults) {
	results = c
	removed = cmd
	if cmd == nil {
		return
	}
	if cmd == c {
		results = c.ResultsContext.Next
	}
	prev := cmd.ResultsContext.Prev
	next := cmd.ResultsContext.Next
	if prev != nil {
		prev.ResultsContext.Next = next
	}
	if next != nil {
		next.ResultsContext.Prev = prev
	}
	return
}

// UnlinkWhen removes each element in the list when the provided function is true
func (c *CommandResults) UnlinkWhen(f func(*CommandResults) bool) *CommandResults {
	if c == nil {
		return c
	}
	cur := c
	for cur != nil {
		if f(cur) {
			_, c = c.Unlink(cur)
		}
		cur = cur.ResultsContext.Next
	}
	return c.First()
}

// Uniq removes duplicate entries (as dicated by Eq) from the results list
func (c *CommandResults) Uniq() *CommandResults {
	for cur := c; cur.ResultsContext.Next != nil; cur = cur.ResultsContext.Next {
		c = cur.ResultsContext.Next.UnlinkWhen(cur.Eq)
	}
	return c
}

// UniqOp removes duplicate operation entries based just on their operation name
func (c *CommandResults) UniqOp() *CommandResults {
	for cur := c; cur.ResultsContext.Next != nil; cur = cur.ResultsContext.Next {
		c = cur.ResultsContext.Next.UnlinkWhen(c.OpEq)
	}
	return c
}

// Summarize provides an overall summary of the results of the command
func (c *CommandResults) Summarize() string {
	if c == nil {
		return ""
	}
	return fmt.Sprintf(
		"%s (returned: %d)\n\t%s\n\t%s",
		c.ResultsContext.Operation,
		c.ExitStatus,
		c.Stdout,
		c.Stderr,
	)
}

// SummarizeAll returnes a list of summaries of a command result and it's
// ancestors
func (c *CommandResults) SummarizeAll() (summaries []string) {
	for c != nil {
		summaries = append(summaries, c.Summarize())
		c = c.ResultsContext.Next
	}
	return
}

// Eq tests command results for equality
func (c *CommandResults) Eq(cmd *CommandResults) bool {
	if c == nil || cmd == nil {
		return false
	}
	if c.ExitStatus != cmd.ExitStatus {
		return false
	}
	if c.Stdout != cmd.Stdout {
		return false
	}
	if c.Stderr != cmd.Stderr {
		return false
	}
	if c.ResultsContext.Operation != cmd.ResultsContext.Operation {
		return false
	}
	return true
}

// OpEq tests command results for equality using only the operation name
func (c *CommandResults) OpEq(cmd *CommandResults) bool {
	if c == nil || cmd == nil {
		return false
	}
	return c.ResultsContext.Operation == cmd.ResultsContext.Operation
}

// ExitStatuses returns a slice with the exit status of all the commands
func (c *CommandResults) ExitStatuses() (results []uint32) {
	for c != nil {
		results = append(results, c.ExitStatus)
		c = c.ResultsContext.Next
	}
	return
}

// ExitStrings returns a list of strings containing the operation type and exit
// status in the form of Operation (code).
func (c *CommandResults) ExitStrings() (results []string) {
	for c != nil {
		results = append(results, fmt.Sprintf("%s (%d)", c.ResultsContext.Operation, c.ExitStatus))
		c = c.ResultsContext.Next
	}
	return
}

// GetMessages will return a set of output messages from all result sets
func (c *CommandResults) GetMessages() (output []string) {
	for c != nil {
		output = append(output, "in "+c.ResultsContext.Operation+":")
		stdout := strings.TrimRight(c.Stdout, "\r\n\t ")
		stderr := strings.TrimRight(c.Stderr, "\r\n\t ")
		if stdout != "" {
			output = append(output, fmt.Sprintf("stdout: %s", stdout))
		}
		if stderr != "" {
			output = append(output, fmt.Sprintf("stderr: %s", stderr))
		}
		c = c.ResultsContext.Next
	}
	return
}

// Last returns the last element in the list
func (c *CommandResults) Last() *CommandResults {
	if c == nil {
		return nil
	}
	last := c
	for last.ResultsContext.Next != nil {
		last = last.ResultsContext.Next
	}
	return last
}

// First returns the head of the results list
func (c *CommandResults) First() *CommandResults {
	if c == nil {
		return nil
	}
	first := c
	for first.ResultsContext.Prev != nil {
		first = first.ResultsContext.Prev
	}
	return first
}

// reverseNode reverses the direction of a single node
func (c *CommandResults) reverseNode() *CommandResults {
	if c == nil {
		return c
	}
	prev := c.ResultsContext.Prev
	next := c.ResultsContext.Next
	c.ResultsContext.Prev = next
	c.ResultsContext.Next = prev
	return c
}

// Reverse will reverse the list
func (c *CommandResults) Reverse() *CommandResults {
	last := c.Last()
	for cur := last; cur != nil; cur = cur.ResultsContext.Next {
		cur = cur.reverseNode()
	}
	return last
}
