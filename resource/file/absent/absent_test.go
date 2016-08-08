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

package absent_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/asteris-llc/converge/helpers"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/file/absent"
	"github.com/stretchr/testify/assert"
)

func TestTemplateInterface(t *testing.T) {
	t.Parallel()

	assert.Implements(t, (*resource.Task)(nil), new(absent.Absent))
}

func TestCheck(t *testing.T) {
	defer helpers.HideLogs(t)()
	tempDir, err := ioutil.TempDir("", "")
	assert.NoError(t, err)
	defer os.Remove(tempDir)
	tempFile, err := ioutil.TempFile(tempDir, "")
	assert.NoError(t, err)
	defer os.Remove(tempFile.Name())

	tasks := []resource.Task{
		&absent.Absent{Destination: tempFile.Name()},
		&absent.Absent{Destination: tempDir},
	}

	checks := []helpers.CheckValidator{
		helpers.CheckValidatorCreator(fmt.Sprintf("%q does exist. will be deleted", tempFile.Name()), true, ""),
		helpers.CheckValidatorCreator(fmt.Sprintf("%q does exist. will be deleted", tempDir), true, ""),
	}
	helpers.TaskCheckValidator(tasks, checks, t)
}

func TestApply(t *testing.T) {
	defer helpers.HideLogs(t)()
	tempDir, err := ioutil.TempDir("", "")
	assert.NoError(t, err)
	defer os.Remove(tempDir)
	tempFile, err := ioutil.TempFile(tempDir, "")
	assert.NoError(t, err)
	defer os.Remove(tempFile.Name())

	tasks := []resource.Task{
		&absent.Absent{Destination: tempFile.Name()},
		&absent.Absent{Destination: tempDir},
	}

	errs := []string{
		"",
		"",
	}
	helpers.TaskApplyValidator(tasks, errs, t)
}
