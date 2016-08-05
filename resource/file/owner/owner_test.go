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

package owner_test

import (
	"fmt"
	"io/ioutil"
	"os"
	osuser "os/user"
	"strconv"
	"testing"

	"github.com/asteris-llc/converge/helpers"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/file/owner"
	"github.com/stretchr/testify/assert"
)

func TestTemplateInterface(t *testing.T) {
	t.Parallel()

	assert.Implements(t, (*resource.Task)(nil), new(owner.Owner))
}

func TestCheck(t *testing.T) {
	defer helpers.HideLogs(t)()

	tmpfile, err := ioutil.TempFile("", "owner_test")
	assert.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	u, err := osuser.Current()
	assert.NoError(t, err)

	o := owner.Owner{
		Username:    "nobody",
		Destination: tmpfile.Name(),
	}

	status, willChange, err := o.Check()
	assert.NoError(t, err)
	assert.Equal(t, u.Username, status)
	assert.True(t, willChange)
}

func TestApply(t *testing.T) {
	defer helpers.HideLogs(t)()

	tmpfile, err := ioutil.TempFile("", "owner_test")
	assert.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	u, err := osuser.Current()
	assert.NoError(t, err)

	if u.Uid != "0" {
		return
	}

	nobody, err := osuser.Lookup("nobody")
	if err != nil {
		return
	}

	uid, _ := strconv.Atoi(nobody.Uid)
	gid, _ := strconv.Atoi(nobody.Gid)
	o := owner.Owner{
		Username:    nobody.Username,
		UID:         uid,
		GID:         gid,
		Destination: tmpfile.Name(),
	}
	err = o.Apply()
	assert.NoError(t, err)
	status, willChange, err := o.Check()
	fmt.Println(status, willChange, err)
	assert.NoError(t, err)
	assert.Equal(t, nobody.Username, status)
	assert.False(t, willChange)
}
