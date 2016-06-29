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

package parse_test

import (
	"testing"

	"github.com/asteris-llc/converge/parse"
	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	t.Parallel()

	resources, err := parse.Parse([]byte(`task x {}`))

	assert.NoError(t, err)
	assert.Equal(t, len(resources), 1)
}

func TestParseBad(t *testing.T) {
	t.Parallel()

	resources, err := parse.Parse([]byte(`}`))

	assert.Error(t, err)
	assert.Equal(t, len(resources), 0)
}

func TestParseInvalid(t *testing.T) {
	t.Parallel()

	_, err := parse.Parse([]byte(`task {}`))

	if assert.Error(t, err) {
		assert.EqualError(t, err, "1 error(s) occurred:\n\n* 1:1: missing name")
	}
}
