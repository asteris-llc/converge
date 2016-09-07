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
	"testing"

	"github.com/asteris-llc/converge/helpers/fakerenderer"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/file"
	"github.com/asteris-llc/converge/resource/file/content"
	"github.com/stretchr/testify/assert"
)

func TestPreparerInterface(t *testing.T) {
	t.Parallel()

	assert.Implements(t, (*resource.Resource)(nil), new(content.Preparer))
}

func TestPreparerDestinationIsRequired(t *testing.T) {
	p := &file.Preparer{}
	_, err := p.Prepare(fakerenderer.New())
	if assert.Error(t, err) {
		assert.EqualError(t, err, "file requires a destination parameter")
	}
}

func TestPreparerInvalidState(t *testing.T) {
	p := &file.Preparer{Destination: "/aster/is", State: "foo"}
	_, err := p.Prepare(fakerenderer.New())
	if assert.Error(t, err) {
		assert.EqualError(t, err, "state should be one of present, absent, got \"foo\"")
	}
}

func TestPreparerInvalidType(t *testing.T) {
	p := &file.Preparer{Destination: "/aster/is", Type: "bar"}
	_, err := p.Prepare(fakerenderer.New())
	if assert.Error(t, err) {
		assert.EqualError(t, err, "type should be one of directory, file, hardlink, symlink, got \"bar\"")
	}
}

func TestTargetNotDefinedForSymlink(t *testing.T) {
	p := &file.Preparer{Destination: "/aster/is", Type: "symlink"}
	_, err := p.Prepare(fakerenderer.New())
	if assert.Error(t, err) {
		assert.EqualError(t, err, "must define a target if you are using a \"symlink\"")
	}
}

func TestTargetDefinedForSymlink(t *testing.T) {
	p := &file.Preparer{Destination: "/aster/is", Type: "symlink", Target: "/get/converge"}
	_, err := p.Prepare(fakerenderer.New())
	if !assert.Nil(t, err) {
		assert.EqualError(t, err, "target + symlink should pass")
	}
}

func TestTargetNotDefinedForHardlink(t *testing.T) {
	p := &file.Preparer{Destination: "/aster/is", Type: "hardlink"}
	_, err := p.Prepare(fakerenderer.New())
	if assert.Error(t, err) {
		assert.EqualError(t, err, "must define a target if you are using a \"hardlink\"")
	}
}

func TestTargetDefinedForHardLink(t *testing.T) {
	p := &file.Preparer{Destination: "/aster/is", Type: "hardlink", Target: "/get/converge"}
	_, err := p.Prepare(fakerenderer.New())
	if !assert.Nil(t, err) {
		assert.EqualError(t, err, "target + hardlink should pass")
	}
}

func TestBadPermissions1(t *testing.T) {
	p := &file.Preparer{Destination: "/aster/is", Mode: "999"}
	_, err := p.Prepare(fakerenderer.New())
	if assert.Error(t, err) {
		assert.EqualError(t, err, "\"999\" is not a valid file mode")
	}
}

//text permissions not supported yet
func TestBadPermissions2(t *testing.T) {
	p := &file.Preparer{Destination: "/aster/is", Mode: "rwxrwxrwx"}
	_, err := p.Prepare(fakerenderer.New())
	if assert.Error(t, err) {
		assert.EqualError(t, err, "\"rwxrwxrwx\" is not a valid file mode")
	}
}

func TestValidPermissions(t *testing.T) {
	p := &file.Preparer{Destination: "/aster/is", Mode: "0755"}
	_, err := p.Prepare(fakerenderer.New())
	if !assert.Nil(t, err) {
		assert.EqualError(t, err, "correct permissions of 0755 should pass")
	}
}

func TestValidConfig1(t *testing.T) {
	p := &file.Preparer{
		Destination: "/aster/is",
		Mode:        "0755",
		Type:        "file",
		Force:       "true",
		User:        "root",
		Group:       "root",
	}
	_, err := p.Prepare(fakerenderer.New())
	if !assert.Nil(t, err) {
		assert.EqualError(t, err, "correct configuration should pass")
	}
}

func TestValidConfig2(t *testing.T) {
	p := &file.Preparer{
		Destination: "/aster/is",
		Mode:        "4700",
		Type:        "directory",
		Force:       "false",
		User:        "root",
		Group:       "root",
	}
	_, err := p.Prepare(fakerenderer.New())
	if !assert.Nil(t, err) {
		assert.EqualError(t, err, "correct configuration should pass")
	}
}
