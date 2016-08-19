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

package directory

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/asteris-llc/converge/helpers"
	"github.com/asteris-llc/converge/resource/file/mode"
	"github.com/asteris-llc/converge/resource/file/owner"
)

// Content renders a content to disk
type Directory struct {
	Destination string
	Recurse     bool

	Mode  *mode.Mode
	Owner *owner.Owner
}

// Check if the content needs to be rendered
func (d *Directory) Check() (string, bool, error) {
	if !d.Recurse {
		return d.check()
	}
	rootStat, err := os.Stat(d.Destination)
	if os.IsNotExist(err) {
		return fmt.Sprintf("%q does not exist. will be created", d.Destination), false, nil
	}
	if !rootStat.IsDir() {
		return fmt.Sprintf("'recurse' paramter used on file %q", d.Destination), false, nil
	}
	type check struct {
		status     string
		willChange bool
		err        error
	}
	checks := []check{}
	err = filepath.Walk(d.Destination, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if d.Mode != nil {

			modeTask := &mode.Mode{
				Destination: path,
				Mode:        d.Mode.Mode,
			}
			s, w, e := modeTask.Check()
			checked := check{s, w, e}
			checks = append(checks, checked)
		}
		if d.Owner != nil {

			ownerTask := &owner.Owner{
				Destination: path,
				Username:    d.Owner.Username,
				UID:         d.Owner.UID,
				GID:         d.Owner.GID,
			}
			s, w, e := ownerTask.Check()
			checked := check{s, w, e}
			checks = append(checks, checked)
		}
		return nil
	})

	if err != nil {
		checks = append(checks, check{"", false, err})
	}
	status, willChange, err := checks[0].status, checks[0].willChange, checks[0].err
	for i := 1; i < len(checks); i++ {
		status, willChange, err = helpers.SquashCheck(status, willChange, err, checks[1].status, checks[1].willChange, checks[1].err)
	}
	return status, willChange, err
}

func (d *Directory) check() (string, bool, error) {
	stat, err := os.Stat(d.Destination)
	if os.IsNotExist(err) {
		return fmt.Sprintf("%q does not exist. will be created", d.Destination), true, nil
	} else if err != nil {
		return "", false, err
	}
	if !stat.IsDir() {
		return fmt.Sprintf("file named %q already exist", d.Destination), false, nil
	} else {
		s1, c1, e1 := fmt.Sprintf("folder named %q already exist. ", d.Destination), false, error(nil)
		s2, c2, e2 := d.checkDir()
		return helpers.SquashCheck(s1, c1, e1, s2, c2, e2)
	}
}

// Apply writes the content to disk
func (d *Directory) Apply() error {
	err := d.apply()
	if !d.Recurse {
		return err
	}
	rootStat, err := os.Stat(d.Destination)
	if os.IsNotExist(err) {
		return fmt.Errorf("%q does not exist. cannot recurse", d.Destination)
	}
	if !rootStat.IsDir() {
		return fmt.Errorf("'recurse' paramter used on file %q", d.Destination)
	}
	err = helpers.MultiErrorAppend(err, filepath.Walk(d.Destination, func(path string, info os.FileInfo, walkerr error) error {
		if walkerr != nil {
			return walkerr
		}
		if d.Mode != nil {
			modeTask := &mode.Mode{
				Destination: path,
				Mode:        d.Mode.Mode,
			}
			err = helpers.MultiErrorAppend(err, modeTask.Apply())
		}
		if d.Owner != nil {
			ownerTask := &owner.Owner{
				Destination: path,
				Username:    d.Owner.Username,
				UID:         d.Owner.UID,
				GID:         d.Owner.GID,
			}
			err = helpers.MultiErrorAppend(err, ownerTask.Apply())
		}
		return nil
	}))

	return err
}

// Apply writes the content to disk
//TODO change the owner of all created directories.
func (d *Directory) apply() (err error) {
	_, err = os.Stat(d.Destination)
	var fileMode os.FileMode = 0777
	if d.Mode != nil {
		fileMode = d.Mode.Mode
	}
	//Doesn't exist -> create
	if os.IsNotExist(err) {
		return os.MkdirAll(d.Destination, fileMode)
	}
	return d.applyDir()
	return err

}

func (d *Directory) applyDir() (err error) {
	return helpers.MultiErrorAppend(d.applyMode(), d.applyOwner())
}

func (d *Directory) checkDir() (status string, willChange bool, err error) {
	s1, c1, e1 := d.checkMode()
	s2, c2, e2 := d.checkOwner()
	return helpers.SquashCheck(s1, c1, e1, s2, c2, e2)
}

func (d *Directory) checkMode() (status string, willChange bool, err error) {
	if d.Mode != nil {
		return d.Mode.Check()
	} else {
		return "", false, nil
	}
}

func (d *Directory) applyMode() (err error) {
	if d.Mode != nil {
		return d.Mode.Apply()
	} else {
		return nil
	}
}

func (d *Directory) checkOwner() (status string, willChange bool, err error) {
	if d.Owner != nil {
		return d.Owner.Check()
	} else {
		return "", false, nil
	}
}

func (d *Directory) applyOwner() (err error) {
	if d.Owner != nil {
		return d.Owner.Apply()
	} else {
		return nil
	}
}
