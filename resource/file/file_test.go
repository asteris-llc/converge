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
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/asteris-llc/converge/helpers/fakerenderer"
	"github.com/asteris-llc/converge/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Case contains a test case
type Case struct {
	F          File
	Err        error //expected
	HasChanges bool  //expected
}

// run a check/apply/check against test cases
func runner(t *testing.T, name string, tests []Case) {
	t.Run(name, func(t *testing.T) {
		var err error
		for _, tt := range tests {
			name := filepath.Base(tt.F.Destination)
			t.Run("check-"+name, func(t *testing.T) {
				status, err := tt.F.Check(fakerenderer.New())
				assert.Equal(t, tt.Err, err)
				assert.Equal(t, tt.HasChanges, status.HasChanges(), fmt.Sprintf("Differences %+v", status.Diffs()))
			})

			t.Run("apply-"+name, func(t *testing.T) {
				_, err = tt.F.Apply()
				assert.Equal(t, tt.Err, err)
			})

			// check should be false after files are created
			t.Run("post-apply-check-"+name, func(t *testing.T) {
				status, err := tt.F.Check(fakerenderer.New())
				assert.Equal(t, tt.Err, err)
				assert.Equal(t, false, status.HasChanges(), fmt.Sprintf("Differences %+v", status.Diffs()))
			})
		}
	})
}

// TestCheckApply creates, modifies and deletes files in temporary directory
func TestCheckApply(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "converge-file-check-apply")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	var create = []Case{
		{File{Destination: filepath.Join(tmpDir, "file-test"), Mode: mode(0750), State: "present"}, nil, true},
		{File{Destination: filepath.Join(tmpDir, "file-missing"), State: "absent"}, nil, false},
		{File{Destination: filepath.Join(tmpDir, "file-content"), Mode: mode(0666), Content: []byte("converge/content test\n"), State: "present"}, nil, true},
		{File{Destination: filepath.Join(tmpDir, "dir-A"), Mode: mode(0700), State: "present", Type: "directory"}, nil, true},
		{File{Destination: filepath.Join(tmpDir, "dir-missing"), State: "absent"}, nil, false},
		{File{Destination: filepath.Join(tmpDir, "dir-A/dir-B"), State: "present", Type: "directory"}, nil, true},
		{File{Destination: filepath.Join(tmpDir, "dir-C/dir-D/dir-E/dir-F"), State: "present", Type: "directory"}, nil, true},
		{File{Destination: filepath.Join(tmpDir, "file-symlink"), Target: filepath.Join(tmpDir, "file-content"), Type: "symlink"}, nil, true},
		{File{Destination: filepath.Join(tmpDir, "file-hardlink"), Target: filepath.Join(tmpDir, "file-test"), Type: "hardlink"}, nil, true},
	}

	var modify = []Case{
		{File{Destination: filepath.Join(tmpDir, "file-test"), Mode: mode(0700), State: "present"}, nil, true},
		{File{Destination: filepath.Join(tmpDir, "file-content"), Mode: mode(0600), Content: []byte("converge/updated content\n"), State: "present"}, nil, true},
		{File{Destination: filepath.Join(tmpDir, "dir-A"), Mode: mode(0750), State: "present", Type: "directory"}, nil, true},
		{File{Destination: filepath.Join(tmpDir, "dir-C/dir-D/dir-E/dir-F"), Force: true, State: "present", Type: "file"}, nil, true},
		{File{Destination: filepath.Join(tmpDir, "file-symlink"), Force: true, Target: filepath.Join(tmpDir, "dir-C/dir-D/dir-E/dir-F"), Type: "symlink"}, nil, true},
		{File{Destination: filepath.Join(tmpDir, "file-hardlink"), Mode: mode(0550), Force: true, Target: filepath.Join(tmpDir, "file-content"), Type: "hardlink"}, nil, true},
	}

	var delete = []Case{
		{File{Destination: filepath.Join(tmpDir, "file-test"), Mode: mode(0750), State: "absent"}, nil, true},
		{File{Destination: filepath.Join(tmpDir, "file-missing"), State: "absent"}, nil, false},
		{File{Destination: filepath.Join(tmpDir, "file-content"), Content: []byte("converge/content test\n"), State: "absent"}, nil, true},
		{File{Destination: filepath.Join(tmpDir, "dir-A"), Mode: mode(0700), State: "absent", Type: "directory"}, nil, true},
		{File{Destination: filepath.Join(tmpDir, "dir-missing"), State: "absent"}, nil, false},
		{File{Destination: filepath.Join(tmpDir, "dir-A/dir-B"), State: "absent", Type: "directory"}, nil, false},       //parent dir removed first
		{File{Destination: filepath.Join(tmpDir, "dir-C"), State: "absent", Type: "directory", Force: true}, nil, true}, //remove child dirs
		{File{Destination: filepath.Join(tmpDir, "file-symlink"), State: "absent", Target: filepath.Join(tmpDir, "file-content"), Type: "symlink"}, nil, true},
		{File{Destination: filepath.Join(tmpDir, "file-hardlink"), State: "absent", Target: filepath.Join(tmpDir, "file-test"), Type: "hardlink"}, nil, true},
	}

	runner(t, "create", create)
	runner(t, "modify", modify)
	runner(t, "delete", delete)

}

