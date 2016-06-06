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

package resource_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/asteris-llc/converge/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTemplateInterfaces(t *testing.T) {
	t.Parallel()

	assert.Implements(t, (*resource.Resource)(nil), new(resource.Template))
	assert.Implements(t, (*resource.Monitor)(nil), new(resource.Template))
	assert.Implements(t, (*resource.Task)(nil), new(resource.Template))
}

func TestTemplateValid(t *testing.T) {
	t.Parallel()

	tmpl := resource.Template{RawContent: `{{.}}`}
	assert.NoError(t, tmpl.Prepare(&resource.Module{}))

	assert.NoError(t, tmpl.Validate())
}

func TestTemplateInvalid(t *testing.T) {
	t.Parallel()

	tmpl := resource.Template{RawContent: "{{"}
	assert.NoError(t, tmpl.Prepare(&resource.Module{}))

	err := tmpl.Validate()
	assert.Error(t, err)
}

func TestTemplateDestinationRenders(t *testing.T) {
	t.Parallel()

	tmpl := resource.Template{RawDestination: "{{1}}"}
	assert.NoError(t, tmpl.Prepare(&resource.Module{}))
	assert.NoError(t, tmpl.Validate())

	assert.Equal(t, "1", tmpl.Destination())
}

func TestTemplateCheckEmptyFile(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "test-check-empty-file")
	require.NoError(t, err)
	defer func() { require.NoError(t, os.Remove(tmpfile.Name())) }()

	tmpl := resource.Template{
		RawDestination: tmpfile.Name(),
		RawContent:     "this is a test",
	}
	assert.NoError(t, tmpl.Prepare(&resource.Module{}))
	assert.NoError(t, tmpl.Validate())

	current, change, err := tmpl.Check()
	assert.Equal(t, "", current)
	assert.True(t, change)
	assert.NoError(t, err)
}

func TestTemplateCheckEmptyDir(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "test-check-empty-dir")
	require.NoError(t, err)
	defer func() { require.NoError(t, os.RemoveAll(tmpdir)) }()

	tmpl := resource.Template{
		RawDestination: tmpdir,
		RawContent:     "this is a test",
	}
	assert.NoError(t, tmpl.Prepare(&resource.Module{}))
	assert.NoError(t, tmpl.Validate())

	current, change, err := tmpl.Check()
	assert.Equal(t, "", current)
	assert.True(t, change)
	if assert.Error(t, err) {
		assert.EqualError(
			t,
			err,
			fmt.Sprintf("cannot template %q, is a directory", tmpdir),
		)
	}
}

func TestTemplateCheckContentGood(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "test-check-content-good")
	require.NoError(t, err)
	defer func() { require.NoError(t, os.RemoveAll(tmpfile.Name())) }()

	_, err = tmpfile.Write([]byte("this is a test"))
	require.NoError(t, err)
	require.NoError(t, tmpfile.Sync())

	tmpl := resource.Template{
		RawDestination: tmpfile.Name(),
		RawContent:     "this is a test",
	}
	assert.NoError(t, tmpl.Prepare(&resource.Module{}))
	assert.NoError(t, tmpl.Validate())

	current, change, err := tmpl.Check()
	assert.Equal(t, "this is a test", current)
	assert.False(t, change)
	assert.NoError(t, err)
}

func TestTemplateApply(t *testing.T) {
	t.Parallel()

	tmpfile, err := ioutil.TempFile("", "test-check-empty-file")
	require.NoError(t, err)
	defer func() { require.NoError(t, os.Remove(tmpfile.Name())) }()

	tmpl := resource.Template{
		RawDestination: tmpfile.Name(),
		RawContent:     "{{1}}",
	}
	assert.NoError(t, tmpl.Prepare(&resource.Module{}))
	assert.NoError(t, tmpl.Validate())

	new, success, err := tmpl.Apply()
	assert.Equal(t, "1", new)
	assert.True(t, success)
	assert.NoError(t, err)

	// read the new file
	content, err := ioutil.ReadFile(tmpfile.Name())
	assert.Equal(t, "1", string(content))
	assert.NoError(t, err)
}

func TestTemplateApplyPermission(t *testing.T) {
	t.Parallel()

	tmpfile, err := ioutil.TempFile("", "test-template-apply-permission")
	require.NoError(t, err)
	defer func() { require.NoError(t, os.Remove(tmpfile.Name())) }()

	tmpl := resource.Template{
		RawDestination: tmpfile.Name(),
		RawContent:     "{{1}}",
	}
	assert.NoError(t, tmpl.Prepare(&resource.Module{}))
	assert.NoError(t, tmpl.Validate())

	_, success, err := tmpl.Apply()
	assert.True(t, success)
	assert.NoError(t, err)

	// stat the new file
	stat, err := os.Stat(tmpfile.Name())
	assert.NoError(t, err)

	perm := stat.Mode().Perm()
	assert.Equal(t, os.FileMode(0600), perm)
}
