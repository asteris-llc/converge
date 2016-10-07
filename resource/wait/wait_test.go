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

package wait_test

import (
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/asteris-llc/converge/helpers/logging"
	"github.com/asteris-llc/converge/resource/shell"
	"github.com/asteris-llc/converge/resource/wait"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// TestApply tests that apply retries a task
func TestApply(t *testing.T) {
	t.Parallel()
	defer logging.HideLogs(t)()

	newWait := func() *wait.Wait {
		return &wait.Wait{
			Shell: &shell.Shell{
				CheckStmt: "test",
				Status:    &shell.CommandResults{},
			},
			Retrier: &wait.Retrier{
				MaxRetry: 3,
				Interval: 10 * time.Millisecond,
			},
		}
	}

	t.Run("passed", func(t *testing.T) {
		wait := newWait()
		m := new(mockExecutor)
		wait.Shell.CmdGenerator = m

		m.On("Run", mock.Anything).
			Return(&shell.CommandResults{ExitStatus: 0}, nil)

		_, err := wait.Apply()
		require.NoError(t, err)

		t.Run("retry count", func(t *testing.T) {
			assert.Equal(t, 1, wait.RetryCount)
		})

		t.Run("status is set", func(t *testing.T) {
			assert.NotNil(t, wait.Status)
		})

		t.Run("check status is set", func(t *testing.T) {
			assert.NotNil(t, wait.CheckStatus)
		})
	})

	t.Run("retried", func(t *testing.T) {
		wait := newWait()
		m := new(mockExecutor)
		wait.Shell.CmdGenerator = m

		m.On("Run", mock.Anything).Return(&shell.CommandResults{
			ResultsContext: shell.ResultsContext{},
			ExitStatus:     1,
		}, nil)

		_, err := wait.Apply()
		require.NoError(t, err)

		t.Run("retry count", func(t *testing.T) {
			assert.Equal(t, 3, wait.RetryCount)
		})
	})

	t.Run("returns an error when command executor fails", func(t *testing.T) {
		wait := newWait()
		m := new(mockExecutor)
		wait.Shell.CmdGenerator = m

		m.On("Run", mock.Anything).
			Return(&shell.CommandResults{ExitStatus: 0}, errors.New("cmd failed"))

		_, err := wait.Apply()
		assert.Error(t, err)
	})
}

// TestMessages tests that messages are added
func TestMessages(t *testing.T) {
	t.Parallel()
	defer logging.HideLogs(t)()

	newWait := func(exitStatus uint32) *wait.Wait {
		return &wait.Wait{
			Shell: &shell.Shell{
				CheckStmt: "test",
				Status: &shell.CommandResults{
					ExitStatus: exitStatus,
				},
			},
			Retrier: &wait.Retrier{},
		}
	}

	t.Run("passed message", func(t *testing.T) {
		wait := newWait(0)
		assert.Equal(t, 1, len(wait.Messages()))

		t.Run("failed messages", func(t *testing.T) {
			assert.Regexp(t, regexp.MustCompile("^Passed after"), wait.Messages()[0])
		})
	})

	t.Run("failed messages", func(t *testing.T) {
		wait := newWait(1)
		assert.Equal(t, 2, len(wait.Messages()))

		t.Run("failed after", func(t *testing.T) {
			assert.Regexp(t, regexp.MustCompile("^Failed after"), wait.Messages()[0])
		})

		t.Run("last attempt", func(t *testing.T) {
			assert.Regexp(t, regexp.MustCompile("^Last attempt"), wait.Messages()[1])
		})
	})
}

type mockExecutor struct {
	mock.Mock
}

func (m *mockExecutor) Run(script string) (*shell.CommandResults, error) {
	args := m.Called(script)
	return args.Get(0).(*shell.CommandResults), args.Error(1)
}
