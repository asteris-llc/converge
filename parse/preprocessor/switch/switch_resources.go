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

	"github.com/asteris-llc/converge/resource"
	"golang.org/x/net/context"
)

// SwitchPreparer represents a switch resource; the task it generates simply
// wraps the values and will not do anything during check or apply.
type SwitchPreparer struct {
	Branches []string `hcl:"branches"`
}

// Prepare does stuff
func (s *SwitchPreparer) Prepare(context.Context, resource.Renderer) (resource.Task, error) {
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
func (s *SwitchTask) Check(context.Context, resource.Renderer) (resource.TaskStatus, error) {
	return &resource.Status{}, nil
}

// Apply does stuff
func (s *SwitchTask) Apply(context.Context) (resource.TaskStatus, error) {
	return &resource.Status{}, nil
}

// String displays a switch statement; useful for debugging and errors
func (s *SwitchTask) String() string {
	if s == nil {
		return "<nil>"
	}
	fmtStr := `Switch:
	Branches:
		%s
	Cases:
		%s
`
	var branchStrs string
	for _, br := range s.Branches {
		branchStrs += fmt.Sprintf("\t\t%s\n", br)
	}
	var caseStrs string
	for _, c := range s.Cases() {
		caseStrs += fmt.Sprintf("\t\t%s\n", helperIndent(c.String(), 1))
	}
	return fmt.Sprintf(fmtStr, branchStrs, caseStrs)
}

func helperIndent(s string, count int) string {
	var tabs string
	for idx := 0; idx < count; idx++ {
		tabs += "\t"
	}
	return strings.NewReplacer("\n", "\n"+tabs).Replace(s)
}
