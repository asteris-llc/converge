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

package shell_test

import (
	"testing"

	"github.com/asteris-llc/converge/helpers/fakerenderer"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/shell"
	"github.com/stretchr/testify/assert"
)

func TestPreparerInterface(t *testing.T) {
	t.Parallel()

	assert.Implements(t, (*resource.Resource)(nil), new(shell.Preparer))
}

func TestShellValidateValid(t *testing.T) {
	t.Parallel()

	sp := &shell.Preparer{
		Check: "echo test",
		Apply: "echo test",
	}

	_, err := sp.Prepare(&fakerenderer.FakeRenderer{})

	assert.NoError(t, err)
}

func TestShellValidateInvalidCheck(t *testing.T) {
	t.Parallel()

	sp := &shell.Preparer{
		Check: "if do then; esac",
	}

	_, err := sp.Prepare(&fakerenderer.FakeRenderer{})

	if assert.Error(t, err) {
		assert.EqualError(t, err, "exit status 2")
	}
}

func TestShellValidateInvalidApply(t *testing.T) {
	t.Parallel()

	sp := &shell.Preparer{
		Apply: "if do then; esac",
	}

	_, err := sp.Prepare(&fakerenderer.FakeRenderer{})

	if assert.Error(t, err) {
		assert.EqualError(t, err, "exit status 2")
	}
}
