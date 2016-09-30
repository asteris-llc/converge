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

	"github.com/asteris-llc/converge/resource/wait"
	"github.com/stretchr/testify/assert"
)

// TestRetryUntilSetsDuration tests that RetryUntil sets the RetryCount
func TestRetryUntilSetsRetryCount(t *testing.T) {
	t.Parallel()
	r := &wait.Retrier{
		Interval: 100 * time.Millisecond,
		MaxRetry: 3,
	}
	r.RetryUntil(func() (bool, error) { return false, nil })
	assert.Equal(t, 3, r.RetryCount)
}

// TestRetryUntilSetsDuration tests that RetryUntil sets the Duration
func TestRetryUntilSetsDuration(t *testing.T) {
	t.Parallel()
	r := &wait.Retrier{
		Interval: 100 * time.Millisecond,
		MaxRetry: 3,
	}
	r.RetryUntil(func() (bool, error) { return false, nil })
	assert.True(t, r.Duration >= 0)
}

// TestRetryUntilFailure tests that RetryUntil returns correctly on a failed
// check
func TestRetryUntilFailure(t *testing.T) {
	t.Parallel()
	r := &wait.Retrier{
		Interval: 100 * time.Millisecond,
		MaxRetry: 3,
	}
	b, err := r.RetryUntil(func() (bool, error) { return false, nil })
	assert.NoError(t, err)
	assert.False(t, b)
}

// TestRetryUntilSuccess tests that RetryUntil returns correctly on a succesful
// check
func TestRetryUntilSuccess(t *testing.T) {
	t.Parallel()
	r := &wait.Retrier{
		Interval: 100 * time.Millisecond,
		MaxRetry: 3,
	}
	b, err := r.RetryUntil(func() (bool, error) { return true, nil })
	assert.NoError(t, err)
	assert.True(t, b)
}

// TestRetryBreaksOnSuccess tests that the Retrier will break out of the loop on
// a successful check
func TestRetryBreaksOnSuccess(t *testing.T) {
	t.Parallel()
	r := &wait.Retrier{
		Interval: 100 * time.Millisecond,
		MaxRetry: 3,
	}
	b, err := r.RetryUntil(func() (bool, error) { return (r.RetryCount == 2), nil })
	assert.NoError(t, err)
	assert.True(t, b)
	assert.Equal(t, 2, r.RetryCount)
}
