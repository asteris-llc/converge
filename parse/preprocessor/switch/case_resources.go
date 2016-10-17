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

	"github.com/asteris-llc/converge/render"
	"github.com/asteris-llc/converge/render/extensions"
	"github.com/asteris-llc/converge/resource"
	"github.com/pkg/errors"
)

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

	return &CaseTask{
		Predicate: predicate,
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
		return false, errors.New("case is nil")
	}
	if c.parent == nil {
		return false, errors.New("parent is nil")
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

// GetCase provides a default implementation of SwitchBranch
func (c *CaseTask) GetCase() (*CaseTask, error) {
	return c, nil
}

// ThunkedCaseTask provides a thunked wrapper around the case task, allowing
// deferred execution case nodes to be included in switch statements and as
// evaluation controllers.
type ThunkedCaseTask struct {
	*render.PrepareThunk
	caseTask *CaseTask
}

// ShouldEvaluate provides an implementation of EvaluationController for a
// thunked case task, it requires that the inner value has been thunked before
// it can do anything.
func (t *ThunkedCaseTask) ShouldEvaluate() bool {
	if t == nil {
		return false
	}
	if t.caseTask == nil {
		return false
	}
	return t.caseTask.ShouldEvaluate()
}

// GetCase for a thunked task returns true if the thunk has been evaluated,
// otherwise it returns false.
func (t *ThunkedCaseTask) GetCase() (*CaseTask, error) {
	if t.caseTask != nil {
		return t.caseTask, nil
	}

	return nil, errors.New("unable to resolve thunked case")
}

// ApplyThunk applys the inner thunk then sets the inner caseTask value
func (t *ThunkedCaseTask) ApplyThunk(f *render.Factory) (resource.Task, error) {
	result, err := t.PrepareThunk.ApplyThunk(f)
	if asCase, ok := result.(*CaseTask); ok {
		t.caseTask = asCase
	}
	return result, err
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
	if c == nil {
		return "<nil>"
	}
	statusStr := "unevaluated"
	if c.Status != nil {
		statusStr = "evaluated"
	}
	fmtStr := `Case:
	Name: %s
	Predicate: %s
	Status: %s
	Parent: %p
`
	return fmt.Sprintf(fmtStr, c.Name, c.Predicate, statusStr, c.GetParent())
}
