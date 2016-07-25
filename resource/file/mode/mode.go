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

package mode

import (
	"os"
	"strconv"
)

// Mode monitors the file Mode of a file
type Mode struct {
	Destination string
	Mode        os.FileMode
}

// Check whether the Destination has the right Mode
func (m *Mode) Check() (status string, willChange bool, err error) {
	stat, err := os.Stat(m.Destination)
	if err != nil {
		return
	}

	mode := stat.Mode().Perm()
	return strconv.FormatUint(uint64(mode), 8), m.Mode.Perm() != mode, nil
}

// Apply the changes the Mode
func (m *Mode) Apply() error {
	return os.Chmod(m.Destination, m.Mode.Perm())
}
