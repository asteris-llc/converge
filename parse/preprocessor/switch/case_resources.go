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

	"github.com/asteris-llc/converge/resource"
	"golang.org/x/net/context"
)

// CasePreparer contains generated case macro information
type CasePreparer struct {
	Name string `hcl:"name"`
}

// Prepare does stuff
func (c *CasePreparer) Prepare(ctx context.Context, r resource.Renderer) (resource.Task, error) {
	return &CaseTask{
		Name: c.Name,
	}, nil
}

// CaseTask represents a task and is used to determine whether a conditional
// task should evaluate or not
type CaseTask struct {
	Name   string
	parent *SwitchTask
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

// Check does stuff
func (c *CaseTask) Check(context.Context, resource.Renderer) (resource.TaskStatus, error) {
	return &resource.Status{}, nil
}

// Apply does stuff
func (c *CaseTask) Apply(context.Context) (resource.TaskStatus, error) {
	return &resource.Status{}, nil
}

func (c *CaseTask) String() string {
	if c == nil {
		return "<nil>"
	}
	fmtStr := `Case:
	Name: %s
	Parent: %p
`
	return fmt.Sprintf(fmtStr, c.Name, c.GetParent())
}
