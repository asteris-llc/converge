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
	"fmt"
	"time"

	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/shell"
)

const (
	defaultInterval = 5 * time.Second
	defaultTimeout  = 10 * time.Second
	defaultRetries  = 5
)

// Wait waits for a shell task to return 0 or reaches max failure threshold
type Wait struct {
	*shell.Shell
	GracePeriod time.Duration
	Interval    time.Duration
	MaxRetry    int
	RetryCount  int
	Duration    time.Duration
}

// Apply retries the check until it passes or returns max failure threshold
func (w *Wait) Apply() (resource.TaskStatus, error) {
	startTime := time.Now()

	retries := w.MaxRetry
	if retries <= 0 {
		retries = defaultRetries
	}

	interval := w.Interval
	if interval <= 0 {
		interval = defaultInterval
	}

	after := w.GracePeriod
waitLoop:
	for {
		select {
		case <-time.After(after):
			w.RetryCount++
			after = interval
			results, err := w.CmdGenerator.Run(w.CheckStmt)
			if err != nil {
				return nil, err
			}

			w.Status = w.Status.Cons(fmt.Sprintf("check %d", w.RetryCount), results)
			w.CheckStatus = results

			if results.ExitStatus == 0 {
				break waitLoop
			}

			if w.RetryCount >= retries {
				break waitLoop
			}
		}
	}

	w.Duration = time.Since(startTime)
	return w, nil
}

// Messages returns a summary of the attempts
func (w *Wait) Messages() []string {
	var messages []string
	passed := w.StatusCode() == resource.StatusNoChange

	if passed {
		messages = append(messages, fmt.Sprintf("Passed after %d retries (%v)", w.RetryCount, w.Duration))
	} else {
		messages = append(messages, fmt.Sprintf("Failed after %d retries (%v)", w.RetryCount, w.Duration))
		last := w.Status.Last()
		if last != nil {
			messages = append(messages, fmt.Sprintf("Last attempt: %s", last.Summarize()))
		}
	}

	return messages
}
