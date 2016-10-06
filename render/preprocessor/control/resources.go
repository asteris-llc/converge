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

package control

import (
	"fmt"
	"strings"

	"github.com/asteris-llc/converge/load/registry"
	"github.com/asteris-llc/converge/render/extensions"
	"github.com/asteris-llc/converge/resource"
	"github.com/pkg/errors"
)

// SwitchPreparer represents a switch resourc; the task it generates simply
// wraps the values and will not do anything during check or apply.
type SwitchPreparer struct {
	Branches []string `hcl:"branches"`
}

// Prepare does stuff
func (s *SwitchPreparer) Prepare(resource.Renderer) (resource.Task, error) {
	task := &SwitchTask{Branches: s.Branches}

	return task, nil
}

// SwitchTask represents a resource.Task for a switch node.  It does not
// perform any operations and exists to provide structure to conditional
// evaluation in the graph and holds predicate state information.
type SwitchTask struct {
	Branches []string
	cases    []*CaseTask
}

// AppendCase adds a case statement to the list of cases
func (s *SwitchTask) AppendCase(c *CaseTask) {
	s.cases = append(s.cases, c)
	for _, caseObj := range s.cases {
		if caseObj != nil {
			caseObj.SetParent(s)
		}
	}
}

// Cases returns the embedded cases
func (s *SwitchTask) Cases() []*CaseTask {
	if s == nil {
		return []*CaseTask{}
	}
	return s.cases
}

// SortCases ensures that the ordering of the cases slice mirrors the ordering
// of the Branches slice.  Because branches is the canonical order of evaluation
// based on the HCL file and cases may be appended to the list in an unknown
// order due to non-deterministic map evaluation we need to re-sort the list.
func (s *SwitchTask) SortCases() {
	workingMap := map[string]*CaseTask{}
	sorted := []*CaseTask{}
	for _, c := range s.cases {
		workingMap[c.Name] = c
	}
	for _, br := range s.Branches {
		sorted = append(sorted, workingMap[br])
	}
	s.cases = sorted
}

// Check does stuff
func (s *SwitchTask) Check(resource.Renderer) (resource.TaskStatus, error) {
	return &resource.Status{}, nil
}

// Apply does stuff
func (s *SwitchTask) Apply() (resource.TaskStatus, error) {
	return &resource.Status{}, nil
}

func (s *SwitchTask) String() string {
	var out string
	if s == nil {
		return "<nil>"
	}
	out += fmt.Sprintf("Switch: \n")
	out += fmt.Sprintf("\tBranches: \n")
	for _, br := range s.Branches {
		out += fmt.Sprintf("\t\t%s\n", br)
	}
	out += fmt.Sprintf("\tCases: \n")
	for _, c := range s.Cases() {
		out += fmt.Sprintf("\t\t%s\n", helperIndent(c.String(), 1))
	}
	return out
}

// CasePreparer contains generated case macro information
type CasePreparer struct {
	Predicate string `hcl:"predicate"`
	Name      string `hcl:"name"`
}

// Prepare does stuff
func (c *CasePreparer) Prepare(r resource.Renderer) (resource.Task, error) {
	predicate, err := r.Render("predicate", c.Predicate)
	if err != nil {
		return nil, err
	}

	c.Predicate = predicate
	return &CaseTask{
		Predicate: c.Predicate,
		Name:      c.Name,
	}, nil
}

// CaseTask represents a task and is used to determine whether a conditional
// task should evaluate or not
type CaseTask struct {
	Predicate string
	Name      string
	Status    *bool
	parent    *SwitchTask
}

// IsDefault returns true if the case is a default statement
func (c *CaseTask) IsDefault() bool {
	if c == nil {
		return false
	}
	return c.Name == keywords["default"]
}

// SetParent set's the parent of a case statement
func (c *CaseTask) SetParent(s *SwitchTask) {
	c.parent = s
}

// GetParent gets the parent of a case
func (c *CaseTask) GetParent() *SwitchTask {
	return c.parent
}

// ShouldEvaluate returns true if the case has a valid parent and it is the
// selected branch for that parent
func (c *CaseTask) ShouldEvaluate() bool {
	if c.parent == nil {
		return false
	}
	for _, br := range c.parent.Branches {
		if c.Name == br {
			t, _ := c.IsTrue()
			return t
		}
	}
	return false
}

