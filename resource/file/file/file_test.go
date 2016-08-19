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
	"os"
	"testing"

	"github.com/asteris-llc/converge/helpers"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/file/directory"
	"github.com/asteris-llc/converge/resource/file/file"
	"github.com/asteris-llc/converge/resource/file/touch"
	"github.com/stretchr/testify/assert"
)

func TestTemplateInterface(t *testing.T) {
	t.Parallel()

	assert.Implements(t, (*resource.Task)(nil), new(file.File))
}

func TestCheck(t *testing.T) {
	defer helpers.HideLogs(t)()

	nonexistentDirectory := "/tmp/DirecctoryThatllneverExist"
	nonexistentFile := nonexistentDirectory + "/fileNaMeThaTwilLNevereverexist"
	os.RemoveAll(nonexistentDirectory)

	dir := &directory.Directory{Destination: nonexistentDirectory}
	f := &touch.Touch{Destination: nonexistentFile}
	tasks := []resource.Task{
		&file.File{Directory: dir, Touch: f},
	}

	checks := []helpers.CheckValidator{
		helpers.CheckValidatorCreator(fmt.Sprintf("%q does not exist. will be created\n%q does not exist. will be created", nonexistentDirectory, nonexistentFile), true, ""),
	}
	helpers.TaskCheckValidator(tasks, checks, t)
}

func TestApply(t *testing.T) {
	defer helpers.HideLogs(t)()
	nonexistentDirectory := "/tmp/DirecctoryThatllneverExist"
	nonexistentFile := nonexistentDirectory + "/fileNaMeThaTwilLNevereverexist"
	os.RemoveAll(nonexistentDirectory)

	dir := &directory.Directory{Destination: nonexistentDirectory}
	f := &touch.Touch{Destination: nonexistentFile}
	tasks := []resource.Task{
		&file.File{Directory: dir, Touch: f},
	}

	errs := []string{
		"",
	}
	helpers.TaskApplyValidator(tasks, errs, t)
}
