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

package link

import (
	"fmt"
	"os"
)

type LinkType string

const (
	LTSoft LinkType = "soft"
	LTHard LinkType = "hard"
)

// Content renders a content to disk
type Link struct {
	Source      string
	Destination string
	Type        LinkType
}

func (l *Link) Check() (status string, willChange bool, err error) {
	switch l.Type {
	case LTHard:
		return l.checkHardLink()
	case LTSoft:
		fallthrough
	default:
		return l.checkSoftLink()
	}
}

func (l *Link) checkSoftLink() (status string, willChange bool, err error) {
	//check if destination exist and is a symbolic link
	_, err = os.Stat(l.Source)
	if os.IsNotExist(err) {
		return fmt.Sprintf("source %q does not exist", l.Source, l.Destination), false, nil
	} else if err != nil {
		return "", false, err
	}
	dest, err := os.Lstat(l.Destination)

	// True if the file is a symlink
	if err == nil {
		if (dest.Mode() & os.ModeSymlink) != 0 {
			return fmt.Sprintf("%q is already linked to %q", l.Destination, l.Source), false, nil
		} else {
			return fmt.Sprintf("%q is not linked to %q. will be soft linked to %q", l.Destination, l.Source, l.Source), false, nil
		}
	} else if os.IsNotExist(err) {
		return fmt.Sprintf("source %q will be soft linked to %q", l.Source, l.Destination), true, nil
	} else {
		return "", false, err
	}

}

func (l *Link) checkHardLink() (status string, willChange bool, err error) {
	//check if destination exist and is a symbolic link
	src, err := os.Stat(l.Source)
	if os.IsNotExist(err) {
		return fmt.Sprintf("source %q does not exist", l.Source), false, nil
	} else if err != nil {
		return "", false, err
	}
	dest, err := os.Lstat(l.Destination)
	if os.IsNotExist(err) {
		return fmt.Sprintf("source %q will be hard linked to %q", l.Source, l.Destination), true, nil
	} else if err != nil {
		return "", false, err
	}

	if os.SameFile(src, dest) {
		return fmt.Sprintf("%q is hard linked to %q", l.Source, l.Destination), false, nil
	} else {
		return fmt.Sprintf("%q is not hard linked to %q. will be hard linked", l.Source, l.Destination), true, nil
	}

}

// Apply writes the content to disk
func (l *Link) Apply() (err error) {
	switch l.Type {
	case LTHard:
		return l.applyHardLink()
	case LTSoft:
		fallthrough
	default:
		return l.applySoftLink()
	}
}

func (l *Link) applyHardLink() (err error) {
	os.Remove(l.Destination)
	return os.Link(l.Source, l.Destination)
}

func (l *Link) applySoftLink() (err error) {
	os.Remove(l.Destination)
	return os.Symlink(l.Source, l.Destination)

}
