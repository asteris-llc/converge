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

package logging_test

import (
	"fmt"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/asteris-llc/converge/helpers/logging"
	"github.com/fgrid/uuid"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func BenchmarkFormatter(b *testing.B) {
	formatter := new(logging.Formatter)

	// get an entry with plenty of parts
	entry := logrus.WithFields(logrus.Fields{
		"function":  "benchmarkFormatter",
		"runID":     uuid.NewV4().String(),
		"component": "benchmark",
		"run":       "STARTED",
		"stage":     "BENCHMARK",
	})
	entry.Message = "this is a test"

	b.ResetTimer()
	for i := 0; i <= b.N; i++ {
		formatter.Format(entry)
	}
}

// TestLogging tests logging formatting
func TestLogging(t *testing.T) {
	t.Parallel()

	f := logging.Formatter{}
	log := logrus.New()

	t.Run("standard error", func(t *testing.T) {
		testErr := fmt.Errorf("test error=%v", 100)
		entry := log.WithError(testErr)
		data, err := f.Format(entry)
		assert.Equal(t, string(data[:]), "timestamp=\"0001-01-01T00:00:00Z\" level=\"UNKNOWN\" msg=\"\" error=\"test error=100\"\n")
		assert.Nil(t, err)
	})

	t.Run("wrapped error", func(t *testing.T) {
		testErr := errors.Wrapf(fmt.Errorf("test error=%v", 100), "wrap")
		entry := log.WithError(testErr)
		data, err := f.Format(entry)
		assert.Equal(t, string(data[:]), "timestamp=\"0001-01-01T00:00:00Z\" level=\"UNKNOWN\" msg=\"\" error=\"wrap: test error=100\"\n")
		assert.Nil(t, err)
	})
}
