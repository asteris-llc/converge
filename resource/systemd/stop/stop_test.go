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

package stop_test

import (
	"os/user"
	"testing"

	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/systemd/start"
	"github.com/asteris-llc/converge/resource/systemd/stop"
	"github.com/stretchr/testify/assert"
)

func TestTemplateInterface(t *testing.T) {
	t.Parallel()

	assert.Implements(t, (*resource.Task)(nil), new(stop.Stop))
}

func TestCheck(t *testing.T) {
	task := &stop.Stop{Unit: "systemd-journald.service", Mode: "replace"}
	assert.NoError(t, task.Validate())

	status, err := task.Check()
	assert.NoError(t, err)
	assert.Equal(t, "property \"ActiveState\" of unit \"systemd-journald.service\" is \"active\", expected one of [\"inactive\"]", status.Value())
	assert.True(t, status.HasChanges())
}

func TestApply(t *testing.T) {
	u, err := user.Current()
	assert.NoError(t, err)

	if u.Uid != "0" {
		return
	}
	revert := &start.Start{Unit: "systemd-journald.service", Mode: "replace"}
	defer revert.Apply()
	task := &stop.Stop{Unit: "systemd-journald.service", Mode: "replace"}
	assert.NoError(t, task.Validate())

	assert.NoError(t, task.Apply())

	status, err := task.Check()
	assert.NoError(t, err)
	assert.Equal(t, "property \"ActiveState\" of unit \"systemd-journald.service\" is \"inactive\", expected one of [\"inactive\"]", status.Value())
	assert.False(t, status.HasChanges())

}
