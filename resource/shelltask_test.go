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
	"fmt"
	"testing"

	"github.com/asteris-llc/converge/resource"
	"github.com/stretchr/testify/assert"
)

func TestShellTaskInterfaces(t *testing.T) {
	t.Parallel()

	assert.Implements(t, (*resource.Resource)(nil), new(resource.ShellTask))
	assert.Implements(t, (*fmt.Stringer)(nil), new(resource.ShellTask))
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
}

func TestShellTaskValidateInvalidCheck(t *testing.T) {
	t.Parallel()

	st := resource.ShellTask{
		Name:           "test",
		RawCheckSource: "if do then; esac",
	}

	err := st.Prepare(nil)
	if assert.Error(t, err) {
		assert.EqualError(t, err, "task.test.check: exit status 2")
	}
}

func TestShellTaskValidateInvalidApply(t *testing.T) {
	t.Parallel()

	st := resource.ShellTask{
		Name:           "test",
		RawApplySource: "if do then; esac",
	}

	err := st.Prepare(nil)
	if assert.Error(t, err) {
		assert.EqualError(t, err, "task.test.apply: exit status 2")
	}
}

func TestShellTaskCheckNeedsChange(t *testing.T) {
	t.Parallel()

	st := resource.ShellTask{
		RawCheckSource: "echo test && exit 1",
	}
	assert.NoError(t, st.Prepare(&resource.Module{}))

	current, change, err := st.Check()
	assert.Equal(t, "test\n", current)
	assert.True(t, change)
	assert.Nil(t, err)
}

func TestShellTaskCheckNoChange(t *testing.T) {
	t.Parallel()

	st := resource.ShellTask{
		RawCheckSource: "echo test",
	}
	assert.NoError(t, st.Prepare(&resource.Module{}))

	current, change, err := st.Check()
	assert.Equal(t, "test\n", current)
	assert.False(t, change)
	assert.Nil(t, err)
}

func TestShellTaskApplySuccess(t *testing.T) {
	t.Parallel()

	st := resource.ShellTask{
		RawApplySource: "echo test",
	}
	assert.NoError(t, st.Prepare(&resource.Module{}))

	new, success, err := st.Apply()
	assert.Equal(t, "test\n", new)
	assert.True(t, success)
	assert.NoError(t, err)
}

func TestShellTaskApplyError(t *testing.T) {
	t.Parallel()

	st := resource.ShellTask{
		RawApplySource: "echo bad && exit 1",
	}
	assert.NoError(t, st.Prepare(&resource.Module{}))

	new, success, err := st.Apply()
	assert.Equal(t, "bad\n", new)
	assert.False(t, success)
	assert.NoError(t, err)
}

func TestShellTaskApplyDependencies(t *testing.T) {
	t.Parallel()

	var (
		param = &resource.Param{
			Name: "test",
		}

		st = &resource.ShellTask{
			RawApplySource: "{{param `test`}}",
		}

		mod = &resource.Module{
			Resources: []resource.Resource{
				param,
				st,
			},
			RenderedArgs: resource.Values{
				"test": resource.Value("test"),
			},
		}
	)

	assert.NoError(t, param.Prepare(mod))
	assert.NoError(t, st.Prepare(mod))

	assert.Equal(
		t,
		[]string{"param.test"},
		st.Depends(),
	)
}

func TestShellTaskCheckDependencies(t *testing.T) {
	t.Parallel()

	var (
		param = &resource.Param{
			Name: "test",
		}

		st = &resource.ShellTask{
			RawCheckSource: "{{param `test`}}",
		}

		mod = &resource.Module{
			Resources: []resource.Resource{
				param,
				st,
			},
			RenderedArgs: resource.Values{
				"test": resource.Value("test"),
			},
		}
	)

	assert.NoError(t, param.Prepare(mod))
	assert.NoError(t, st.Prepare(mod))

	assert.Equal(
		t,
		[]string{"param.test"},
		st.Depends(),
	)
}
