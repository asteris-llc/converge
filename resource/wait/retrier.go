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

import (
	"time"

	log "github.com/Sirupsen/logrus"
)

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
	after := r.GracePeriod
	ok := false
	var err error
waitLoop:
	for {
		select {
		case <-time.After(after):
			if ok {
				break waitLoop
			}

			r.RetryCount++
			after = r.Interval

			ok, err = retryFunc()

			if ok {
				after = r.GracePeriod
				continue
			}

			if r.RetryCount >= r.MaxRetry {
				break waitLoop
			}
		}
	}

	r.Duration = time.Since(startTime)
	return ok, err
}

// PrepareRetrier generates a Retrier from preparer input
func PrepareRetrier(interval, gracePeriod string, maxRetry int) *Retrier {
	if maxRetry <= 0 {
		maxRetry = DefaultRetries
	}
	return &Retrier{
		MaxRetry:    maxRetry,
		GracePeriod: parseDuration(gracePeriod, "grace_period", DefaultGracePeriod),
		Interval:    parseDuration(interval, "interval", DefaultInterval),
	}
}

func parseDuration(durstr, field string, defdur time.Duration) time.Duration {
	if durstr != "" {
		if dur, err := time.ParseDuration(durstr); err == nil {
			if dur != time.Duration(0) {
				return dur
			}
		} else {
			log.WithFields(log.Fields{
				"module": "preparer",
				"field":  field,
				"value":  durstr,
			}).Warn("could not parse as a duration.")
		}
	}
	return defdur
}
