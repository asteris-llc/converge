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
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/asteris-llc/converge/helpers/logging"
	"github.com/fgrid/uuid"
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
