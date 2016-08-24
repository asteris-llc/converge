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

package directory_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/file/directory"
	"github.com/stretchr/testify/assert"
)

func TestTemplateInterface(t *testing.T) {
	t.Parallel()

	assert.Implements(t, (*resource.Task)(nil), new(directory.Directory))
}

// Case folder doesn't exist
func TestCheckFolderDoesNotExist(t *testing.T) {

	tmpDir := "/tmp/nonexistentdirectory"
	dir := directory.Directory{Destination: tmpDir, Force: true}
	assert.NoError(t, dir.Validate())

	status, err := dir.Check()
	assert.NoError(t, err)
	msgs := status.Messages()
	assert.Equal(t, 2, len(msgs))
	assert.Equal(t, fmt.Sprintf("directory %q does not exist", tmpDir), status.Value())
	assert.Equal(t, fmt.Sprintf("directory %q does not exist", tmpDir), msgs[0])
	assert.Equal(t, resource.StatusWillChange, status.StatusCode())
	assert.True(t, status.Changes())
}

// Case folder doesnt exist
func TestApplyFolderDoesNotExist(t *testing.T) {

	tmpDir := "/tmp/nonexistentdirectory"
	defer os.Remove(tmpDir)

	dir := directory.Directory{Destination: tmpDir, Force: true}
	assert.NoError(t, dir.Validate())
	assert.NoError(t, dir.Apply())
	status, err := dir.Check()
	assert.NoError(t, err)
	msgs := status.Messages()
	assert.Equal(t, 2, len(msgs))
	assert.Equal(t, fmt.Sprintf("directory %q exists", tmpDir), status.Value())
	assert.Equal(t, fmt.Sprintf("directory %q exists", tmpDir), msgs[0])
	assert.Equal(t, resource.StatusWontChange, status.StatusCode())
	assert.False(t, status.Changes())
}

// Case folder already exist
func TestCheckFolderExist(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "directory_test")
	assert.NoError(t, err)
	defer os.Remove(tmpDir)

	dir := directory.Directory{Destination: tmpDir, Force: true}
	assert.NoError(t, dir.Validate())

	status, err := dir.Check()
	assert.NoError(t, err)
	msgs := status.Messages()
	assert.Equal(t, 2, len(msgs))
	assert.Equal(t, fmt.Sprintf("directory %q exists", tmpDir), status.Value())
	assert.Equal(t, fmt.Sprintf("directory %q exists", tmpDir), msgs[0])
	assert.Equal(t, resource.StatusWontChange, status.StatusCode())
	assert.False(t, status.Changes())
}

// Case folder already exists
func TestApplyFolderExist(t *testing.T) {

	tmpDir := "/tmp/nonexistentdirectory"
	defer os.Remove(tmpDir)

	dir := directory.Directory{Destination: tmpDir, Force: true}
	assert.NoError(t, dir.Validate())
	assert.NoError(t, dir.Apply())
	status, err := dir.Check()
	assert.NoError(t, err)
	msgs := status.Messages()
	assert.Equal(t, 2, len(msgs))
	assert.Equal(t, fmt.Sprintf("directory %q exists", tmpDir), status.Value())
	assert.Equal(t, fmt.Sprintf("directory %q exists", tmpDir), msgs[0])
	assert.Equal(t, resource.StatusWontChange, status.StatusCode())
	assert.False(t, status.Changes())
}

// Case file already exist
func TestCheckFileExist(t *testing.T) {
	tmpFile, err := ioutil.TempFile("", "directory_test")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	dir := directory.Directory{Destination: tmpFile.Name(), Force: true}
	assert.NoError(t, dir.Validate())

	status, err := dir.Check()
	assert.NoError(t, err)
	msgs := status.Messages()
	assert.Equal(t, 2, len(msgs))
	assert.Equal(t, fmt.Sprintf("file %q exists", tmpFile.Name()), status.Value())
	assert.Equal(t, fmt.Sprintf("file %q exists", tmpFile.Name()), msgs[0])
	assert.Equal(t, resource.StatusWillChange, status.StatusCode())
	assert.True(t, status.Changes())
}

// Case folder already exists
func TestApplyFileExist(t *testing.T) {

	tmpFile, err := ioutil.TempFile("", "directory_test")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	dir := directory.Directory{Destination: tmpFile.Name(), Force: true}
	assert.NoError(t, dir.Validate())
	assert.NoError(t, dir.Apply())

	status, err := dir.Check()
	assert.NoError(t, err)
	msgs := status.Messages()
	assert.Equal(t, 2, len(msgs))
	assert.Equal(t, fmt.Sprintf("directory %q exists", tmpFile.Name()), status.Value())
	assert.Equal(t, fmt.Sprintf("directory %q exists", tmpFile.Name()), msgs[0])
	assert.Equal(t, resource.StatusWontChange, status.StatusCode())
	assert.False(t, status.Changes())
}
