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

package resource_test

import (
	"testing"

	"github.com/asteris-llc/converge/resource"
	"github.com/stretchr/testify/assert"
)

func TestShellTaskInterfaces(t *testing.T) {
	t.Parallel()

	assert.Implements(t, (*resource.Resource)(nil), new(resource.ShellTask))
	assert.Implements(t, (*resource.Monitor)(nil), new(resource.ShellTask))
	assert.Implements(t, (*resource.Task)(nil), new(resource.ShellTask))
}

func TestShellTaskValidate(t *testing.T) {
	t.Parallel()
	st := resource.ShellTask{
		TaskName:    "testing validation: should pass",
		CheckSource: "echo test",
		ApplySource: "echo test",
	}
	err := st.Validate()
	if err != nil {
		t.Error("validation failed: false positive")
	}
	st = resource.ShellTask{
		TaskName:    "testing validation: should fail",
		CheckSource: "if while do then; fi end esac",
		ApplySource: "echo test",
	}
	err = st.Validate()
	if err == nil {
		t.Error("validation failed: false negative")
	}
}
