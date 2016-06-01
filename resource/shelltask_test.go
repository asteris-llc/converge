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

func TestShellTaskValidateValid(t *testing.T) {
	t.Parallel()

	st := resource.ShellTask{
		RawCheckSource: "echo test",
		RawApplySource: "echo test",
	}
	assert.NoError(t, st.Prepare(nil))
	assert.NoError(t, st.Validate())
}

func TestShellTaskValidateInvalidCheck(t *testing.T) {
	t.Parallel()

	st := resource.ShellTask{RawCheckSource: "if do then; esac"}
	assert.NoError(t, st.Prepare(nil))

	err := st.Validate()
	if assert.Error(t, err) {
		assert.EqualError(t, err, "check: exit status 2")
	}
}

func TestShellTaskValidateInvalidApply(t *testing.T) {
	t.Parallel()

	st := resource.ShellTask{RawApplySource: "if do then; esac"}
	assert.NoError(t, st.Prepare(nil))

	err := st.Validate()
	if assert.Error(t, err) {
		assert.EqualError(t, err, "apply: exit status 2")
	}
}

func TestShellTaskCheckSource(t *testing.T) {
	t.Parallel()

	st := resource.ShellTask{RawCheckSource: "{{1}}"}
	assert.NoError(t, st.Prepare(&resource.Module{}))

	check, err := st.CheckSource()
	assert.NoError(t, err)
	assert.Equal(t, "1", check)
}

func TestShellTaskApplySource(t *testing.T) {
	t.Parallel()

	st := resource.ShellTask{RawApplySource: "{{1}}"}
	assert.NoError(t, st.Prepare(&resource.Module{}))

	check, err := st.ApplySource()
	assert.NoError(t, err)
	assert.Equal(t, "1", check)
}
