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

package fs_test

import (
	"github.com/asteris-llc/converge/helpers/comparison"
	"github.com/asteris-llc/converge/helpers/fakerenderer"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/lvm/fs"
	"github.com/asteris-llc/converge/resource/lvm/lowlevel"
	"github.com/asteris-llc/converge/resource/lvm/sampledata"
	"github.com/asteris-llc/converge/resource/lvm/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"

	"fmt"
	"testing"
)

// TestFSCheck tests Check() from filesystem resource
func TestFSCheck(t *testing.T) {
	t.Run("normal flow", func(t *testing.T) {
		lvm, m := testhelpers.MakeFakeLvm()
		setupNormalFlowCheck(m, "xfs", true, true)
		status, _ := simpleCheckSuccess(t, lvm)
		assert.True(t, status.HasChanges())
	})

	t.Run("missing tools failure", func(t *testing.T) {
		lvm, m := testhelpers.MakeFakeLvm()
		m.On("CheckFilesystemTools", mock.Anything).Return(fmt.Errorf("failure"))
		_ = simpleCheckFailure(t, lvm)
		m.AssertCalled(t, "CheckFilesystemTools", "xfs")
	})

	t.Run("Blkid() failure", func(t *testing.T) {
		lvm, m := testhelpers.MakeFakeLvm()
		m.On("CheckFilesystemTools", mock.Anything).Return(nil)
		m.On("Blkid", mock.Anything).Return("", fmt.Errorf("failure"))
		_ = simpleCheckFailure(t, lvm)
		m.AssertCalled(t, "Blkid", "/dev/mapper/vg0-data")
	})

	t.Run("Blkid() report, that filesystem already formatted with different fs", func(t *testing.T) {
		lvm, m := testhelpers.MakeFakeLvm()
		m.On("CheckFilesystemTools", mock.Anything).Return(nil)
		m.On("Blkid", mock.Anything).Return("ext4", nil)
		_ = simpleCheckFailure(t, lvm)
		m.AssertCalled(t, "Blkid", "/dev/mapper/vg0-data")
	})

	t.Run("CheckUnit() failure", func(t *testing.T) {
		lvm, m := testhelpers.MakeFakeLvm()
		m.On("CheckFilesystemTools", mock.Anything).Return(nil)
		m.On("Blkid", mock.Anything).Return("xfs", nil)
		m.On("CheckUnit", "/etc/systemd/system/mnt-data.mount", mock.Anything).Return(false, fmt.Errorf("failure"))
		_ = simpleCheckFailure(t, lvm)
		m.AssertCalled(t, "CheckUnit", "/etc/systemd/system/mnt-data.mount", mock.Anything)
	})

	t.Run("Mountpoint failure", func(t *testing.T) {
		lvm, m := testhelpers.MakeFakeLvm()
		m.On("CheckFilesystemTools", mock.Anything).Return(nil)
		m.On("Blkid", mock.Anything).Return("xfs", nil)
		m.On("CheckUnit", "/etc/systemd/system/mnt-data.mount", mock.Anything).Return(false, nil)
		m.On("Mountpoint", "/mnt/data").Return(false, fmt.Errorf("failure"))
		_ = simpleCheckFailure(t, lvm)
		m.AssertCalled(t, "Mountpoint", "/mnt/data")
	})

	t.Run("do nothing (no changes)", func(t *testing.T) {
		lvm, m := testhelpers.MakeFakeLvm()
		setupNormalFlowCheck(m, "xfs", false, true) // "xfs", no unit diffs, mount is mounted
		status, _ := simpleCheckSuccess(t, lvm)
		assert.False(t, status.HasChanges())
	})
}