// IsTrue returns true if the template precicate evaluates to "true", or "t",
// false if it returns "false", or "f", or if the pointer is nil, and returns
// false with an error otherwise.
func (c *CaseTask) IsTrue() (bool, error) {
	if c == nil {
		return false, nil
	}
	if c.parent == nil {
		fmt.Println(c.Name, ": parent is nil")
	}
	for _, otherCase := range c.parent.Cases() {
		if otherCase == c {
			break
		}
		if isTrue, _ := otherCase.IsTrue(); isTrue {
			return false, nil
		}
	}
	if c.Status == nil {
		status, err := EvaluatePredicate(c.Predicate)
		if err != nil {
			return false, err
		}
		c.Status = new(bool)
		*c.Status = status
	}
	return *c.Status, nil
}

// EvaluatePredicate looks at a templated string and returns true if template
// execution results in the string "true" or t", and false if the string is
// "false" or "f".  In any other case an error is returned.
func EvaluatePredicate(predicate string) (bool, error) {
	lang := extensions.DefaultLanguage()
	if predicate == "" {
		return false, BadPredicate(predicate)
	}
	template := "{{ " + predicate + " }}"
	result, err := lang.Render(
		struct{}{},
		"predicate evaluation",
		template,
	)
	if err != nil {
		fmt.Println("\treturned an error: ", err)
		return false, errors.Wrap(err, "case evaluation failed")
	}

	truthiness := strings.TrimSpace(strings.ToLower(result.String()))

	switch truthiness {
	case "true", "t":
		return true, nil
	case "false", "f":
		return false, nil
	}
	return false, fmt.Errorf("%s: not a valid truth value; should be one of [f false t true]", truthiness)
}

// Check does stuff
func (c *CaseTask) Check(resource.Renderer) (resource.TaskStatus, error) {
	return &resource.Status{}, nil
}

// Apply does stuff
func (c *CaseTask) Apply() (resource.TaskStatus, error) {
	return &resource.Status{}, nil
}

func (c *CaseTask) String() string {
	var out string
	if nil == c {
		return "<nil>"
	}
	out += fmt.Sprintf("Case: \n")
	out += fmt.Sprintf("\tName: %s\n", c.Name)
	out += fmt.Sprintf("\tPredicate: %s\n", c.Predicate)
	if c.Status == nil {
		out += fmt.Sprintf("\tStatus: unevaluated\n")
	} else {
		out += fmt.Sprintf("\tStatus: %v\n", *(c.Status))
	}
	out += fmt.Sprintf("\tParent: %p\n", c.GetParent())
	return out
}

// ConditionalTask represents a task that may or may not be executed. It's
// evaluation is determined by it's parent control-structure predicate.
type ConditionalTask struct {
	resource.Task
	Name       string
	controller EvaluationController
}

// EvaluationController represents an interface for a thing that can control
// conditional execution (e.g. a CasePreparer or CaseTask)
type EvaluationController interface {
	ShouldEvaluate() bool
}

// SetExecutionController sets the private execution controller
func (c *ConditionalTask) SetExecutionController(ctrl EvaluationController) {
	c.controller = ctrl
}

// GetTask will return the task if it should be evaluated, and a nop-task
// otherwise.  The nop task will embed the original task so fields will still be
// resolvable.
func (c *ConditionalTask) GetTask() (resource.Task, bool) {
	if c.controller.ShouldEvaluate() {
		return c.Task, true
	}
	return &NopTask{c.Task}, true
}

// Apply will conditionally apply a task
func (c *ConditionalTask) Apply() (resource.TaskStatus, error) {
	if c.controller.ShouldEvaluate() {
		return c.Task.Apply()
	}
	return &resource.Status{}, nil
}

// Check will conditionally check a task
func (c *ConditionalTask) Check(r resource.Renderer) (resource.TaskStatus, error) {
	if c == nil {
		return &resource.Status{}, errors.New("conditional task is nil")
	}
	if c.controller.ShouldEvaluate() {
		return c.Task.Check(r)
	}
	return &resource.Status{}, nil
}

// NopTask is a task with accessible fields that will never execute
type NopTask struct {
	resource.Task
}

// Check is a NOP
func (n *NopTask) Check(resource.Renderer) (resource.TaskStatus, error) {
	msg := "Check: pruned branch not executing task"
	return &resource.Status{Output: []string{msg}}, nil
}

// Apply is a NOP
func (n *NopTask) Apply() (resource.TaskStatus, error) {
	msg := "Apply: pruned branch not executing task"
	return &resource.Status{Output: []string{msg}}, nil
}

func init() {
	registry.Register("macro.switch", (*SwitchPreparer)(nil), (*SwitchTask)(nil))
	registry.Register("macro.case", (*CasePreparer)(nil), (*CaseTask)(nil))
}

func helperIndent(s string, count int) string {
	var tabs string
	for idx := 0; idx < count; idx++ {
		tabs += "\t"
	}
	return strings.NewReplacer("\n", "\n"+tabs).Replace(s)
}
