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
	"testing"

	"github.com/asteris-llc/converge/helpers"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/systemd/common"
	"github.com/asteris-llc/converge/resource/systemd/start"
	"github.com/asteris-llc/converge/resource/systemd/stop"
	"github.com/stretchr/testify/assert"
)

func TestTemplateInterface(t *testing.T) {
	t.Parallel()

	assert.Implements(t, (*resource.Task)(nil), new(stop.Stop))
}

func TestCheck(t *testing.T) {
	defer helpers.HideLogs(t)()

	tasks := []resource.Task{
		&stop.Stop{Unit: "systemd-journald.service"},
	}

	checks := []helpers.CheckValidator{
		helpers.CheckValidatorCreator("", true, ""),
	}
	helpers.TaskCheckValidator(tasks, checks, t)
}

func TestApply(t *testing.T) {
	restart := &start.Start{Unit: "systemd-journald.service", Mode: common.SMReplace}
	defer restart.Apply()

	tasks := []resource.Task{
		&stop.Stop{Unit: "systemd-journald.service", Mode: common.SMReplace},
	}

	errs := []string{
		"",
	}
	helpers.TaskApplyValidator(tasks, errs, t)
}
