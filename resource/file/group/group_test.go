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

package group_test

import (
	"fmt"
	"io/ioutil"
	"os"
	osuser "os/user"
	"testing"

	"github.com/asteris-llc/converge/helpers"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/file/group"
	"github.com/stretchr/testify/assert"
)

func TestTemplateInterface(t *testing.T) {
	t.Parallel()

	assert.Implements(t, (*resource.Task)(nil), new(group.Group))
}

func TestCheck(t *testing.T) {
	defer helpers.HideLogs(t)()

	tmpfile, err := ioutil.TempFile("", "group_test")
	assert.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	u, err := osuser.Current()
	assert.NoError(t, err)
	ids, err := u.GroupIds()
	assert.NoError(t, err)
	grp, err := osuser.LookupGroupId(ids[0])
	assert.NoError(t, err)

	g := group.Group{
		Group:       grp,
		Destination: tmpfile.Name(),
	}

	status, err := g.Check()
	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("file belongs to group %q should be %q", u.Username, grp.Name), status.Value())
	assert.Equal(t, resource.StatusWontChange, status.StatusCode())
}

func TestApply(t *testing.T) {
	defer helpers.HideLogs(t)()

	tmpfile, err := ioutil.TempFile("", "owner_test")
	assert.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	u, err := osuser.Current()
	if u.Uid != "0" {
		return
	}

	ids, err := u.GroupIds()
	assert.NoError(t, err)
	grp, err := osuser.LookupGroupId(ids[0])
	assert.NoError(t, err)

	g := group.Group{
		Group:       grp,
		Destination: tmpfile.Name(),
	}
	err = g.Apply()
	assert.NoError(t, err)
	status, err := g.Check()
	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("file belongs to group %q should be %q", u.Username, grp.Name), status.Value())
	assert.Equal(t, resource.StatusWontChange, status.StatusCode())
}
