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
	"os/exec"

	"github.com/asteris-llc/converge/resource"
)

// Preparer for Shell tasks
type Preparer struct {
	Check string `hcl:"check"`
	Apply string `hcl:"apply"`
}

// Prepare a new task
func (p *Preparer) Prepare(render resource.Renderer) (resource.Task, error) {
	check, err := render.Render("check", p.Check)
	if err != nil {
		return nil, err
	}

	if err := p.validateScriptSyntax(check); err != nil {
		return nil, err
	}

	apply, err := render.Render("apply", p.Apply)
	if err != nil {
		return nil, err
	}

	if err := p.validateScriptSyntax(apply); err != nil {
		return nil, err
	}

	return &Shell{check, apply}, nil
}

func (p *Preparer) validateScriptSyntax(script string) error {
	command := exec.Command("sh", "-n")

	in, err := command.StdinPipe()
	if err != nil {
		return err
	}

	if err := command.Start(); err != nil {
		return err
	}

	if _, err := in.Write([]byte(script)); err != nil {
		return err
	}

	if err := in.Close(); err != nil {
		return err
	}

	if err := command.Wait(); err != nil {
		return err
	}

	return nil
}
