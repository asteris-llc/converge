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
	"github.com/asteris-llc/converge/types"
	"github.com/stretchr/testify/assert"
)

func TestShellTaskInterfaces(t *testing.T) {
	t.Parallel()

	assert.Implements(t, (*types.Resource)(nil), new(resource.ShellTask))
	assert.Implements(t, (*types.Monitor)(nil), new(resource.ShellTask))
	assert.Implements(t, (*types.Task)(nil), new(resource.ShellTask))
}

func TestShellTaskValidateCheckSource(t *testing.T) {
	t.Parallel()
	st := resource.ShellTask{
		CheckSource: "echo test",
		ApplySource: "echo test",
	}
	assert.Nil(t, st.Validate())
	st = resource.ShellTask{CheckSource: "if do then; esac"}
	assert.NotNil(t, st.Validate())
}

func TestShellTaskValidateApplySource(t *testing.T) {
	t.Parallel()
	st := resource.ShellTask{
		CheckSource: "echo test",
		ApplySource: "echo test",
	}
	assert.Nil(t, st.Validate())
	st = resource.ShellTask{ApplySource: "if do then; esac"}
	assert.NotNil(t, st.Validate())
}
