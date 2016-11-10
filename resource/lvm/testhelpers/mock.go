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

package testhelpers

import (
	"os"

	"github.com/asteris-llc/converge/resource/lvm/lowlevel"
	"github.com/stretchr/testify/mock"
)

// MockExecutor is a lowlevel.Exec impleentation for faking system interoperation
type MockExecutor struct {
	mock.Mock

	// NB: need more proper injection for `lvs` executing, return different output on different calls
	// Related issue: https://github.com/asteris-llc/converge/issues/456
	LvsFirstCall bool // ugly hack
}

// MakeLvmWithMockExec creates LVM backed with MockExecutor
func MakeLvmWithMockExec() (lowlevel.LVM, *MockExecutor) {
	me := &MockExecutor{}
	lvm := lowlevel.MakeRealLVM(me)
	return lvm, me
}

// Run is mock for Exec.Run()
func (mex *MockExecutor) Run(prog string, args []string) error {
	c := mex.Called(prog, args)
	return c.Error(0)
}

// RunWithExitCode is mock for Exec.RunWithExitCode()
func (mex *MockExecutor) RunWithExitCode(prog string, args []string) (int, error) {
	c := mex.Called(prog, args)
	return c.Int(0), c.Error(1)
}

// Read is mock for Exec.Read()
func (mex *MockExecutor) Read(prog string, args []string) (string, error) {
	if mex.LvsFirstCall {
		mex.LvsFirstCall = false
		return "", nil
	}
	c := mex.Called(prog, args)
	return c.String(0), c.Error(1)
}

// ReadWithExitCode is mock for Exec.ReadWithExitCode()
func (mex *MockExecutor) ReadWithExitCode(prog string, args []string) (string, int, error) {
	c := mex.Called(prog, args)
	return c.String(0), c.Int(1), c.Error(2)
}

// ReadFile is mock for Exec.ReadFile()
func (mex *MockExecutor) ReadFile(fn string) ([]byte, error) {
	c := mex.Called(fn)
	return c.Get(0).([]byte), c.Error(1)
}

// Lookup is mock for Exec.Lookup()
func (mex *MockExecutor) Lookup(prog string) error {
	return mex.Called(prog).Error(0)
}

// WriteFile is mock for Exec.WriteFile()
func (mex *MockExecutor) WriteFile(fn string, content []byte, perm os.FileMode) error {
	c := mex.Called(fn, content, perm)
	return c.Error(0)
}

// MkdirAll is mock for Exec.MkdirAll()
func (mex *MockExecutor) MkdirAll(path string, perm os.FileMode) error {
	return mex.Called(path, perm).Error(0)
}

// Exists is mock for Exec.Exists()
func (mex *MockExecutor) Exists(path string) (bool, error) {
	c := mex.Called(path)
	return c.Bool(0), c.Error(1)
}

// Getuid is mock for Getuid()
func (mex *MockExecutor) Getuid() int {
	return mex.Called().Int(0)
}

// EvalSymlinks mocks symlink evaluation
func (mex *MockExecutor) EvalSymlinks(s string) (string, error) {
	return s, nil
}
