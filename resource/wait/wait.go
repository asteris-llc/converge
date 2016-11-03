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

	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/shell"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

// Wait waits for a shell task to return 0 or reaches max failure threshold
type Wait struct {
	*shell.Shell
	*Retrier
}

// Apply retries the check until it passes or returns max failure threshold
func (w *Wait) Apply(context.Context) (resource.TaskStatus, error) {
	_, err := w.RetryUntil(func() (bool, error) {
		results, err := w.CmdGenerator.Run(w.CheckStmt)
		if err != nil {
			return false, err
		}

		w.Status = w.Status.Cons(fmt.Sprintf("check %d", w.RetryCount), results)
		w.CheckStatus = results
		return results.ExitStatus == 0, nil
	})

	if err != nil {
		return w, errors.Wrapf(err, "failed to run check: %s", w.CheckStmt)
	}
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
