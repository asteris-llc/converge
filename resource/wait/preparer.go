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

package wait

import (
	"fmt"
	"time"

	"github.com/asteris-llc/converge/load/registry"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/shell"
)

// Preparer handles wait.query tasks
type Preparer struct {
	// the shell interpreter that will be used for your scripts. `/bin/sh` is
	// used by default.
	Interpreter string `hcl:"interpreter"`

	// the script to run to check if a resource is ready. exit with exit code 0 if
	// the resource is healthy, and 1 (or above) otherwise.
	Check string `hcl:"check"`

	// flags to pass to the `interpreter` binary to check validity. For
	// `/bin/sh` this is `-n`.
	CheckFlags []string `hcl:"check_flags"`

	// flags to pass to the interpreter at execution time.
	ExecFlags []string `hcl:"exec_flags"`

	// the amount of time the command will wait before halting forcefully. The
	// format is Go's duraction string. A duration string is a possibly signed
	// sequence of decimal numbers, each with optional fraction and a unit
	// suffix, such as "300ms", "-1.5h" or "2h45m". Valid time units are "ns",
	// "us" (or "µs"), "ms", "s", "m", "h".
	Timeout string `hcl:"timeout" doc_type:"duration string"`

	// the working directory this command should be run in.
	Dir string `hcl:"dir"`

	// any environment variables that should be passed to the command.
	Env map[string]string `hcl:"env"`

	// the amount of time to wait in between checks. The format is Go's duraction
	// string. A duration string is a possibly signed sequence of decimal numbers,
	// each with optional fraction and a unit suffix, such as "300ms", "-1.5h" or
	// "2h45m". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
	Interval string `hcl:"interval" doc_type:"duration string"`

	// the amount of time to wait before running the first check. The format is
	// Go's duraction string. A duration string is a possibly signed sequence of
	// decimal numbers, each with optional fraction and a unit suffix, such as
	// "300ms", "-1.5h" or "2h45m". Valid time units are "ns", "us" (or "µs"),
	// "ms", "s", "m", "h".
	GracePeriod string `hcl:"grace_period" doc_type:"duration string"`

	// the maximum number of attempts before the wait fails.
	MaxRetry interface{} `hcl:"max_retry"`
}

// Prepare creates a new wait type
func (p *Preparer) Prepare(render resource.Renderer) (resource.Task, error) {
	shPrep := &shell.Preparer{
		Interpreter: p.Interpreter,
		Check:       p.Check,
		CheckFlags:  p.CheckFlags,
		ExecFlags:   p.ExecFlags,
		Timeout:     p.Timeout,
		Dir:         p.Dir,
		Env:         p.Env,
	}

	task, err := shPrep.Prepare(render)

	if err != nil {
		return &Wait{}, err
	}

	shell, ok := task.(*shell.Shell)
	if !ok {
		return &Wait{}, fmt.Errorf("expected *shell.Shell but got %T", task)
	}

	wait := &Wait{Shell: shell, Retrier: &Retrier{}}

	interval, err := render.Render("interval", p.Interval)
	if err != nil {
		return wait, err
	}

	if intervalDuration, perr := time.ParseDuration(interval); perr == nil {
		wait.Interval = intervalDuration
	}

	gracePeriod, err := render.Render("grace_period", p.GracePeriod)
	if err != nil {
		return wait, err
	}

	if gracePeriodDuration, perr := time.ParseDuration(gracePeriod); perr == nil {
		wait.GracePeriod = gracePeriodDuration
	}

	if maxRetry, ok := p.MaxRetry.(int); ok {
		wait.MaxRetry = maxRetry
	}

	return wait, nil
}

func init() {
	registry.Register("wait.query", (*Preparer)(nil), (*Wait)(nil))
}
