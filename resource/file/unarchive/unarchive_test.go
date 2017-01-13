// Copyright © 2017 Asteris, LLC
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

package unarchive_test

import (
	"testing"

	"github.com/asteris-llc/converge/helpers/fakerenderer"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/file/unarchive"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

// TestUnarchiveInterface tests that Unarchive is properly implemented
func TestUnarchiveInterface(t *testing.T) {
	t.Parallel()

	assert.Implements(t, (*resource.Task)(nil), new(unarchive.Unarchive))
}

// TestCheck tests the cases Check handles
func TestCheck(t *testing.T) {
	t.Parallel()

	u := unarchive.Unarchive{}
	_, err := u.Check(context.Background(), fakerenderer.New())
	assert.NoError(t, err)
}

// TestApply tests the cases Apply handles
func TestApply(t *testing.T) {
	t.Parallel()

	u := unarchive.Unarchive{}
	_, err := u.Apply(context.Background())
	assert.NoError(t, err)
}
