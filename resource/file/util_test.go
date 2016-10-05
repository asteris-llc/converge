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

package file

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestModeType checks to see if type bits are set correctly
func TestModeType(t *testing.T) {
	t.Run("modeType", func(t *testing.T) {
		var tests = []struct {
			mode     uint32
			filetype Type
			expected string
		}{
			{uint32(0750), TypeDirectory, "drwxr-x---"},
			{uint32(0557), TypeFile, "-r-xr-xrwx"},
			{uint32(0440), TypeSymlink, "Lr--r-----"},
		}

		for _, tt := range tests {
			m := ModeType(tt.mode, tt.filetype)
			assert.Equal(t, tt.expected, os.FileMode(m).String())
		}
	})
}
