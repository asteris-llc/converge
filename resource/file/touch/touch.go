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

package touch

import (
	"fmt"
	"os"

	"github.com/asteris-llc/converge/helpers"
	"github.com/asteris-llc/converge/resource/file/mode"
	"github.com/asteris-llc/converge/resource/file/owner"
)

// Content renders a content to disk
type Touch struct {
	Destination string

	Mode  *mode.Mode
	Owner *owner.Owner
}

// Check if the content needs to be rendered
func (t *Touch) Check() (status string, willChange bool, err error) {
	_, err = os.Stat(t.Destination)
	if os.IsNotExist(err) {
		return fmt.Sprintf("%q does not exist. will be created", t.Destination), true, nil
	} else if err == nil {
		s1, c1, w1 := fmt.Sprintf("%q already exist.", t.Destination), false, error(nil)
		s2, c2, w2 := t.checkFile()
		return helpers.SquashCheck(s1, c1, w1, s2, c2, w2)
	} else {
		return "", false, err
	}
}

func (t *Touch) Apply() (err error) {
	_, err = os.Stat(t.Destination)
	if os.IsNotExist(err) {
		_, err = os.Create(t.Destination)
	}
	if err != nil {
		return err
	}
	t.applyFile()

	return err
}

func (t *Touch) checkFile() (status string, willChange bool, err error) {
	s1, c1, e1 := t.checkMode()
	s2, c2, e2 := t.checkOwner()
	return helpers.SquashCheck(s1, c1, e1, s2, c2, e2)
}

func (t *Touch) applyFile() (err error) {
	return helpers.MultiErrorAppend(t.applyMode(), t.applyOwner())
}

func (t *Touch) checkMode() (status string, willChange bool, err error) {
	if t.Mode != nil {
		return t.Mode.Check()
	} else {
		return "", false, nil
	}
}

func (t *Touch) applyMode() (err error) {
	if t.Mode != nil {
		return t.Mode.Apply()
	} else {
		return nil
	}
}

func (t *Touch) checkOwner() (status string, willChange bool, err error) {
	if t.Owner != nil {
		return t.Owner.Check()
	} else {
		return "", false, nil
	}
}

func (t *Touch) applyOwner() (err error) {
	if t.Owner != nil {
		return t.Owner.Apply()
	} else {
		return nil
	}
}
