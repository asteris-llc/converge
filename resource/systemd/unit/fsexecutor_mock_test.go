// Copyright © 2017 Asteris, LLC
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

package unit

import (
	"os"
	"strings"
	"time"

	"github.com/stretchr/testify/mock"
)

type walkFuncArgs struct {
	path string
	info os.FileInfo
	err  error
}

type mockFsExecutor struct {
	mock.Mock
	walkWith []walkFuncArgs
}

func (m *mockFsExecutor) EvalSymlinks(path string) (string, error) {
	args := m.Called(path)
	return args.String(0), args.Error(1)
}

func (m *mockFsExecutor) Walk(root string, f func(string, os.FileInfo, error) error) error {
	args := m.Called(root, f)
	for _, node := range m.walkWith {
		if !strings.HasPrefix(node.path, root) {
			continue
		}
		f(node.path, node.info, node.err)
	}
	return args.Error(0)
}

func newMockWithPaths(path ...string) *mockFsExecutor {
	var args []walkFuncArgs
	for _, p := range path {
		a := walkFuncArgs{
			path: p,
			err:  nil,
			info: mockFileInfo{p},
		}
		args = append(args, a)
	}
	return &mockFsExecutor{walkWith: args}
}

type mockFileInfo struct {
	path string
}

func (n mockFileInfo) Name() string       { s := strings.Split(n.path, "/"); return s[len(s)-1] }
func (n mockFileInfo) Size() int64        { return 0 }
func (n mockFileInfo) Mode() os.FileMode  { return 0777 }
func (n mockFileInfo) ModTime() time.Time { return time.Now() }
func (n mockFileInfo) IsDir() bool        { return false }
func (n mockFileInfo) Sys() interface{}   { return nil }