// TestFSApply tests Apply() from filesystem resource
func TestFSApply(t *testing.T) {
	t.Run("normal flow", func(t *testing.T) {
		lvm, m := testhelpers.MakeFakeLvm()
		setupNormalFlowApply(m, "", true, true)
		_ = simpleApplySuccess(t, lvm)
		m.AssertCalled(t, "Mkfs", "/dev/mapper/vg0-data", "xfs")
		m.AssertCalled(t, "UpdateUnit", "/etc/systemd/system/mnt-data.mount", mock.Anything)
		m.AssertCalled(t, "StartUnit", "mnt-data.mount")
	})

	t.Run("no mkfs, only unit update and mount", func(t *testing.T) {
		lvm, m := testhelpers.MakeFakeLvm()
		setupNormalFlowApply(m, "xfs", true, false) // "xfs", no unit diffs, mount is NOT mounted
		_ = simpleApplySuccess(t, lvm)
		m.AssertNotCalled(t, "Mkfs", "/dev/mapper/vg0-data", "xfs")
		m.AssertCalled(t, "UpdateUnit", "/etc/systemd/system/mnt-data.mount", mock.Anything)
		// start unit is cascade action after UpdateUnit
		m.AssertCalled(t, "StartUnit", "mnt-data.mount")
	})

	t.Run("no mkfs, no unit, only mount", func(t *testing.T) {
		lvm, m := testhelpers.MakeFakeLvm()
		setupNormalFlowApply(m, "xfs", false, false)
		_ = simpleApplySuccess(t, lvm)
		m.AssertNotCalled(t, "Mkfs", "/dev/mapper/vg0-data", "xfs")
		m.AssertNotCalled(t, "UpdateUnit", "/etc/systemd/system/mnt-data.mount", mock.Anything)
		m.AssertCalled(t, "StartUnit", "mnt-data.mount")
	})

	t.Run("Mkfs() failure", func(t *testing.T) {
		lvm, m := testhelpers.MakeFakeLvm()
		setupNormalFlowCheck(m, "", false, false)
		m.On("Mkfs", mock.Anything, mock.Anything).Return(fmt.Errorf("failure"))
		_ = simpleApplyFailure(t, lvm)
		m.AssertCalled(t, "Mkfs", "/dev/mapper/vg0-data", "xfs")
	})

	t.Run("UpdateUnit() failure", func(t *testing.T) {
		lvm, m := testhelpers.MakeFakeLvm()
		setupNormalFlowCheck(m, "", true, true)
		m.On("Mkfs", mock.Anything, mock.Anything).Return(nil)
		m.On("UpdateUnit", mock.Anything, mock.Anything).Return(fmt.Errorf("failure"))
		_ = simpleApplyFailure(t, lvm)
		m.AssertCalled(t, "UpdateUnit", "/etc/systemd/system/mnt-data.mount", mock.Anything)
	})

	t.Run("StartUnit() failure", func(t *testing.T) {
		lvm, m := testhelpers.MakeFakeLvm()
		setupNormalFlowCheck(m, "", true, true)
		m.On("Mkfs", mock.Anything, mock.Anything).Return(nil)
		m.On("UpdateUnit", mock.Anything, mock.Anything).Return(nil)
		m.On("StartUnit", mock.Anything).Return(fmt.Errorf("failure"))
		_ = simpleApplyFailure(t, lvm)
		m.AssertCalled(t, "StartUnit", "mnt-data.mount")
	})
}

