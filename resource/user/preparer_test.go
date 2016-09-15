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

package user_test

import (
	"fmt"
	"math"
	"strconv"
	"testing"

	"github.com/asteris-llc/converge/helpers/fakerenderer"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/user"
	"github.com/stretchr/testify/assert"
)

func TestPreparerInterface(t *testing.T) {
	t.Parallel()

	assert.Implements(t, (*resource.Resource)(nil), new(user.Preparer))
}

func TestValidPreparer(t *testing.T) {
	t.Parallel()

	fr := fakerenderer.FakeRenderer{}
	p := user.Preparer{UID: "1234", GID: "123", Username: "test", Name: "test", HomeDir: "tmp", State: string(user.StateAbsent)}
	_, err := p.Prepare(&fr)

	assert.NoError(t, err)
}

func TestInvalidPreparerNoUsername(t *testing.T) {
	t.Parallel()

	fr := fakerenderer.FakeRenderer{}
	p := user.Preparer{UID: "1234", GID: "123", Name: "test", HomeDir: "tmp"}
	_, err := p.Prepare(&fr)

	assert.EqualError(t, err, fmt.Sprintf("user requires a \"username\" parameter"))
}

func TestValidPreparerNoUID(t *testing.T) {
	t.Parallel()

	fr := fakerenderer.FakeRenderer{}
	p := user.Preparer{GID: "123", Username: "test", Name: "test", HomeDir: "tmp", State: string(user.StateAbsent)}
	_, err := p.Prepare(&fr)

	assert.NoError(t, err)
}

func TestValidPreparerMaxUID(t *testing.T) {
	t.Parallel()

	uid := strconv.Itoa(math.MaxUint32 - 1)
	fr := fakerenderer.FakeRenderer{}
	p := user.Preparer{UID: uid, Username: "test"}
	_, err := p.Prepare(&fr)

	assert.NoError(t, err)
}

func TestInvalidPreparerInvalidUID(t *testing.T) {
	t.Parallel()

	uid := strconv.Itoa(math.MaxUint32)
	fr := fakerenderer.FakeRenderer{}
	p := user.Preparer{UID: uid, Username: "test"}
	_, err := p.Prepare(&fr)

	assert.EqualError(t, err, fmt.Sprintf("user \"uid\" parameter out of range"))
}

func TestValidPreparerNoGID(t *testing.T) {
	t.Parallel()

	fr := fakerenderer.FakeRenderer{}
	p := user.Preparer{UID: "1234", Username: "test", Name: "test", HomeDir: "tmp", State: string(user.StateAbsent)}
	_, err := p.Prepare(&fr)

	assert.NoError(t, err)
}

func TestValidPreparerMaxGID(t *testing.T) {
	t.Parallel()

	gid := strconv.Itoa(math.MaxUint32 - 1)
	fr := fakerenderer.FakeRenderer{}
	p := user.Preparer{GID: gid, Username: "test"}
	_, err := p.Prepare(&fr)

	assert.NoError(t, err)
}

func TestInvalidPreparerInvalidGID(t *testing.T) {
	t.Parallel()

	gid := strconv.Itoa(math.MaxUint32)
	fr := fakerenderer.FakeRenderer{}
	p := user.Preparer{GID: gid, Username: "test"}
	_, err := p.Prepare(&fr)

	assert.EqualError(t, err, fmt.Sprintf("user \"gid\" parameter out of range"))
}

func TestValidPreparerNoName(t *testing.T) {
	t.Parallel()

	fr := fakerenderer.FakeRenderer{}
	p := user.Preparer{UID: "1234", GID: "123", Username: "test", HomeDir: "tmp", State: string(user.StateAbsent)}
	_, err := p.Prepare(&fr)

	assert.NoError(t, err)
}

func TestValidPreparerNoHomeDir(t *testing.T) {
	t.Parallel()

	fr := fakerenderer.FakeRenderer{}
	p := user.Preparer{UID: "1234", GID: "123", Username: "test", Name: "test", State: string(user.StateAbsent)}
	_, err := p.Prepare(&fr)

	assert.NoError(t, err)
}

func TestValidPreparerNoState(t *testing.T) {
	t.Parallel()

	fr := fakerenderer.FakeRenderer{}
	p := user.Preparer{UID: "1234", GID: "123", Username: "test", Name: "test", HomeDir: "tmp"}
	_, err := p.Prepare(&fr)

	assert.NoError(t, err)
}
