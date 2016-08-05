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

package file

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/asteris-llc/converge/resource/file/helpers"
	"github.com/asteris-llc/converge/resource/file/mode"
	"github.com/asteris-llc/converge/resource/file/owner"
	"github.com/hashicorp/go-multierror"
)

type FileState string

const (
	FSFile      FileState = "file"
	FSLink      FileState = "link" //Soft Link
	FSDirectory FileState = "directory"
	FSHard      FileState = "hard" //Hard Link
	FSTouch     FileState = "touch"
	FSAbsent    FileState = "absent"
)

// File wraps basic file properties.
type File struct {
	//file properties
	State       FileState
	Source      string
	Destination string
	Recurse     bool

	*mode.Mode
	*owner.Owner
}

// Check decideds what operations to perform
func (f *File) Check() (status string, willChange bool, err error) {
	if !f.Recurse {
		switch f.State {
		case FSAbsent:
			return f.checkFSAbsent()
		case FSTouch:
			return f.checkFSTouch()
		case FSLink:
			return f.checkFSLink()
		case FSHard:
			return f.checkFSLink()
		case FSDirectory:
			return f.checkFSDirectory()
		case FSFile:
			fallthrough
		case "":
			return f.checkFSFile()
		default:
			return "", false, errors.New("invalid value for 'state' parameter")
		}
	}
	if f.State != FSDirectory {
		return "", false, fmt.Errorf("cannot use the 'recurse' parameter when 'state' is not set to %s", FSDirectory)
	}
	type check struct {
		status     string
		willChange bool
		err        error
	}
	rootStat, err := os.Stat(f.Destination)
	if os.IsNotExist(err) {
		return fmt.Sprintf("%q does not exist. cannot recurse", f.Destination), false, nil
	}
	if !rootStat.IsDir() {
		return fmt.Sprintf("'recurse' paramter used on file %q", f.Destination), false, nil
	}
	checks := []check{}
	err = filepath.Walk(f.Destination, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		checks = append(checks, check{status: "", willChange: false, err: err})
		filemodule := File{
			Destination: path,
		}
		if f.Mode != nil {
			filemodule.Mode = &mode.Mode{
				Destination: path,
				Mode:        f.Mode.Mode,
			}
		}
		if f.Owner != nil {
			filemodule.Owner = &owner.Owner{
				Destination: path,
				Username:    f.Owner.Username,
				UID:         f.Owner.UID,
				GID:         f.Owner.GID,
			}
		}
		status, willChange, err := filemodule.checkFSFile()
		checked := check{status: status, willChange: willChange, err: err}
		checks = append(checks, checked)
		return nil
	})

	if err != nil {
		checks = append(checks, check{"", false, err})
	}
	status, willChange, err = checks[0].status, checks[0].willChange, checks[0].err
	for i := 1; i < len(checks); i++ {
		status, willChange, err = helpers.SquashCheck(status, willChange, err, checks[1].status, checks[1].willChange, checks[1].err)
	}
	return status, willChange, err
}

// Check decideds what operations to perform
func (f *File) Apply() (err error) {
	if !f.Recurse {
		switch f.State {
		case FSAbsent:
			return f.applyFSAbsent()
		case FSTouch:
			return f.applyFSTouch()
		case FSLink:
			return f.applyFSLink()
		case FSHard:
			return f.applyFSLink()
		case FSDirectory:
			return f.applyFSDirectory()
		case FSFile:
			fallthrough
		case "":
			return f.applyFSFile()
		default:
			return nil
		}
	}
	if f.State != FSDirectory {
		return fmt.Errorf("cannot use the 'recurse' parameter when 'state' is not set to %s", FSDirectory)
	}
	//First create all parent directories
	f.applyFSDirectory()
	rootStat, err := os.Stat(f.Destination)
	if os.IsNotExist(err) {
		return fmt.Errorf("%q does not exist. cannot recurse", f.Destination)
	}
	if !rootStat.IsDir() {
		return fmt.Errorf("'recurse' paramter used on file %q", f.Destination)
	}
	err = filepath.Walk(f.Destination, func(path string, info os.FileInfo, walkerr error) error {
		if walkerr != nil {
			return err
		}
		filemodule := File{
			Destination: path,
		}
		if f.Mode != nil {
			filemodule.Mode = &mode.Mode{
				Destination: path,
				Mode:        f.Mode.Mode,
			}
		}
		if f.Owner != nil {
			filemodule.Owner = &owner.Owner{
				Destination: path,
				Username:    f.Owner.Username,
				UID:         f.Owner.UID,
				GID:         f.Owner.GID,
			}
		}
		return filemodule.applyFSFile()
	})

	return err
}

func (f *File) checkFSAbsent() (status string, willChange bool, err error) {
	_, err = os.Stat(f.Destination)
	if os.IsNotExist(err) {
		return fmt.Sprintf("%q does not exist", f.Destination), false, nil
	} else if err == nil {
		return fmt.Sprintf("%q does exist. will be deleted", f.Destination), true, nil
	} else {
		return "", false, nil
	}
}
func (f *File) applyFSAbsent() (err error) {
	_, err = os.Stat(f.Destination)
	if os.IsNotExist(err) {
		return nil
	} else {
		return os.Remove(f.Destination)
	}
}