// TestCreateFilesystem is a full-blown test, using fake execution engine, to look
// which commands should be executed from given node.
//
// It covers only basic case, for detailed testing, tests with mock-LVM should be used
func TestCreateFilesystem(t *testing.T) {
	lvm, me := testhelpers.MakeLvmWithMockExec()
	me.LvsFirstCall = true
	me.On("Getuid").Return(0)
	me.On("Lookup", "mkfs.xfs").Return(nil)
	me.On("Read", "pvs", mock.Anything).Return(sampledata.Pvs, nil)
	me.On("Read", "vgs", mock.Anything).Return(sampledata.Vgs, nil)
	me.On("Read", "lvs", mock.Anything).Return(sampledata.Lvs, nil)
	me.On("ReadWithExitCode", "blkid", []string{"-c", "/dev/null", "-o", "value", "-s", "TYPE", "/dev/mapper/vg0-data"}).Return("", 0, nil)
	me.On("ReadFile", "/etc/systemd/system/mnt-data.mount").Return([]byte(""), nil)
	me.On("WriteFile", "/etc/systemd/system/mnt-data.mount", mock.Anything, mock.Anything).Return(nil)
	me.On("Exists", "/dev/mapper/vg0-data").Return(true, nil)
	me.On("Run", "mkfs", []string{"-t", "xfs", "/dev/mapper/vg0-data"}).Return(nil)
	me.On("RunWithExitCode", "mountpoint", []string{"-q", "/mnt/data"}).Return(1, nil)
	me.On("Run", "systemctl", []string{"daemon-reload"}).Return(nil)
	me.On("Run", "systemctl", []string{"start", "mnt-data.mount"}).Return(nil)

	fr := fakerenderer.New()

	mount := defaultMount()
	r, e := fs.NewResourceFS(lvm, mount)
	require.NoError(t, e)
	status, err := r.Check(context.Background(), fr)
	require.NoError(t, err)
	assert.True(t, status.HasChanges())
	comparison.AssertDiff(t, status.Diffs(), "format", "<unformatted>", "xfs")

	// Only basic actions checked here
	status, err = r.Apply(context.Background())
	require.NoError(t, err)
	me.AssertCalled(t, "Run", "mkfs", []string{"-t", "xfs", "/dev/mapper/vg0-data"})
	me.AssertCalled(t, "Run", "systemctl", []string{"start", "mnt-data.mount"})
}

func defaultMount() *fs.Mount {
	mount := &fs.Mount{
		What:  "/dev/mapper/vg0-data",
		Where: "/mnt/data",
		Type:  "xfs",
	}
	return mount
}

func simpleCheckSuccess(t *testing.T, lvm lowlevel.LVM) (resource.TaskStatus, resource.Task) {
	fr := fakerenderer.New()
	res, e := fs.NewResourceFS(lvm, defaultMount())
	require.NoError(t, e)
	status, err := res.Check(context.Background(), fr)
	assert.NoError(t, err)
	assert.NotNil(t, status)
	return status, res
}

func simpleCheckFailure(t *testing.T, lvm lowlevel.LVM) resource.TaskStatus {
	fr := fakerenderer.New()
	res, e := fs.NewResourceFS(lvm, defaultMount())
	require.NoError(t, e)
	status, err := res.Check(context.Background(), fr)
	assert.Error(t, err)
	assert.Nil(t, status)
	return status
}

func simpleApplySuccess(t *testing.T, lvm lowlevel.LVM) resource.TaskStatus {
	checkStatus, res := simpleCheckSuccess(t, lvm)
	require.True(t, checkStatus.HasChanges())
	status, err := res.Apply(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, status)
	return status
}

func simpleApplyFailure(t *testing.T, lvm lowlevel.LVM) resource.TaskStatus {
	checkStatus, res := simpleCheckSuccess(t, lvm)
	require.True(t, checkStatus.HasChanges())
	status, err := res.Apply(context.Background())
	assert.Error(t, err)
	assert.Nil(t, status)
	return status
}

func setupNormalFlowCheck(m *testhelpers.FakeLVM, blkid string, triggerUnit bool, triggerMountpoint bool) {
	m.On("CheckFilesystemTools", mock.Anything).Return(nil) // pretend that we compat with any FS
	m.On("Blkid", mock.Anything).Return(blkid, nil)
	m.On("CheckUnit", mock.Anything, mock.Anything).Return(triggerUnit, nil)
	m.On("Mountpoint", mock.Anything).Return(triggerMountpoint, nil)
}

func setupNormalFlowApply(m *testhelpers.FakeLVM, blkid string, triggerUnit bool, triggerMountpoint bool) {
	setupNormalFlowCheck(m, blkid, triggerUnit, triggerMountpoint)
	m.On("Mkfs", mock.Anything, mock.Anything).Return(nil)
	m.On("UpdateUnit", mock.Anything, mock.Anything).Return(nil)
	m.On("StartUnit", mock.Anything).Return(nil)
}
