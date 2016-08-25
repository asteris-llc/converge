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

	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/file/absent"
	"github.com/stretchr/testify/assert"
)

func TestTemplateInterface(t *testing.T) {
	t.Parallel()

	assert.Implements(t, (*resource.Task)(nil), new(absent.Absent))
}

func TestCheckFileExist(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "absent_test")
	defer os.Remove(tmpfile.Name())

	abs := absent.Absent{Destination: tmpfile.Name()}
	assert.NoError(t, abs.Validate())

	status, err := abs.Check()
	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("%q exist", tmpfile.Name()), status.Value())
	assert.True(t, status.HasChanges())
}

func TestCheckFileDoesNotExist(t *testing.T) {

	testFile := "/tmp/nonexistent"

	abs := absent.Absent{Destination: testFile}
	assert.NoError(t, abs.Validate())

	status, err := abs.Check()
	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("%q does not exist", testFile), status.Value())
	assert.False(t, status.HasChanges())
}

func TestApply(t *testing.T) {

	tmpfile, err := ioutil.TempFile("", "absent_test")
	defer os.Remove(tmpfile.Name())

	abs := absent.Absent{Destination: tmpfile.Name()}
	assert.NoError(t, abs.Validate())
	assert.NoError(t, abs.Apply())

	status, err := abs.Check()
	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("%q does not exist", tmpfile.Name()), status.Value())
	assert.False(t, status.HasChanges())
}
