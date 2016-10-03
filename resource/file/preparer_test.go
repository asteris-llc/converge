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

func TestPrepare(t *testing.T) {
	t.Parallel()
	t.Run("validConfig1", func(t *testing.T) {
		perms := new(uint32)
		*perms = uint32(0755)
		p := &file.Preparer{
			Destination: "/aster/is",
			Mode:        perms,
			Type:        "file",
			Force:       true,
			User:        "root",
			Group:       "wheel",
		}
		_, err := p.Prepare(fakerenderer.New())
		if !assert.Nil(t, err) {
			assert.EqualError(t, err, "correct configuration should pass")
		}
	})

	t.Run("validConfig2", func(t *testing.T) {
		perms := new(uint32)
		*perms = 4700
		p := &file.Preparer{
			Destination: "/aster/is",
			Mode:        perms,
			Type:        "directory",
			Force:       false,
			User:        "root",
			Group:       "wheel",
		}
		_, err := p.Prepare(fakerenderer.New())
		if !assert.Nil(t, err) {
			assert.EqualError(t, err, "correct configuration should pass")
		}
	})

	t.Run("badConfigNoDestination", func(t *testing.T) {
		perms := new(uint32)
		*perms = 4700
		p := &file.Preparer{
			Mode:  perms,
			Type:  "directory",
			Force: false,
			User:  "root",
			Group: "wheel",
		}
		_, err := p.Prepare(fakerenderer.New())
		assert.Error(t, err, "file requires a destination parameter")
	})

	t.Run("badConfigType", func(t *testing.T) {
		p := &file.Preparer{
			Destination: "/aster/is",
			Type:        "badType",
		}
		_, err := p.Prepare(fakerenderer.New())
		assert.NotNil(t, err)
		assert.EqualError(t, err, fmt.Sprintf("type should be one of directory, file, hardlink, symlink, got %q", p.Type))
	})

	t.Run("badConfigState", func(t *testing.T) {
		p := &file.Preparer{
			Destination: "/aster/is",
			State:       "badState",
		}
		_, err := p.Prepare(fakerenderer.New())
		assert.NotNil(t, err)
		assert.EqualError(t, err, fmt.Sprintf("state should be one of present, absent, got %q", p.State))
	})

	t.Run("badConfigSymlink", func(t *testing.T) {
		p := &file.Preparer{
			Destination: "/aster/is",
			Type:        "symlink",
		}
		_, err := p.Prepare(fakerenderer.New())
		assert.NotNil(t, err)
		assert.EqualError(t, err, fmt.Sprintf("must define a target if you are using a %q", p.Type))
	})

	t.Run("badConfigHardlink", func(t *testing.T) {
		p := &file.Preparer{
			Destination: "/aster/is",
			Type:        "hardlink",
		}
		_, err := p.Prepare(fakerenderer.New())
		assert.NotNil(t, err)
		assert.EqualError(t, err, fmt.Sprintf("must define a target if you are using a %q", p.Type))
	})

	t.Run("badConfigTargetNolink", func(t *testing.T) {
		p := &file.Preparer{
			Destination: "/aster/is",
			Target:      "/converge",
		}
		_, err := p.Prepare(fakerenderer.New())
		assert.NotNil(t, err)
		assert.EqualError(t, err, fmt.Sprintf("cannot define target on a type of \"file\": target: %q", p.Target))
	})

}
