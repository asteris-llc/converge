package lowlevel_test

import (
	"fmt"
	"github.com/asteris-llc/converge/resource/lvm/lowlevel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

type MockExecutor struct {
	mock.Mock
}

func (me *MockExecutor) Run(prog string, args []string) error {
	c := me.Called(prog, args)
	return c.Error(0)
}

func (e *MockExecutor) RunExitCode(prog string, args []string) (int, error) {
	c := e.Called(prog, args)
	return c.Int(0), c.Error(1)
}

func (me *MockExecutor) Read(prog string, args []string) (string, error) {
	c := me.Called(prog, args)
	return c.String(0), c.Error(1)
}

func TestBlkid(t *testing.T) {
	e := &MockExecutor{}
	e.On("Read", "blkid", mock.Anything).Return("xfs", nil)
	lvm := &lowlevel.LVM{Backend: e}
	fs, err := lvm.Blkid("/dev/sda1")
	assert.Equal(t, "xfs", fs)
	assert.NoError(t, err)
	e.AssertCalled(t, "Read", "blkid", []string{"-c", "/dev/null", "-o", "value", "-s", "TYPE", "/dev/sda1"})
}

func TestBlkidError(t *testing.T) {
	e := &MockExecutor{}
	e.On("Read", "blkid", mock.Anything).Return("", fmt.Errorf("failed"))
	lvm := &lowlevel.LVM{Backend: e}
	_, err := lvm.Blkid("/dev/sda1")
	assert.Error(t, err)
}
