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

package comparsion

import (
	"testing"

	"github.com/asteris-llc/converge/resource"
	"github.com/stretchr/testify/assert"
)

// AssertDiff makes sure that the relevant fields are set
func AssertDiff(t *testing.T, diffs map[string]resource.Diff, name, original, current string) bool {
	var ok bool

	if ok = assert.NotEmpty(t, diffs); !ok {
		return false
	}

	if ok = assert.NotNil(t, diffs[name]); !ok {
		return false
	}

	if ok = assert.Equal(t, original, diffs[name].Original()); !ok {
		return false
	}

	if ok = assert.Equal(t, current, diffs[name].Current()); !ok {
		return false
	}

	return true
}
