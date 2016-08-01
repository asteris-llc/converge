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
	"log"
	"os"

	"github.com/hashicorp/logutils"
)

var levels = []logutils.LogLevel{"TRACE", "DEBUG", "INFO", "WARN", "ERROR", "FATAL"}

// SetLogLevel sets the log level to the passed-in level, erroring if invalid
func SetLogLevel(level string) error {
	if !validLevel(level) {
		return fmt.Errorf("%q is not a valid log level", level)
	}

	filter := &logutils.LevelFilter{
		Levels:   levels,
		MinLevel: logutils.LogLevel(level),
		Writer:   os.Stderr,
	}
	log.SetOutput(filter)

	return nil
}

func validLevel(level string) bool {
	for _, lvl := range levels {
		if lvl == logutils.LogLevel(level) {
			return true
		}
	}

	return false
}
