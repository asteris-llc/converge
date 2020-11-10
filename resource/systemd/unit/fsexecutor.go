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
	"path/filepath"
)

type fsexecutor interface {
	EvalSymlinks(path string) (string, error)
	Walk(root string, f func(string, os.FileInfo, error) error) error
}

type realFsExecutor struct{}

func (r realFsExecutor) EvalSymlinks(path string) (string, error) {
	return filepath.EvalSymlinks(path)
}

func (r realFsExecutor) Walk(root string, f func(string, os.FileInfo, error) error) error {
	return filepath.Walk(root, f)
}

func filesystemExecutor() fsexecutor { return realFsExecutor{} }
