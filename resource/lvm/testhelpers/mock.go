package testhelpers

import (
	"github.com/asteris-llc/converge/resource/lvm/lowlevel"
	"github.com/stretchr/testify/mock"
	"os"
)

// MockExecutor is a lowlevel.Exec impleentation for faking system interoperation
type MockExecutor struct {
	mock.Mock

	// FIXME: need more proper injection for `lvs` executing, return different output on different calls
	LvsFirstCall bool // ugly hack
}

func MakeLvmWithMockExec() (lowlevel.LVM, *MockExecutor) {
	me := &MockExecutor{}
	lvm := lowlevel.MakeRealLVM(me)
	return lvm, me
}

func (me *MockExecutor) Run(prog string, args []string) error {
	c := me.Called(prog, args)
	return c.Error(0)
}

func (e *MockExecutor) RunWithExitCode(prog string, args []string) (int, error) {
	c := e.Called(prog, args)
	return c.Int(0), c.Error(1)
}

func (me *MockExecutor) Read(prog string, args []string) (string, error) {
	if me.LvsFirstCall {
		me.LvsFirstCall = false
		return "", nil
	}
	c := me.Called(prog, args)
	return c.String(0), c.Error(1)
}

func (me *MockExecutor) ReadWithExitCode(prog string, args []string) (string, int, error) {
	c := me.Called(prog, args)
	return c.String(0), c.Int(1), c.Error(2)
}

func (me *MockExecutor) ReadFile(fn string) ([]byte, error) {
	c := me.Called(fn)
	return c.Get(0).([]byte), c.Error(1)
}

func (me *MockExecutor) Lookup(prog string) error {
	return me.Called(prog).Error(0)
}

func (me *MockExecutor) WriteFile(fn string, content []byte, perm os.FileMode) error {
	c := me.Called(fn, content, perm)
	return c.Error(0)
}

func (me *MockExecutor) MkdirAll(path string, perm os.FileMode) error {
	return me.Called(path, perm).Error(0)
}

func (me *MockExecutor) Exists(path string) (bool, error) {
	c := me.Called(path)
	return c.Bool(0), c.Error(1)
}

func (me *MockExecutor) Getuid() int {
	return me.Called().Int(0)
}
