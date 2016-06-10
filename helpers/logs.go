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

package helpers

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"testing"
)

// HideLogs hides logs during test execution
func HideLogs(t *testing.T) func() {
	var b bytes.Buffer
	writer := bufio.NewWriter(&b)
	log.SetOutput(writer)

	return func() {
		if t.Failed() {
			writer.Flush()
			fmt.Print(b.String())
		}

		log.SetOutput(os.Stderr)
	}
}
