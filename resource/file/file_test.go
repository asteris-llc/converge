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

package file_test

import (
	"os"
	"runtime"
	"testing"

	"github.com/asteris-llc/converge/resource/file"
	"github.com/stretchr/testify/assert"
)

// Ensure that GetFileInfo is poulating File structs with the correct information
func TestFileInfo(t *testing.T) {

	var fileTests []file.File

	switch goos := runtime.GOOS; goos {
	case "darwin":
		fileTests = []file.File{
			{Destination: "/bin", State: "present", Type: "directory", FileMode: os.FileMode(0755), User: "root", Group: "wheel"},
			{Destination: "/etc/sudoers", State: "present", Type: "file", FileMode: os.FileMode(0440), User: "root", Group: "wheel"},
			{Destination: "/etc", State: "present", Type: "symlink", FileMode: os.FileMode(0755), User: "root", Group: "wheel"},
			{Destination: "/var/run", State: "present", Type: "directory", FileMode: os.FileMode(0775), User: "root", Group: "daemon"},
		}
	}
	for _, ft := range fileTests {
		fi, err := os.Lstat(ft.Destination)
		if err == nil {
			actual := &file.File{Destination: ft.Destination}
			file.GetFileInfo(actual, fi)
			assert.Equal(t, ft.State, actual.State)
			assert.Equal(t, ft.Type, actual.Type)
			assert.Equal(t, ft.FileMode.String(), actual.FileMode.String())
			assert.Equal(t, ft.User, actual.User)
			assert.Equal(t, ft.Group, actual.Group)
		}
	}

}
