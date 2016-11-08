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
	"errors"
	"fmt"
	"time"

	"github.com/asteris-llc/converge/load/registry"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/shell"
	"golang.org/x/net/context"
)

// Preparer handles wait.query tasks
type Preparer struct {
	// the shell interpreter that will be used for your scripts. `/bin/sh` is
	// used by default.
	Interpreter string `hcl:"interpreter"`

	// the script to run to check if a resource is ready. exit with exit code 0 if
	// the resource is healthy, and 1 (or above) otherwise.
	Check string `hcl:"check" required:"true"`

	// flags to pass to the `interpreter` binary to check validity. For
	// `/bin/sh` this is `-n`.
	CheckFlags []string `hcl:"check_flags"`

	// flags to pass to the interpreter at execution time.
	ExecFlags []string `hcl:"exec_flags"`

	// the amount of time the command will wait before halting forcefully.
	Timeout *time.Duration `hcl:"timeout"`

	// the working directory this command should be run in.
	Dir string `hcl:"dir"`

	// any environment variables that should be passed to the command.
	Env map[string]string `hcl:"env"`

	// the amount of time to wait in between checks. If the interval is not
	// specified, it will default to 5 seconds.
	Interval *time.Duration `hcl:"interval"`

	// the amount of time to wait before running the first check and after a
	// successful check. If no grace period is specified, no grace period will be
	// taken into account.
	GracePeriod *time.Duration `hcl:"grace_period"`

	// the maximum number of attempts before the wait fails. If the maximum number
	// of retries is not set, it will default to 5.
	MaxRetry *int `hcl:"max_retry"`
}

// Prepare creates a new wait type
func (p *Preparer) Prepare(ctx context.Context, render resource.Renderer) (resource.Task, error) {
	if p.Check == "" {
		return nil, errors.New("Check is required and cannot be empty")
	}

	shPrep := &shell.Preparer{
		Interpreter: p.Interpreter,
		Check:       p.Check,
		CheckFlags:  p.CheckFlags,
		ExecFlags:   p.ExecFlags,
		Timeout:     p.Timeout,
		Dir:         p.Dir,
		Env:         p.Env,
	}

	task, err := shPrep.Prepare(ctx, render)

	if err != nil {
		return &Wait{}, err
	}

	shell, ok := task.(*shell.Shell)
	if !ok {
		return &Wait{}, fmt.Errorf("expected *shell.Shell but got %T", task)
	}

	wait := &Wait{
		Shell:   shell,
		Retrier: PrepareRetrier(p.Interval, p.GracePeriod, p.MaxRetry),
	}

	return wait, nil
}

func init() {
	registry.Register("wait.query", (*Preparer)(nil), (*Wait)(nil))
}
