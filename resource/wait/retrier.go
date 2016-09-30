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
	defaultInterval = 5 * time.Second
	defaultTimeout  = 10 * time.Second
	defaultRetries  = 5
)

// Retrier can be included in resources to provide retry capabilities
type Retrier struct { // TODO: rename?
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

	retries := r.MaxRetry
	if retries <= 0 {
		retries = defaultRetries
	}

	interval := r.Interval
	if interval <= 0 {
		interval = defaultInterval
	}

	after := r.GracePeriod
	ok := false
waitLoop:
	for {
		select {
		case <-time.After(after):
			if ok {
				break waitLoop
			}

			r.RetryCount++
			after = interval

			var err error
			ok, err = retryFunc()
			if err != nil {
				return false, err
			}

			if ok {
				after = r.GracePeriod
				continue
			}

			if r.RetryCount >= retries {
				break waitLoop
			}
		}
	}

	r.Duration = time.Since(startTime)
	return ok, nil
}
