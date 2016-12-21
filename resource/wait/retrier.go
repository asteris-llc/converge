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

package wait

import "time"

const (
	// DefaultInterval is the default amount of time to wait in between checks
	DefaultInterval = 5 * time.Second

	// DefaultGracePeriod is the amount of time to wait before running the first
	// check and after a successful check
	DefaultGracePeriod = time.Duration(0)

	// DefaultRetries is the default number of times to retry before failing
	DefaultRetries = 5
)

// Retrier can be included in resources to provide retry capabilities
type Retrier struct {
	GracePeriod time.Duration
	Interval    time.Duration
	MaxRetry    int
	RetryCount  int
	Duration    time.Duration
}

// RetryFunc is the function to retry
type RetryFunc func() (bool, error)

// RetryUntil implements a retry loop
func (r *Retrier) RetryUntil(retryFunc RetryFunc) (bool, error) {
	startTime := time.Now()

	for {
		ok, err := retryFunc()
		if err != nil {
			return false, err
		}

		if !ok {
			r.RetryCount++

			if r.RetryCount >= r.MaxRetry {
				return false, nil
			}
			time.Sleep(r.GracePeriod)
		} else {
			break
		}
	}

	r.Duration = time.Since(startTime)
	return true, nil
}

// PrepareRetrier generates a Retrier from preparer input
func PrepareRetrier(interval, gracePeriod *time.Duration, maxRetry *int) *Retrier {
	// set the default values for the retrier interval, grace period, and retries
	retrierInterval := DefaultInterval
	retrierGracePeriod := DefaultGracePeriod
	retrierMaxRetry := DefaultRetries

	// if any of the fields were set by the preparer, use those values instead of
	// the defaults
	if interval != nil {
		retrierInterval = *interval
	}

	if gracePeriod != nil {
		retrierGracePeriod = *gracePeriod
	}

	if maxRetry != nil {
		retrierMaxRetry = *maxRetry
	}

	return &Retrier{
		MaxRetry:    retrierMaxRetry,
		GracePeriod: retrierGracePeriod,
		Interval:    retrierInterval,
	}
}
