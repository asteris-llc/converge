package lowlevel_test

import (
	"fmt"
	"github.com/asteris-llc/converge/resource/lvm/lowlevel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestBlkid(t *testing.T) {
	e := &MockExecutor{}
	expected := []string{"-c", "/dev/null", "-o", "value", "-s", "TYPE", "/dev/sda1"}
	e.On("Read", "blkid", expected).Return("xfs", nil)
	lvm := &lowlevel.RealLVM{Backend: e}
	fs, err := lvm.Blkid("/dev/sda1")
	assert.Equal(t, "xfs", fs)
	assert.NoError(t, err)
	e.AssertCalled(t, "Read", "blkid", expected)
}

func TestBlkidError(t *testing.T) {
	e := &MockExecutor{}
	e.On("Read", "blkid", mock.Anything).Return("", fmt.Errorf("failed"))
	lvm := &lowlevel.RealLVM{Backend: e}
	_, err := lvm.Blkid("/dev/sda1")
	assert.Error(t, err)
}
