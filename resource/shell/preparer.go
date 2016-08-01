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
	"log"
	"os/exec"

	"github.com/asteris-llc/converge/resource"
	"github.com/pkg/errors"
)

// Preparer for Shell tasks
type Preparer struct {
	Interpreter   string  `hcl:"interpreter"`
	SyntaxChecker *string `hcl:"syntaxchecker"`
	Check         string  `hcl:"check"`
	Apply         string  `hcl:"apply"`
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

	interpreter, err := render.Render("interpreter", p.Interpreter)
	if err != nil {
		return nil, err
	}

	if interpreter == "" {
		interpreter = "sh" // TODO: make this work on Windows?
	}

	return &Shell{interpreter, check, apply}, nil
}

func (p *Preparer) validateScriptSyntax(script string) error {

	command, cmdError := p.getValidatorCommand()

	if cmdError != nil {
		log.Printf("[WARN] script validator returned '%s', will not attempt to validate script syntax\n", cmdError)
		return nil
	}

	in, err := command.StdinPipe()
	if err != nil {
		return errors.Wrap(err, "unable to create pipe")
	}

	if err := command.Start(); err != nil {
		return errors.Wrap(err, "failed to start interpreter")
	}

	if _, err := in.Write([]byte(script)); err != nil {
		return errors.Wrap(err, "unable to write to interpreter")
	}

	if err := in.Close(); err != nil {
		return errors.Wrap(err, "failed to close pipe")
	}

	if err := command.Wait(); err != nil {
		return errors.Wrap(err, "syntax error")
	}

	return nil
}

func (p *Preparer) getValidatorCommand() (*exec.Cmd, error) {
	interpreter := p.Interpreter
	checkFlag := p.SyntaxChecker
	if interpreter != "" && checkFlag != nil {
		return exec.Command(interpreter, *checkFlag), nil
	}
	if interpreter == "" && checkFlag == nil {
		return exec.Command("sh", "-n"), nil
	}
	if interpreter == "" {
		return nil, errors.New("syntaxcheck was set without a user-specified interpreter")
	}
	return nil, errors.New("custom interpreter defined without syntaxcheck flag")
}
