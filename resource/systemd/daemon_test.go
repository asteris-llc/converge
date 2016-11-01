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
// limitations under the License.package systemd

package systemd_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/asteris-llc/converge/resource/systemd"
	"github.com/stretchr/testify/assert"
)

// Test that every goroutine reloads daemon in order it's called.
func TestDaemonReload(t *testing.T) {
	t.Parallel()
	var wg sync.WaitGroup
	goroutineCount := 10
	order := make(chan int, goroutineCount)

	for i := 0; i < goroutineCount; i++ {
		index := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			systemd.ApplyDaemonReload()
			order <- index
		}()
	}
	wg.Wait()

	// Test that daemon reloads occured in the right order
	prev := -1
	for {
		select {
		case current := <-order:
			assert.True(t, current > prev, fmt.Sprintf("Daemon reloads did not occur in the right order. %d before %d", prev, current))
			current = prev
		default:
			return
		}
	}
}

// Test that every goroutine resets failed units in order it's called.
func TestResettingFailed(t *testing.T) {
	t.Parallel()
	var wg sync.WaitGroup
	goroutineCount := 10
	order := make(chan int, goroutineCount)

	for i := 0; i < goroutineCount; i++ {
		index := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			systemd.ApplyResetFailed("")
			order <- index
		}()
	}
	wg.Wait()

	// Test that daemon reloads occured in the right order
	prev := -1
	for {
		select {
		case current := <-order:
			assert.True(t, current > prev, fmt.Sprintf("Resetting failed units did not occur in the right order. %d before %d", prev, current))
			current = prev
		default:
			return
		}
	}
}
