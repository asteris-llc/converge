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

func TestFileTypes(t *testing.T) {
	type typeTest struct {
		filename string
		fileType string
		err      error
	}
	var typeTests []typeTest

	switch goos := runtime.GOOS; goos {
	case "darwin":
		typeTests = []typeTest{
			{"/bin", "directory", nil},
			{"/etc/group", "file", nil},
			{"nofile", "", &os.PathError{Op: "lstat", Path: "nofile"}},
			{"/var", "symlink", nil},
		}
	}
	for _, tt := range typeTests {
		fi, err := os.Lstat(tt.filename)
		if err == nil {
			ftype, ferr := file.FileType(fi)
			assert.Equal(t, ftype, tt.fileType)
			assert.Equal(t, ferr, tt.err)
		} else {
			assert.Equal(t, tt.err.(*os.PathError).Path, err.(*os.PathError).Path)
		}
	}

}

func TestFileUid(t *testing.T) {
	type uidTest struct {
		filename string
		uid      int
		err      error
	}
	var uidTests []uidTest

	switch goos := runtime.GOOS; goos {
	case "darwin":
		uidTests = []uidTest{
			{"/bin", 0, nil},
			{"/etc/group", 0, nil},
			{"nofile", -1, &os.PathError{Op: "lstat", Path: "nofile"}},
			{"/var/at", 1, nil},
		}
	}

	for _, tt := range uidTests {
		fi, err := os.Lstat(tt.filename)
		if err == nil {
			assert.Equal(t, tt.err, err)
			uid := file.FileUid(fi)
			assert.Equal(t, tt.uid, uid)
		} else {
			assert.Equal(t, tt.err.(*os.PathError).Path, err.(*os.PathError).Path)
		}
	}
}

func TestFileGid(t *testing.T) {
	type gidTest struct {
		filename string
		gid      int
		err      error
	}
	var gidTests []gidTest

	switch goos := runtime.GOOS; goos {
	case "darwin":
		gidTests = []gidTest{
			{"/bin", 0, nil},       //dir
			{"/etc", 0, nil},       //symlink
			{"/etc/group", 0, nil}, //file
			{"/var/empty", 3, nil},
			{"nofile", -1, &os.PathError{Op: "lstat", Path: "nofile"}},
		}
	}

	for _, tt := range gidTests {
		fi, err := os.Lstat(tt.filename)
		if err == nil {
			assert.Equal(t, tt.err, err)
			gid := file.FileGid(fi)
			assert.Equal(t, tt.gid, gid)
		} else {
			assert.Equal(t, tt.err.(*os.PathError).Path, err.(*os.PathError).Path)
		}
	}
}

func TestFileUsername(t *testing.T) {
	type usernameTest struct {
		filename string
		username string
		err      error
	}
	var usernameTests []usernameTest

	switch goos := runtime.GOOS; goos {
	case "darwin":
		usernameTests = []usernameTest{
			{"/bin", "root", nil},
			{"/etc/group", "root", nil},
			{"nofile", "", &os.PathError{Op: "lstat", Path: "nofile"}},
			{"/var/at", "daemon", nil},
		}
	}

	for _, tt := range usernameTests {
		fi, err := os.Lstat(tt.filename)
		if err == nil {
			assert.Equal(t, tt.err, err)
			username, ferr := file.FileOwner(fi)
			if ferr == nil {
				assert.Equal(t, tt.username, username)
			}
		} else {
			assert.Equal(t, tt.err.(*os.PathError).Path, err.(*os.PathError).Path)
		}
	}
}

func TestFileGroup(t *testing.T) {
	type groupTest struct {
		filename string
		group    string
		err      error
	}
	var groupTests []groupTest

	switch goos := runtime.GOOS; goos {
	case "darwin":
		groupTests = []groupTest{
			{"/bin", "wheel", nil},
			{"/etc/group", "wheel", nil},
			{"nofile", "", &os.PathError{Op: "lstat", Path: "nofile"}},
			{"/var/empty", "sys", nil},
		}
	}

	for _, tt := range groupTests {
		fi, err := os.Lstat(tt.filename)
		if err == nil {
			assert.Equal(t, tt.err, err)
			group, ferr := file.FileGroup(fi)
			if ferr == nil {
				assert.Equal(t, tt.group, group)
			}
		} else {
			assert.Equal(t, tt.err.(*os.PathError).Path, err.(*os.PathError).Path)
		}
	}
}
