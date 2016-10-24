package lowlevel_test

import (
	"fmt"
	"testing"

	"github.com/asteris-llc/converge/resource/lvm/testdata"
	"github.com/asteris-llc/converge/resource/lvm/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// TestLVMBlkid tests LVM.Blkid()
func TestLVMBlkid(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		lvm, e := testhelpers.MakeLvmWithMockExec()
		expected := []string{"-c", "/dev/null", "-o", "value", "-s", "TYPE", "/dev/sda1"}
		e.On("ReadWithExitCode", "blkid", expected).Return("xfs", 0, nil)
		fs, err := lvm.Blkid("/dev/sda1")
		assert.Equal(t, "xfs", fs)
		assert.NoError(t, err)
		e.AssertCalled(t, "ReadWithExitCode", "blkid", expected)
	})

	t.Run("error during blkid call", func(t *testing.T) {
		lvm, e := testhelpers.MakeLvmWithMockExec()
		e.On("ReadWithExitCode", "blkid", mock.Anything).Return("", 0, fmt.Errorf("failed"))
		_, err := lvm.Blkid("/dev/sda1")
		assert.Error(t, err)
	})

	t.Run("blkid exit with nonzero (missing device)", func(t *testing.T) {
		lvm, e := testhelpers.MakeLvmWithMockExec()
		e.On("ReadWithExitCode", "blkid", mock.Anything).Return("", 2, nil)
		fs, err := lvm.Blkid("/dev/sda1")
		assert.NoError(t, err)
		assert.Equal(t, "", fs)
	})
}

// TestLVMQuery tests LVM.QueryXXXX()
func TestLVMQuery(t *testing.T) {
	t.Run("physical volumes", func(t *testing.T) {
		lvm, e := testhelpers.MakeLvmWithMockExec()
		e.On("Read", "pvs", []string{"--nameprefix", "--noheadings", "--unquoted", "--units", "b", "-o", "pv_all,vg_name", "--separator", ";"}).Return(testdata.Pvs, nil)
		pvs, err := lvm.QueryPhysicalVolumes()
		require.NoError(t, err)
		require.Contains(t, pvs, "/dev/md127")
		pv := pvs["/dev/md127"]
		assert.Equal(t, "/dev/md127", pv.Name)
		assert.Equal(t, "vg0", pv.Group)
		assert.Equal(t, "/dev/md127", pv.Device)
	})

	t.Run("volume groups", func(t *testing.T) {
		lvm, e := testhelpers.MakeLvmWithMockExec()
		e.On("Read", "vgs", []string{"--nameprefix", "--noheadings", "--unquoted", "--units", "b", "-o", "all", "--separator", ";"}).Return(testdata.Vgs, nil)
		vgs, err := lvm.QueryVolumeGroups()
		require.NoError(t, err)
		require.Contains(t, vgs, "vg0")
	})

	t.Run("logical volume", func(t *testing.T) {
		lvm, e := testhelpers.MakeLvmWithMockExec()
		e.On("Read", "lvs", []string{"--nameprefix", "--noheadings", "--unquoted", "--units", "b", "-o", "all", "--separator", ";", "vg0"}).Return(testdata.Lvs, nil)
		lvs, err := lvm.QueryLogicalVolumes("vg0")
		require.NoError(t, err)
		require.Contains(t, lvs, "data")
	})

	// TestQueryParseEmptyString test for LVM.Query{Physical,Logical}Volumes and .VolumeGroups() with empty command output
	// .query() is not exported in interface, so use QueryPhysicalVolumes() which call it under the hood.
	t.Run("parse empty string", func(t *testing.T) {
		lvm, e := testhelpers.MakeLvmWithMockExec()
		e.On("Read", "pvs", mock.Anything).Return("", nil)
		pvs, err := lvm.QueryPhysicalVolumes()
		assert.NoError(t, err)
		assert.Empty(t, pvs)
	})
}
