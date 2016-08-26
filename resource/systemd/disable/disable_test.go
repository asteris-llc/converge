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

package disable_test

import (
	"testing"

	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/systemd"
	"github.com/asteris-llc/converge/resource/systemd/disable"
	"github.com/stretchr/testify/assert"
)

func TestTemplateInterface(t *testing.T) {
	t.Parallel()

	assert.Implements(t, (*resource.Task)(nil), new(disable.Disable))
}

func TestCheck(t *testing.T) {
	_, err := systemd.GetDbusConnection()
	if err != nil {
		t.Skip(err)
	}
	task := &disable.Disable{Unit: "systemd-journald.service"}
	assert.NoError(t, task.Validate())

	status, err := task.Check()
	assert.NoError(t, err)
	assert.Equal(t, "property \"UnitFileState\" of unit \"systemd-journald.service\" is \"static\", expected one of [\"disabled\"]", status.Value())
	assert.True(t, status.HasChanges())
}
