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
	"errors"
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
	Timedout   bool
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
	if toAppend == nil {
		return c
	}
	toAppend.ResultsContext = ResultsContext{Operation: op, Next: c}
	c.Prev = toAppend
	return toAppend
}

// Unlink removes an element from the results list, and returns a tuple of the
// removed element and the updated list
func (c *CommandResults) Unlink(cmd *CommandResults) (removed *CommandResults, results *CommandResults) {
	results = c
	removed = cmd
	if cmd == nil {
		return
	}
	prev := cmd.ResultsContext.Prev
	next := cmd.ResultsContext.Next
	prev.ResultsContext.Next = next
	next.ResultsContext.Prev = prev
	return
}

// Uniq removes duplicate entries (as dicated by Eq) from the results list
func (c *CommandResults) Uniq() *CommandResults {
	cur := c
	for cur != nil {
		cur.ResultsContext.Next.Foldl(nil, func(cmd *CommandResults, _ interface{}) interface{} {
			if cur.Eq(cmd) {
				cur.Unlink(cmd)
			}
			return nil
		})
		cur = cur.ResultsContext.Next
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
func (c *CommandResults) SummarizeAll() []string {
	var summaries []string
	c.Foldl(nil, func(cmd *CommandResults, _ interface{}) interface{} {
		summaries = append(summaries, cmd.Summarize())
		return nil
	})
	return summaries
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

// ExitStatuses returns a slice with the exit status of all the commands
func (c *CommandResults) ExitStatuses() ([]uint32, error) {
	statuses, ok := c.Foldl([]uint32{}, func(cmd *CommandResults, exitList interface{}) interface{} {
		lst := exitList.([]uint32)
		if cmd == nil {
			return exitList
		}
		return append(lst, cmd.ExitStatus)
	}).([]uint32)
	if !ok {
		return nil, errors.New("cannot get exit statuses (type error)")
	}
	return statuses, nil
}

// ExitStrings returns a list of strings containing the operation type and exit
// status in the form of Operation (code).
func (c *CommandResults) ExitStrings() ([]string, error) {
	stats, ok := c.Foldl([]string{}, func(cmd *CommandResults, exitList interface{}) interface{} {
		if cmd == nil {
			return exitList
		}
		lst := exitList.([]string)
		return append(lst, fmt.Sprintf("%s (%d)", cmd.ResultsContext.Operation, cmd.ExitStatus))
	}).([]string)
	if !ok {
		return nil, errors.New("cannot get exit statuses (type error)")
	}
	return stats, nil
}

// GetMessages will return a set of output messages from all result sets
func (c *CommandResults) GetMessages() []string {
	var output []string
	c.Foldl(output, func(cmd *CommandResults, _ interface{}) interface{} {
		output = append(output, "in "+cmd.ResultsContext.Operation+":")
		stdout := strings.TrimRight(cmd.Stdout, "\r\n\t ")
		stderr := strings.TrimRight(cmd.Stderr, "\r\n\t ")
		if stdout != "" {
			output = append(output, fmt.Sprintf("stdout: %s", stdout))
		}
		if stderr != "" {
			output = append(output, fmt.Sprintf("stderr: %s", stderr))
		}
		return nil
	})
	return output
}

// Foldl will fold a function over the list of CommandResults.
func (c *CommandResults) Foldl(start interface{}, f func(*CommandResults, interface{}) interface{}) interface{} {
	if c == nil {
		return start
	}
	parent := c.ResultsContext.Next
	return f(c, parent.Foldl(start, f))
}
