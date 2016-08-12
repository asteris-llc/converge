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

package param_test

import (
	"testing"

	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/param"
	"github.com/stretchr/testify/assert"
)

func TestParamInterface(t *testing.T) {
	t.Parallel()

	assert.Implements(t, (*resource.Task)(nil), new(param.Param))
}

func TestParamCheck(t *testing.T) {
	t.Parallel()

	param := &param.Param{Value: "test"}

	status, err := param.Check()
	assert.Equal(t, param.Value, status.Messages()[0])
	assert.False(t, status.Changes())
	assert.NoError(t, err)
}

func TestParamApply(t *testing.T) {
	t.Parallel()

	param := new(param.Param)

	assert.NoError(t, param.Apply())
}
