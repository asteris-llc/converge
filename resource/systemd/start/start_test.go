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

package start_test

import (
	"os/user"
	"testing"

	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/systemd"
	"github.com/asteris-llc/converge/resource/systemd/start"
	"github.com/stretchr/testify/assert"
)

func TestTemplateInterface(t *testing.T) {
	t.Parallel()

	assert.Implements(t, (*resource.Task)(nil), new(start.Start))
}

func TestCheck(t *testing.T) {
	_, err := systemd.GetDbusConnection()
	if err != nil {
		t.Skip(err)
	}
	task := &start.Start{Unit: "systemd-journald.service", Mode: "replace"}
	assert.NoError(t, task.Validate())

	status, err := task.Check()
	assert.NoError(t, err)
	assert.Equal(t, "property \"ActiveState\" of unit \"systemd-journald.service\" is \"active\", expected one of [\"active\"]", status.Value())
	assert.False(t, status.HasChanges())
}

func TestApply(t *testing.T) {
	_, err := systemd.GetDbusConnection()
	if err != nil {
		t.Skip(err)
	}

	u, err := user.Current()
	assert.NoError(t, err)

	if u.Uid != "0" {
		return
	}

	task := &start.Start{Unit: "systemd-journald.service", Mode: "replace"}
	assert.NoError(t, task.Validate())

	assert.NoError(t, task.Apply())

	status, err := task.Check()
	assert.NoError(t, err)
	assert.Equal(t, "property \"ActiveState\" of unit \"systemd-journald.service\" is \"active\", expected one of [\"active\"]", status.Value())
	assert.False(t, status.HasChanges())
}
