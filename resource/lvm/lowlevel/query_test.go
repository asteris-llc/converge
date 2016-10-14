package lowlevel_test

import (
	"fmt"
	"github.com/asteris-llc/converge/resource/lvm/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestBlkid(t *testing.T) {
	lvm, e := testhelpers.MakeLvmWithMockExec()
	expected := []string{"-c", "/dev/null", "-o", "value", "-s", "TYPE", "/dev/sda1"}
	e.On("ReadWithExitCode", "blkid", expected).Return("xfs", 0, nil)
	fs, err := lvm.Blkid("/dev/sda1")
	assert.Equal(t, "xfs", fs)
	assert.NoError(t, err)
	e.AssertCalled(t, "ReadWithExitCode", "blkid", expected)
}

func TestBlkidError(t *testing.T) {
	lvm, e := testhelpers.MakeLvmWithMockExec()
	e.On("ReadWithExitCode", "blkid", mock.Anything).Return("", 0, fmt.Errorf("failed"))
	_, err := lvm.Blkid("/dev/sda1")
	assert.Error(t, err)
}

func TestBlkidNonzero(t *testing.T) {
	lvm, e := testhelpers.MakeLvmWithMockExec()
	e.On("ReadWithExitCode", "blkid", mock.Anything).Return("", 2, nil)
	fs, err := lvm.Blkid("/dev/sda1")
	assert.NoError(t, err)
	assert.Equal(t, "", fs)
}

func TestQueryParseEmptyString(t *testing.T) {
	lvm, e := testhelpers.MakeLvmWithMockExec()
	e.On("Read", "pvs", mock.Anything).Return("", nil)
	// FIXME: .Query() is not exported in interface, so use QueryPhysicalVolumes() which go
	pvs, err := lvm.QueryPhysicalVolumes()
	assert.NoError(t, err)
	assert.Empty(t, pvs)
}
