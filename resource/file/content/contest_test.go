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

package content_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/file/content"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContentInterface(t *testing.T) {
	t.Parallel()

	assert.Implements(t, (*resource.Task)(nil), new(content.Content))
}

func TestContentCheckEmptyFile(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "test-check-empty-file")
	require.NoError(t, err)
	defer func() { require.NoError(t, os.Remove(tmpfile.Name())) }()

	tmpl := content.Content{
		Destination: tmpfile.Name(),
		Content:     "this is a test",
	}

	current, change, err := tmpl.Check()
	assert.Equal(t, "", current)
	assert.True(t, change)
	assert.NoError(t, err)
}

func TestContentCheckEmptyDir(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "test-check-empty-dir")
	require.NoError(t, err)
	defer func() { require.NoError(t, os.RemoveAll(tmpdir)) }()

	tmpl := content.Content{
		Destination: tmpdir,
		Content:     "this is a test",
	}

	current, change, err := tmpl.Check()
	assert.Equal(t, "", current)
	assert.True(t, change)
	if assert.Error(t, err) {
		assert.EqualError(
			t,
			err,
			fmt.Sprintf("cannot content %q, is a directory", tmpdir),
		)
	}
}

func TestContentCheckContentGood(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "test-check-content-good")
	require.NoError(t, err)
	defer func() { require.NoError(t, os.RemoveAll(tmpfile.Name())) }()

	_, err = tmpfile.Write([]byte("this is a test"))
	require.NoError(t, err)
	require.NoError(t, tmpfile.Sync())

	tmpl := content.Content{
		Destination: tmpfile.Name(),
		Content:     "this is a test",
	}

	current, change, err := tmpl.Check()
	assert.Equal(t, "this is a test", current)
	assert.False(t, change)
	assert.NoError(t, err)
}

func TestContentApply(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "test-check-empty-file")
	require.NoError(t, err)
	defer func() { require.NoError(t, os.Remove(tmpfile.Name())) }()

	tmpl := content.Content{
		Destination: tmpfile.Name(),
		Content:     "1",
	}

	assert.NoError(t, tmpl.Apply())

	// read the new file
	content, err := ioutil.ReadFile(tmpfile.Name())
	assert.Equal(t, "1", string(content))
	assert.NoError(t, err)
}

func TestContentApplyPermissionDefault(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "test-content-apply-permission")
	require.NoError(t, err)
	defer func() { require.NoError(t, os.Remove(tmpfile.Name())) }()

	tmpl := content.Content{
		Destination: tmpfile.Name(),
		Content:     "1",
	}

	assert.NoError(t, tmpl.Apply())

	// stat the new file
	stat, err := os.Stat(tmpfile.Name())
	assert.NoError(t, err)

	perm := stat.Mode().Perm()
	assert.Equal(t, os.FileMode(0600), perm)
}

func TestContentApplyKeepPermission(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "test-content-keep-permission")
	require.NoError(t, err)
	defer func() { require.NoError(t, os.Remove(tmpfile.Name())) }()

	var perm os.FileMode = 0777
	require.NoError(t, os.Chmod(tmpfile.Name(), perm))

	tmpl := content.Content{
		Destination: tmpfile.Name(),
		Content:     "1",
	}

	assert.NoError(t, tmpl.Apply())

	// check permissions matched
	stat, err := os.Stat(tmpfile.Name())
	assert.NoError(t, err)

	assert.Equal(t, perm, stat.Mode().Perm())
}