func (f *File) checkFSTouch() (status string, willChange bool, err error) {
	_, err = os.Stat(f.Destination)
	if os.IsNotExist(err) {
		return fmt.Sprintf("%q does not exist. will be created", f.Destination), true, nil
	} else if err == nil {
		s1, c1, w1 := fmt.Sprintf("%q already exist.", f.Destination), false, error(nil)
		s2, c2, w2 := f.checkFSFile()
		return helpers.SquashCheck(s1, c1, w1, s2, c2, w2)
	} else {
		return "", false, err
	}
}

func (f *File) applyFSTouch() (err error) {
	_, err = os.Stat(f.Destination)
	if os.IsNotExist(err) {
		_, err = os.Create(f.Destination)
	}
	if err != nil {
		return err
	}
	err = f.applyMode()
	if err == nil {
		err = f.applyOwner()
	}

	return err
}

func (f *File) checkFSLink() (status string, willChange bool, err error) {
	//TODO Move to preparers
	if f.Source == "" || f.Destination == "" {
		return "", false, errors.New("file.source or file.destination were empty when attemting to create symbolic link")
	}
	//check if destination exist and is a symbolic link
	_, err = os.Stat(f.Source)
	if os.IsNotExist(err) {
		return fmt.Sprintf("source %q does not exist", f.Source, f.Destination), false, nil
	} else if err != nil {
		return "", false, err
	}
	dest, err := os.Lstat(f.Destination)

	// True if the file is a symlink
	if err == nil {
		if (dest.Mode() & os.ModeSymlink) != 0 {
			return fmt.Sprintf("%q is already linked to %q", f.Destination, f.Source), false, nil
		} else {
			return fmt.Sprintf("%q is not linked to %q. will be soft linked to %q", f.Destination, f.Source, f.Source), false, nil
		}
	}

	return fmt.Sprintf("source %q will be linked to %q", f.Source, f.Destination), true, nil
}

func (f *File) applyFSLink() (err error) {
	os.Remove(f.Destination)
	_, err = os.Stat(f.Destination)
	if os.IsNotExist(err) {
		return os.Symlink(f.Source, f.Destination)
	}
	return nil
}

func (f *File) checkFSHard() (status string, willChange bool, err error) {
	if f.Source == "" || f.Destination == "" {
		return "", false, errors.New("file.source or file.destination were empty when attemting to create symbolic link")
	}
	//check if destination exist and is a symbolic link
	src, err := os.Stat(f.Source)
	if os.IsNotExist(err) {
		return fmt.Sprintf("source %q does not exist", f.Source), false, nil
	} else if err != nil {
		return "", false, err
	}
	dest, err := os.Lstat(f.Destination)
	if os.IsNotExist(err) {
		return fmt.Sprintf("%q does not exist. will be hard linked to %q", f.Destination, f.Source), true, nil
	} else if err != nil {
		return "", false, err
	}
	if os.SameFile(src, dest) {
		return fmt.Sprintf("%q is hard linked to %q", f.Destination, f.Source), false, nil
	} else {
		return fmt.Sprintf("%q is not hard linked to %q. will be hard linked", f.Destination, f.Source), true, nil
	}

}

func (f *File) applyFSHard() (err error) {
	_, err = os.Stat(f.Source)
	if os.IsNotExist(err) {
		return fmt.Errorf("%q does not exist. will not create hard link", f.Source)
	} else if err != nil {
		return err
	}
	return os.Link(f.Source, f.Destination)
}

func (f *File) checkFSFile() (status string, willChange bool, err error) {
	s1, c1, e1 := f.checkMode()
	s2, c2, e2 := f.checkOwner()
	return helpers.SquashCheck(s1, c1, e1, s2, c2, e2)
}

func (f *File) applyFSFile() (err error) {
	return multierror.Append(f.applyMode(), f.applyOwner())
}

func (f *File) checkFSDirectory() (string, bool, error) {
	stat, err := os.Stat(f.Destination)
	if os.IsNotExist(err) {
		return fmt.Sprintf("%q does not exist. will be created", f.Destination), true, nil
	} else if err != nil {
		return "", false, err
	}
	if !stat.IsDir() {
		return fmt.Sprintf("file named %q already exist", f.Destination), false, nil
	} else {
		s1, c1, e1 := fmt.Sprintf("folder named %q already exist. ", f.Destination), false, error(nil)
		s2, c2, e2 := f.checkFSFile()
		return helpers.SquashCheck(s1, c1, e1, s2, c2, e2)
	}
}

//TODO change the owner of all created directories.
func (f *File) applyFSDirectory() (err error) {
	_, err = os.Stat(f.Destination)
	var fileMode os.FileMode = 0777
	if f.Mode != nil {
		fileMode = f.Mode.Mode
	}
	//Doesn't exist -> create
	if os.IsNotExist(err) {
		err = os.MkdirAll(f.Destination, fileMode)
	}
	if err == nil {
		err = f.applyOwner()
	}
	if err == nil {
		err = f.applyMode()
	}
	return err
}

func (f *File) checkMode() (status string, willChange bool, err error) {
	if f.Mode != nil {
		return f.Mode.Check()
	} else {
		return "", false, nil
	}
}

func (f *File) applyMode() (err error) {
	if f.Mode != nil {
		return f.Mode.Apply()
	} else {
		return nil
	}
}

func (f *File) checkOwner() (status string, willChange bool, err error) {
	if f.Owner != nil {
		return f.Owner.Check()
	} else {
		return "", false, nil
	}
}

func (f *File) applyOwner() (err error) {
	if f.Owner != nil {
		return f.Owner.Apply()
	} else {
		return nil
	}
}
