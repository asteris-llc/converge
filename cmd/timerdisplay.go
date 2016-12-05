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

package cmd

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/gosuri/uilive"
)

// TimerDisplay displays timing for incremental updates, like plan and apply
type TimerDisplay struct {
	writer  *uilive.Writer
	once    sync.Once
	started bool
	logger  *logrus.Entry

	items []*timerDisplayTimer
	wlock *sync.RWMutex
}

func (td *TimerDisplay) init() {
	td.writer = uilive.New()
	td.writer.RefreshInterval = 250 * time.Millisecond
	td.writer.Out = os.Stderr

	td.logger = logrus.WithField("component", "timer display")
	td.logger.Logger.Out = td.writer.Bypass()

	td.wlock = new(sync.RWMutex)
}

// Message creates the current message to display
func (td *TimerDisplay) Message() string {
	td.once.Do(td.init)

	out := "\nActive:\n"

	td.wlock.RLock()
	defer td.wlock.RUnlock()

	for _, item := range td.items {
		out += item.String() + "\n"
	}

	return out
}

// AddTimer starts a timer display for the given name
func (td *TimerDisplay) AddTimer(name string) {
	td.once.Do(td.init)
	td.logger.WithField("name", name).Debug("adding timer")

	td.wlock.Lock()
	defer td.wlock.Unlock()

	td.items = append(td.items, &timerDisplayTimer{name, time.Now()})
}

// RemoveTimer removes a timer display for the given name
func (td *TimerDisplay) RemoveTimer(name string) {
	td.once.Do(td.init)
	td.logger.WithField("name", name).Debug("removing timer")

	td.wlock.Lock()
	defer td.wlock.Unlock()

	for i, item := range td.items {
		if item.Name == name {
			td.items = append(td.items[:i], td.items[i+1:]...)
		}
	}
}

// Bypass provides an io.Writer that can write without interfering with this display
func (td *TimerDisplay) Bypass() io.Writer {
	td.once.Do(td.init)

	return td.writer.Bypass()
}

// Start starts the display
func (td *TimerDisplay) Start() {
	if !CanUseEscapeSequences() {
		return
	}

	td.once.Do(td.init)
	td.writer.Start()
	td.started = true

	go func() {
		for td.started {
			fmt.Fprint(td.writer, td.Message())
			time.Sleep(td.writer.RefreshInterval)
		}
	}()
}

// Stop stops the display
func (td *TimerDisplay) Stop() {
	if !CanUseEscapeSequences() {
		return
	}

	td.once.Do(td.init)
	td.started = false

	td.writer.Flush()
	fmt.Fprint(td.writer, " ")
	td.writer.Flush()

	td.writer.Stop()
}

type timerDisplayTimer struct {
	Name  string
	Start time.Time
}

func (t *timerDisplayTimer) String() string {
	elapsed := time.Since(t.Start)
	elapsedRounded := elapsed - (elapsed % time.Second)

	// TODO: tab and tabwriter?
	return fmt.Sprintf("%s (running %s)", t.Name, elapsedRounded)
}