func TestDiffMode(t *testing.T) {
	dirMode := ModeType(0750, "directory")
	symlinkMode := ModeType(0750, "symlink")

	var tests = []struct {
		f      *File
		actual *File
		wanted *File
		diff   *resource.TextDiff
	}{
		{&File{Mode: mode(0770)}, &File{Mode: mode(0750)}, &File{Mode: mode(0770)}, &resource.TextDiff{Default: "", Values: [2]string{"-rwxr-x---", "-rwxrwx---"}}},
		{&File{}, &File{Mode: mode(0740)}, &File{Mode: mode(0740)}, nil},
		{&File{}, &File{}, &File{Mode: mode(int(defaultPermissions))}, &resource.TextDiff{Default: "", Values: [2]string{"----------", "-rwxr-x---"}}},
		{&File{Mode: &dirMode}, &File{}, &File{Mode: &dirMode}, &resource.TextDiff{Default: "", Values: [2]string{"----------", "drwxr-x---"}}},
		{&File{Mode: &dirMode}, &File{Mode: (mode(0444))}, &File{Mode: &dirMode}, &resource.TextDiff{Default: "", Values: [2]string{"-r--r--r--", "drwxr-x---"}}},
		{&File{Mode: &dirMode}, &File{Mode: &dirMode}, &File{Mode: &dirMode}, nil},
		{&File{}, &File{Mode: &dirMode}, &File{Mode: &dirMode}, nil},
		{&File{Mode: &symlinkMode}, &File{}, &File{Mode: &symlinkMode}, &resource.TextDiff{Default: "", Values: [2]string{"----------", "Lrwxr-x---"}}},
		{&File{Mode: &symlinkMode}, &File{Mode: (mode(0444))}, &File{Mode: &symlinkMode}, &resource.TextDiff{Default: "", Values: [2]string{"-r--r--r--", "Lrwxr-x---"}}},
		{&File{Mode: &symlinkMode}, &File{Mode: &symlinkMode}, &File{Mode: &symlinkMode}, nil},
		{&File{}, &File{Mode: &symlinkMode}, &File{Mode: &symlinkMode}, nil},
	}
	for _, tt := range tests {
		t.Run("check", func(t *testing.T) {
			s := &resource.Status{}
			tt.f.diffMode(tt.actual, s)
			assert.EqualValues(t, os.FileMode(*tt.wanted.Mode), os.FileMode(*tt.f.Mode))
			if tt.diff != nil {
				d := s.Differences["permissions"].(resource.TextDiff)
				assert.Equal(t, tt.diff, &d)
			} else {
				assert.Nil(t, s.Differences["permissions"])
			}
		})
	}
}

