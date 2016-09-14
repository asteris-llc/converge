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
	"math"
	"strconv"
	"testing"

	"github.com/asteris-llc/converge/helpers/fakerenderer"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/group"
	"github.com/stretchr/testify/assert"
)

func TestPreparerInterface(t *testing.T) {
	t.Parallel()

	assert.Implements(t, (*resource.Resource)(nil), new(group.Preparer))
}

func TestValidPreparer(t *testing.T) {
	t.Parallel()

	fr := fakerenderer.FakeRenderer{}
	p := group.Preparer{GID: "123", Name: "test", State: string(group.StateAbsent)}
	_, err := p.Prepare(&fr)

	assert.NoError(t, err)
}

func TestValidPreparerNoState(t *testing.T) {
	t.Parallel()

	fr := fakerenderer.FakeRenderer{}
	p := group.Preparer{GID: "123", Name: "test"}
	_, err := p.Prepare(&fr)

	assert.NoError(t, err)
}

func TestInvalidPreparerNoGid(t *testing.T) {
	t.Parallel()

	fr := fakerenderer.FakeRenderer{}
	p := group.Preparer{Name: "test"}
	_, err := p.Prepare(&fr)

	assert.EqualError(t, err, fmt.Sprintf("group requires a \"gid\" parameter"))
}

func TestInvalidPreparerInvalidGid(t *testing.T) {
	t.Parallel()

	gid := strconv.Itoa(math.MaxUint32)
	fr := fakerenderer.FakeRenderer{}
	p := group.Preparer{GID: gid, Name: "test"}
	_, err := p.Prepare(&fr)

	assert.EqualError(t, err, fmt.Sprintf("group \"gid\" parameter out of range"))
}

func TestInvalidPreparerMaxGid(t *testing.T) {
	t.Parallel()

	gid := strconv.Itoa(math.MaxUint32 - 1)
	fr := fakerenderer.FakeRenderer{}
	p := group.Preparer{GID: gid, Name: "test"}
	_, err := p.Prepare(&fr)

	assert.NoError(t, err)
}

func TestInvalidPreparerNoName(t *testing.T) {
	t.Parallel()

	fr := fakerenderer.FakeRenderer{}
	p := group.Preparer{GID: "123"}
	_, err := p.Prepare(&fr)

	assert.EqualError(t, err, fmt.Sprintf("group requires a \"name\" parameter"))
}

func TestInvalidPreparerInvalidState(t *testing.T) {
	t.Parallel()

	fr := fakerenderer.FakeRenderer{}
	p := group.Preparer{GID: "123", Name: "test", State: "test"}
	_, err := p.Prepare(&fr)

	assert.EqualError(t, err, fmt.Sprintf("group \"state\" parameter invalid, use present or absent"))
}
