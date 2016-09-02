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
	"os/user"
	"testing"

	"github.com/asteris-llc/converge/helpers/fakerenderer"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/file/owner"
	"github.com/stretchr/testify/assert"
)

func TestPreparerInterface(t *testing.T) {
	t.Parallel()

	assert.Implements(t, (*resource.Resource)(nil), new(owner.Preparer))
}

func TestVaildPreparer(t *testing.T) {
	t.Parallel()
	fr := fakerenderer.FakeRenderer{}
	prep := owner.Preparer{Destination: "path/to/file", User: "nobody"}
	_, err := prep.Prepare(&fr)
	assert.NoError(t, err)
}

func TestComposition(t *testing.T) {
	t.Parallel()
	fr := fakerenderer.FakeRenderer{}
	u, err := user.Current()
	assert.NoError(t, err)
	prep := owner.Preparer{Destination: "path/to/file", User: "nobody", Group: u.Username}
	_, err = prep.Prepare(&fr)
	assert.NoError(t, err)
}

func TestInVaildPreparer(t *testing.T) {
	t.Parallel()
	fr := fakerenderer.FakeRenderer{}
	prep := owner.Preparer{Destination: "path/to/file", User: "nonexistent"}
	_, err := prep.Prepare(&fr)
	assert.EqualError(t, err, "user: unknown user nonexistent")
}

func TestInVaildPreparerNoDestination(t *testing.T) {
	t.Parallel()
	fr := fakerenderer.FakeRenderer{}
	prep := owner.Preparer{User: "nobody"}
	_, err := prep.Prepare(&fr)
	assert.EqualError(t, err, fmt.Sprintf("task requires a \"destination\" parameter"))
}

func TestInVaildPreparerNoMode(t *testing.T) {
	t.Parallel()
	fr := fakerenderer.FakeRenderer{}
	prep := owner.Preparer{Destination: "path/to/file"}
	_, err := prep.Prepare(&fr)
	assert.EqualError(t, err, fmt.Sprintf("task requires a \"user\" parameter"))
}
