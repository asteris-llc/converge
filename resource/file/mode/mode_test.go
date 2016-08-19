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

package mode_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/asteris-llc/converge/helpers"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/file/mode"
	"github.com/stretchr/testify/assert"
)

func TestTemplateInterface(t *testing.T) {
	t.Parallel()

	assert.Implements(t, (*resource.Task)(nil), new(mode.Mode))
}

func TestCheck(t *testing.T) {
	defer helpers.HideLogs(t)()

	tmpfile, err := ioutil.TempFile("", "mode_test")
	assert.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	mode := mode.Mode{Destination: tmpfile.Name(), Mode: os.FileMode(int(0777))}
	status, willChange, err := mode.Check()
	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("%q's mode is 600. should be 777", tmpfile.Name()), status)
	assert.True(t, willChange)
}

func TestApply(t *testing.T) {
	defer helpers.HideLogs(t)()

	tmpfile, err := ioutil.TempFile("", "mode_test")
	assert.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	mode := mode.Mode{Destination: tmpfile.Name(), Mode: os.FileMode(int(0777))}
	err = mode.Apply()
	assert.NoError(t, err)
	status, willChange, err := mode.Check()
	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("%q's mode is 777. should be 777", tmpfile.Name()), status)
	assert.False(t, willChange)
}
