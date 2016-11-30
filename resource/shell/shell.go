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
	"strings"

	"github.com/asteris-llc/converge/resource"
	"golang.org/x/net/context"
)

// Shell is a structure representing a task.
type Shell struct {
	CmdGenerator   CommandExecutor
	CheckStmt      string                 `export:"check"`
	ApplyStmt      string                 `export:"apply"`
	Dir            string                 `export:"dir"`
	Env            []string               `export:"env"`
	Status         *CommandResults        `re-export-as:"status"`
	CheckStatus    *CommandResults        `export:"checkstatus"`
	HealthStatus   *resource.HealthStatus `export:"healthstatus"`
	renderer       resource.Renderer
	ctx            context.Context
	exportedFields resource.FieldMap
}

// Check passes through to shell.Shell.Check() and then sets the health status
func (s *Shell) Check(ctx context.Context, r resource.Renderer) (resource.TaskStatus, error) {
	s.renderer = r
	results, err := s.CmdGenerator.Run(s.CheckStmt)
	if err != nil {
		return nil, err
	}
	if s.Status == nil {
		s.Status = s.Status.Cons("check", results)
	}
	if s.CheckStatus == nil {
		s.CheckStatus = results
	}
	return s, nil
}

// ExportedFields returns the exported field map
func (s *Shell) ExportedFields() resource.FieldMap {
	if s.exportedFields == nil {
		s.exportedFields = make(resource.FieldMap)
	}
	return s.exportedFields
}

// UpdateExportedFields is a nop
func (s *Shell) UpdateExportedFields(resource.Task) error {
	fields, err := resource.LookupMapFromStruct(s)
	if err != nil {
		return err
	}
	s.exportedFields = fields
	return nil
}

// Apply is a NOP for health checks
func (s *Shell) Apply(context.Context) (resource.TaskStatus, error) {
	if cg, ok := s.CmdGenerator.(*CommandGenerator); ok {
		s.CmdGenerator = cg
	}
	results, err := s.CmdGenerator.Run(s.ApplyStmt)
	if err == nil {
		s.Status = s.Status.Cons("apply", results)
	}
	return s, err
}

// resource.TaskStatus functions

// Value provides a value for the shell, which is the stdout data from the last
// executed command.
func (s *Shell) Value() string {
	return s.Status.Stdout
}

// Diffs is required to implement resource.TaskStatus but there is no mechanism
// for defining diffs for shell operations, so returns a nil map.
func (s *Shell) Diffs() map[string]resource.Diff {
	return nil
}

// StatusCode returns the status code of the most recently executed command
func (s *Shell) StatusCode() resource.StatusLevel {
	if s.Status == nil {
		return resource.StatusFatal
	}

	if s.Status.ExitStatus == 0 {
		return resource.StatusNoChange
	}

	return resource.StatusWillChange
}

// Messages returns a summary of the first execution of check and/or apply.
// Subsequent runs are surpressed.
func (s *Shell) Messages() (messages []string) {
	if s.Status == nil {
		return
	}

	if s.Dir != "" {
		messages = append(messages, fmt.Sprintf("dir (%s)", s.Dir))
	}

	if len(s.Env) > 0 {
		messages = append(messages, fmt.Sprintf("env (%s)", strings.Join(s.Env, " ")))
	}

	messages = append(messages, s.Status.Reverse().UniqOp().SummarizeAll()...)
	return
}

// HasChanges returns true if changes are required as determined by the the most
// recent run of check.
func (s *Shell) HasChanges() bool {
	if s.Status == nil {
		return false
	}
	return (s.Status.ExitStatus != 0)
}

// healthcheck.Check functions

// FailingDep tracks a failing dependency
func (s *Shell) FailingDep(name string, task resource.TaskStatus) {
	if s.HealthStatus == nil {
		s.HealthStatus = new(resource.HealthStatus)
		s.HealthStatus.FailingDeps = make(map[string]string)
	}
	s.HealthStatus.FailingDeps[name] = name
}

// HealthCheck performs a health check
func (s *Shell) HealthCheck() (*resource.HealthStatus, error) {
	var err error
	if s.HealthStatus == nil {
		err = s.updateHealthStatus()
	}
	return s.HealthStatus, err
}

// Error is required for TaskStatus
func (s *Shell) Error() error {
	if s.HealthStatus != nil {
		return s.HealthStatus.Error()
	}

	return nil
}

// Warning is required for TaskStatus
func (s *Shell) Warning() string {
	return ""
}

func (s *Shell) updateHealthStatus() error {
	if s.Status == nil {
		fmt.Println("[INFO] health status requested with no plan, running check")
		if _, err := s.Check(s.ctx, s.renderer); err != nil {
			return err
		}
	}
	if s.HealthStatus == nil {
		s.HealthStatus = new(resource.HealthStatus)
	}
	s.HealthStatus.TaskStatus = s
	s.HealthStatus.WarningLevel = exitStatusToWarningLevel(s.Status.ExitStatus)
	return nil
}

func exitStatusToWarningLevel(status uint32) resource.HealthStatusCode {
	if status == 0 {
		return resource.StatusHealthy
	} else if status == 1 {
		return resource.StatusWarning
	}
	return resource.StatusError
}
