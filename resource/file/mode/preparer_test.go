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

package mode_test

import (
	"testing"

	"github.com/asteris-llc/converge/helpers/fakerenderer"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/file/mode"
	"github.com/stretchr/testify/assert"
)

// TestPreparerInterface tests that file mode has been properly implemented
func TestPreparerInterface(t *testing.T) {
	t.Parallel()

	assert.Implements(t, (*resource.Resource)(nil), new(mode.Preparer))
}

// TestValidPreparer tests file mode Prepare()
func TestValidPreparer(t *testing.T) {
	t.Parallel()
	var testMode uint32 = 777
	fr := fakerenderer.FakeRenderer{}
	prep := mode.Preparer{Destination: "path/to/file", Mode: &testMode}
	_, err := prep.Prepare(&fr)
	assert.NoError(t, err)
}
