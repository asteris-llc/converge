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

	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/shell"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	any         = mock.Anything
	exitSuccess = int(0)
	exitFailure = int(1)
)

func Test_Shell_ImplementsTaskInterface(t *testing.T) {
	t.Parallel()
	assert.Implements(t, (*resource.Task)(nil), new(shell.Shell))
}

type MockTask struct {
	mock.Mock
}

func (m *MockTask) Check() (string, bool, error) {
	args := m.Called()
	return args.String(0), args.Bool(1), args.Error(2)
}

func (m *MockTask) Apply() error {
	args := m.Called()
	return args.Error(0)
}
