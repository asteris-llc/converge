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
	"testing"
	"time"

	"github.com/asteris-llc/converge/helpers/logging"
	"github.com/asteris-llc/converge/resource/wait"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRetryUntil tests the behavior of the RetryUntil function
func TestRetryUntil(t *testing.T) {
	t.Parallel()
	defer logging.HideLogs(t)()

	t.Run("sets retry count", func(t *testing.T) {
		r := &wait.Retrier{
			Interval: 100 * time.Millisecond,
			MaxRetry: 3,
		}
		r.RetryUntil(func() (bool, error) { return false, nil })
		assert.Equal(t, 3, r.RetryCount)
	})

	t.Run("sets duration", func(t *testing.T) {
		r := &wait.Retrier{
			Interval: 100 * time.Millisecond,
			MaxRetry: 3,
		}
		r.RetryUntil(func() (bool, error) { return false, nil })
		assert.True(t, r.Duration >= 0)
	})

	t.Run("retry task failure", func(t *testing.T) {
		r := &wait.Retrier{
			Interval: 100 * time.Millisecond,
			MaxRetry: 3,
		}
		b, err := r.RetryUntil(func() (bool, error) { return false, nil })
		require.NoError(t, err)
		assert.False(t, b)
	})

	t.Run("retry task success", func(t *testing.T) {
		r := &wait.Retrier{
			Interval: 100 * time.Millisecond,
			MaxRetry: 3,
		}
		b, err := r.RetryUntil(func() (bool, error) { return true, nil })
		require.NoError(t, err)
		assert.True(t, b)
	})

	t.Run("break on success", func(t *testing.T) {
		r := &wait.Retrier{
			Interval: 100 * time.Millisecond,
			MaxRetry: 3,
		}
		b, err := r.RetryUntil(func() (bool, error) { return (r.RetryCount == 2), nil })
		require.NoError(t, err)
		assert.True(t, b)
		assert.Equal(t, 2, r.RetryCount)
	})
}

// TestPrepareRetrier tests the generation of a Retrier from preparer values
func TestPrepareRetrier(t *testing.T) {
	t.Parallel()

	t.Run("sets max retries", func(t *testing.T) {
		r := wait.PrepareRetrier("3s", "5s", 10)
		assert.Equal(t, 10, r.MaxRetry)
	})

	t.Run("default max retries", func(t *testing.T) {
		r := wait.PrepareRetrier("3s", "5s", 0)
		assert.Equal(t, wait.DefaultRetries, r.MaxRetry)
	})

	t.Run("sets interval", func(t *testing.T) {
		r := wait.PrepareRetrier("3s", "", 1)
		assert.Equal(t, 3*time.Second, r.Interval)
	})

	t.Run("default interval", func(t *testing.T) {
		defer logging.HideLogs(t)()
		tests := []string{"", "0s", "0", "unparseable"}
		for _, test := range tests {
			r := wait.PrepareRetrier(test, "5s", 1)
			assert.Equal(t, wait.DefaultInterval, r.Interval)
		}
	})

	t.Run("sets grace period", func(t *testing.T) {
		r := wait.PrepareRetrier("", "2s", 1)
		assert.Equal(t, 2*time.Second, r.GracePeriod)
	})

	t.Run("default grace period", func(t *testing.T) {
		defer logging.HideLogs(t)()
		tests := []string{"", "0s", "0", "unparseable"}
		for _, test := range tests {
			r := wait.PrepareRetrier("5s", test, 1)
			assert.Equal(t, wait.DefaultGracePeriod, r.GracePeriod)
		}
	})
}
