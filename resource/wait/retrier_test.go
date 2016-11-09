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
	"testing"
	"time"

	"github.com/asteris-llc/converge/helpers/logging"
	"github.com/asteris-llc/converge/resource/wait"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	nilDur   *time.Duration
	zeroDur  time.Duration
	threeDur time.Duration
	fiveDur  time.Duration
	nilInt   *int
	retry    int
	err      error
)

func init() {
	zeroDur = time.Duration(0)

	threeDur, err = time.ParseDuration("3s")
	if err != nil {
		panic(err)
	}

	fiveDur, err = time.ParseDuration("5s")
	if err != nil {
		panic(err)
	}

	retry = 10
}

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

	t.Run("break on error", func(t *testing.T) {
		r := &wait.Retrier{
			Interval: 100 * time.Millisecond,
			MaxRetry: 3,
		}
		b, err := r.RetryUntil(func() (bool, error) {
			if r.RetryCount == 2 {
				return false, errors.New("error")
			}
			return false, nil
		})
		require.Error(t, err)
		assert.False(t, b)
		assert.Equal(t, 2, r.RetryCount)
	})
}

// TestPrepareRetrier tests the generation of a Retrier from preparer values
func TestPrepareRetrier(t *testing.T) {
	defer logging.HideLogs(t)()
	t.Parallel()

	t.Run("sets max retries", func(t *testing.T) {
		r := wait.PrepareRetrier(&threeDur, &fiveDur, &retry)
		assert.Equal(t, 10, r.MaxRetry)
	})

	t.Run("default max retries", func(t *testing.T) {
		r := wait.PrepareRetrier(&threeDur, &fiveDur, nilInt)
		assert.Equal(t, wait.DefaultRetries, r.MaxRetry)
	})

	t.Run("sets interval", func(t *testing.T) {
		r := wait.PrepareRetrier(&threeDur, &zeroDur, &retry)
		assert.Equal(t, 3*time.Second, r.Interval)
	})

	t.Run("default interval", func(t *testing.T) {
		r := wait.PrepareRetrier(nilDur, &fiveDur, &retry)
		assert.Equal(t, wait.DefaultInterval, r.Interval)
	})

	t.Run("sets grace period", func(t *testing.T) {
		r := wait.PrepareRetrier(&zeroDur, &threeDur, &retry)
		assert.Equal(t, 3*time.Second, r.GracePeriod)
	})

	t.Run("default grace period", func(t *testing.T) {
		r := wait.PrepareRetrier(&fiveDur, nilDur, &retry)
		assert.Equal(t, wait.DefaultGracePeriod, r.GracePeriod)
	})
}
