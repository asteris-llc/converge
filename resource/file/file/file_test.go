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

package file_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/asteris-llc/converge/helpers"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/file/file"
	"github.com/asteris-llc/converge/resource/file/mode"
	"github.com/stretchr/testify/assert"
)

func TestTemplateInterface(t *testing.T) {
	t.Parallel()

	assert.Implements(t, (*resource.Task)(nil), new(file.File))
}

func TestCheck(t *testing.T) {
	defer helpers.HideLogs(t)()
	tmpfile, err := ioutil.TempFile("", "file_test")
	assert.NoError(t, err)
	defer os.Remove(tmpfile.Name())
	nonexistentFile := "/tmp/fileNaMeThaTwilLNevereverexist"
	os.Remove(nonexistentFile)

	tasks := []resource.Task{
		&file.File{Destination: tmpfile.Name(), Mode: &mode.Mode{Destination: tmpfile.Name(), Mode: os.FileMode(int(0777))}, State: file.FSFile},
		&file.File{Destination: tmpfile.Name(), State: file.FSAbsent},
		&file.File{Source: tmpfile.Name(), Destination: nonexistentFile, State: file.FSLink},
		&file.File{Source: tmpfile.Name(), Destination: nonexistentFile, State: file.FSHard},
		&file.File{Destination: nonexistentFile, State: file.FSTouch},
		&file.File{Destination: nonexistentFile, State: file.FSDirectory},
	}

	checks := []helpers.CheckValidator{
		helpers.CheckValidatorCreator(fmt.Sprintf("%q's mode is 600. should be 777\n", tmpfile.Name()), true, ""),
		helpers.CheckValidatorCreator(fmt.Sprintf("%q does exist. will be deleted", tmpfile.Name()), true, ""),
		helpers.CheckValidatorCreator(fmt.Sprintf("source %q will be soft linked to %q", tmpfile.Name(), nonexistentFile), true, ""),
		helpers.CheckValidatorCreator(fmt.Sprintf("source %q will be hard linked to %q", tmpfile.Name(), nonexistentFile), true, ""),
		helpers.CheckValidatorCreator(fmt.Sprintf("%q does not exist. will be created", nonexistentFile), true, ""),
		helpers.CheckValidatorCreator(fmt.Sprintf("%q does not exist. will be created", nonexistentFile), true, ""),
	}
	helpers.TaskCheckValidator(tasks, checks, t)
}

func TestApply(t *testing.T) {
	defer helpers.HideLogs(t)()
	tmpfile, err := ioutil.TempFile("", "file_test")
	assert.NoError(t, err)
	defer os.Remove(tmpfile.Name())
	nonexistentFile := "/tmp/fileNaMeThaTwilLNevereverexist"
	os.Remove(nonexistentFile)

	tasks := []resource.Task{
		&file.File{Destination: tmpfile.Name(), Mode: &mode.Mode{Destination: tmpfile.Name(), Mode: os.FileMode(int(0777))}, State: file.FSFile},
		&file.File{Destination: nonexistentFile, State: file.FSTouch},
		&file.File{Destination: nonexistentFile, State: file.FSAbsent},
		&file.File{Source: tmpfile.Name(), Destination: nonexistentFile, State: file.FSLink},
		&file.File{Source: tmpfile.Name(), Destination: nonexistentFile, State: file.FSHard},
		&file.File{Destination: nonexistentFile, State: file.FSDirectory},
	}
	errs := []string{
		"",
		"",
		"",
		"",
		"",
		"",
	}
	helpers.TaskApplyValidator(tasks, errs, t)
}