func TestDiffLink(t *testing.T) {

	tmpDir, err := ioutil.TempDir("", "converge-difflink")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	badfile := filepath.Join(tmpDir, "badfile")
	target := filepath.Join(tmpDir, "link-target")
	link := filepath.Join(tmpDir, "link")

	err = ioutil.WriteFile(target, []byte("content"), os.FileMode(0700))
	require.NoError(t, err)

	var tests = []struct {
		f      *File
		actual *File
		diff   *resource.TextDiff
		err    error
	}{
		{&File{Destination: link, Target: target, Type: "symlink"}, &File{Destination: link, Target: target, Type: "symlink"}, nil, nil},
		{&File{Destination: link, Target: target, Type: "symlink"}, &File{Destination: link}, &resource.TextDiff{Default: "", Values: [2]string{"", target}}, nil},
		{&File{Destination: link, Target: badfile, Type: "hardlink"}, &File{Destination: link}, nil, fmt.Errorf("hardlink target lookup: stat %s: no such file or directory", badfile)},
		{&File{Destination: link, Target: target, Type: "hardlink"}, &File{Destination: badfile}, &resource.TextDiff{Default: "", Values: [2]string{link, target}}, nil},
	}
	for _, tt := range tests {
		t.Run("check", func(t *testing.T) {
			s := &resource.Status{}
			err := tt.f.diffLink(tt.actual, s)
			if tt.err != nil {
				assert.EqualError(t, err, tt.err.Error())
			}
			d := s.Differences[tt.f.Type.String()]
			if tt.diff != nil && d != nil {
				diff := s.Differences[tt.f.Type.String()].(resource.TextDiff)
				assert.Equal(t, tt.diff, &diff)
			} else {
				assert.Nil(t, s.Differences[tt.f.Type.String()])
			}
		})
	}
}

// Ensure that GetFileInfo is poulating File structs with the correct information
func TestFileInfo(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "converge-file-info")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	userInfo, err := user.LookupId(strconv.Itoa(os.Geteuid()))
	require.NoError(t, err)

	groupInfo, err := user.LookupGroupId(strconv.Itoa(os.Getegid()))
	require.NoError(t, err)

	var create = []Case{
		{File{Destination: filepath.Join(tmpDir, "file-test"), Mode: mode(0750), State: "present", Type: "file", UserInfo: userInfo, GroupInfo: groupInfo}, nil, true},
		{File{Destination: filepath.Join(tmpDir, "file-content"), Mode: mode(0666), Content: []byte("converge/content test\n"), State: "present", Type: "file", UserInfo: userInfo, GroupInfo: groupInfo}, nil, true},
		{File{Destination: filepath.Join(tmpDir, "dir"), Mode: modetype(0700, TypeDirectory), State: "present", Type: "directory", UserInfo: userInfo, GroupInfo: groupInfo}, nil, true},
		{File{Destination: filepath.Join(tmpDir, "file-symlink"), State: "present", Target: filepath.Join(tmpDir, "file-content"), Type: "symlink", UserInfo: userInfo, GroupInfo: groupInfo}, nil, true},
		{File{Destination: filepath.Join(tmpDir, "file-hardlink"), Mode: mode(0750), State: "present", Target: filepath.Join(tmpDir, "file-test"), Type: "hardlink", UserInfo: userInfo, GroupInfo: groupInfo}, nil, true},
	}

	// create files to test
	for _, tt := range create {
		name := filepath.Base(tt.F.Destination)
		t.Run("create-"+name, func(t *testing.T) {
			_, err = tt.F.Apply()
			assert.Equal(t, tt.Err, err)
		})
	}

	for _, tt := range create {
		name := filepath.Base(tt.F.Destination)
		t.Run("fileinfo-"+name, func(t *testing.T) {
			fi, err := os.Lstat(tt.F.Destination)
			require.NoError(t, err)
			actual := New()
			actual.Destination = tt.F.Destination
			GetFileInfo(actual, fi)
			assert.Equal(t, tt.F.State, actual.State)
			if tt.F.Type != TypeLink {
				assert.Equal(t, tt.F.Type, actual.Type)
			}
			if tt.F.Type != TypeSymlink {
				assert.Equal(t, os.FileMode(*tt.F.Mode).String(), os.FileMode(*actual.Mode).String())
			}
			assert.Equal(t, tt.F.UserInfo.Username, actual.UserInfo.Username)
			assert.Equal(t, tt.F.UserInfo.Uid, actual.UserInfo.Uid)
			assert.Equal(t, tt.F.GroupInfo.Name, actual.GroupInfo.Name)
			assert.Equal(t, tt.F.GroupInfo.Gid, actual.GroupInfo.Gid)
		})
	}
}

func mode(perms int) *uint32 {
	m := new(uint32)
	*m = uint32(perms)
	return m
}

func modetype(perms int, t Type) *uint32 {
	m := new(uint32)
	*m = ModeType(uint32(perms), t)
	return m
}
