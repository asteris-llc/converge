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
	"github.com/asteris-llc/converge/helpers"
	"github.com/asteris-llc/converge/resource/file/directory"
	"github.com/asteris-llc/converge/resource/file/touch"
)

// File wraps basic file properties.
type File struct {
	//file properties
	Directory *directory.Directory
	Touch     *touch.Touch
}

// Check decideds what operations to perform
func (f *File) Check() (status string, willChange bool, err error) {
	if f.Directory != nil {
		status, willChange, err = f.Directory.Check()
	}
	if f.Touch != nil {
		s, w, e := f.Touch.Check()
		status, willChange, err = helpers.SquashCheck(status, willChange, err, s, w, e)
	}
	return status, willChange, err
}

// Check decideds what operations to perform
func (f *File) Apply() (err error) {
	if f.Directory != nil {
		err = f.Directory.Apply()
	}
	if f.Touch != nil {
		err = helpers.MultiErrorAppend(err, f.Touch.Apply())
		if f.Directory != nil && f.Directory.Recurse {
			err = helpers.MultiErrorAppend(err, f.Directory.Apply())
		}
	}
	return err
}
