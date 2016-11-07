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

package lv_test

import (
	"github.com/asteris-llc/converge/helpers/comparison"
	"github.com/asteris-llc/converge/helpers/fakerenderer"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/lvm/lowlevel"
	"github.com/asteris-llc/converge/resource/lvm/lv"
	"github.com/asteris-llc/converge/resource/lvm/testdata"
	"github.com/asteris-llc/converge/resource/lvm/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"fmt"
	"testing"
)

// TestLVCheck testing Check() for LV resource
func TestLVCheck(t *testing.T) {
	t.Run("create volume", func(t *testing.T) {
		lvm, m := testhelpers.MakeFakeLvmEmpty()
		m.On("Check").Return(nil)
		m.On("CreateLogicalVolume", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("failure"))
		status, _ := simpleCheckSuccess(t, lvm, "vg0", "data", simpleSize(t, "100G"))
		comparison.AssertDiff(t, status.Diffs(), "data", "<not exists>", "created /dev/mapper/vg0-data")
	})

	t.Run("missing tools", func(t *testing.T) {
		lvm, m := testhelpers.MakeFakeLvm()
		m.On("Check").Return(fmt.Errorf("failure"))
		_ = simpleCheckFailure(t, lvm, "vg0", "data", simpleSize(t, "100G"))
	})
}

// TestLVApply tests Apply() for LV resource
func TestLVApply(t *testing.T) {
	t.Run("create volume", func(t *testing.T) {
		lvm, m := testhelpers.MakeFakeLvmNonEmpty()
		m.On("Check").Return(nil)
		m.On("CreateLogicalVolume", mock.Anything, mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
			// Put data about fresh "created" volume to `lvs`, to allow query device path from engine
			m.LvsOutput = map[string]*lowlevel.LogicalVolume{
				"data": &lowlevel.LogicalVolume{
					Name:       "data",
					DevicePath: "/dev/mapper/vg1-data",
				},
			}
		})
		m.On("WaitForDevice", mock.Anything).Return(nil)
		_ = simpleApplySuccess(t, lvm, "vg1", "data", simpleSize(t, "100G"))
		m.AssertCalled(t, "CreateLogicalVolume", "vg1", "data", simpleSize(t, "100G"))
		m.AssertCalled(t, "WaitForDevice", "/dev/mapper/vg1-data")
	})

	t.Run("create volume: unable query device", func(t *testing.T) {
		lvm, m := testhelpers.MakeFakeLvmNonEmpty()
		m.On("Check").Return(nil)
		m.On("CreateLogicalVolume", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		m.On("WaitForDevice", mock.Anything).Return(nil)
		// fail, due `lvs` still unchanged, and engine fails to qyert device paths
		// (see test above for success varian)
		_ = simpleApplyFailure(t, lvm, "vg1", "data", simpleSize(t, "100G"))
		m.AssertCalled(t, "CreateLogicalVolume", "vg1", "data", simpleSize(t, "100G"))
	})

	t.Run("CreateLogicalVolume failure", func(t *testing.T) {
		lvm, m := testhelpers.MakeFakeLvmNonEmpty()
		m.On("Check").Return(nil)
		m.On("CreateLogicalVolume", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("failure"))
		_ = simpleApplyFailure(t, lvm, "vg1", "data", simpleSize(t, "100G"))
		m.AssertCalled(t, "CreateLogicalVolume", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("WaitForDevice failure", func(t *testing.T) {
		lvm, m := testhelpers.MakeFakeLvmNonEmpty()
		m.On("Check").Return(nil)
		m.On("CreateLogicalVolume", mock.Anything, mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
			// Put data about fresh "created" volume to `lvs`, to allow query device path from engine
			m.LvsOutput = map[string]*lowlevel.LogicalVolume{
				"data": &lowlevel.LogicalVolume{
					Name:       "data",
					DevicePath: "/dev/mapper/vg1-data",
				},
			}
		})
		m.On("WaitForDevice", mock.Anything).Return(fmt.Errorf("failure"))
		_ = simpleApplyFailure(t, lvm, "vg1", "data", simpleSize(t, "100G"))
		m.AssertCalled(t, "CreateLogicalVolume", "vg1", "data", simpleSize(t, "100G"))
		m.AssertCalled(t, "WaitForDevice", "/dev/mapper/vg1-data")
	})
}

// TestCreateLogicalVolume is a full-blown integration test based on fake exec engine
// it call highlevel functions, and check how it call underlying lvm' commands
// only simple successful case tracked here, use mock LVM for all high level testing
func TestCreateLogicalVolume(t *testing.T) {
	volname := "data" // Match with existing name in testdata.Lvs, so fool engine to find proper paths, etc
	// after creation
	lvm, me := testhelpers.MakeLvmWithMockExec()
	me.LvsFirstCall = true
	me.On("Getuid").Return(0)                  // assume, that we have root
	me.On("Lookup", mock.Anything).Return(nil) // assume, that all tools are here

	me.On("Read", "pvs", mock.Anything).Return(testdata.Pvs, nil)
	me.On("Read", "vgs", mock.Anything).Return(testdata.Vgs, nil)
	me.On("Read", "lvs", mock.Anything).Return(testdata.Lvs, nil)
	me.On("Run", "lvcreate", []string{"-n", volname, "-L", "100G", "vg0"}).Return(nil)
	me.On("Exists", "/dev/mapper/vg0-data").Return(true, nil)

	fr := fakerenderer.New()

	size, sizeErr := lowlevel.ParseSize("100G")
	require.NoError(t, sizeErr)

	r := lv.NewResourceLV(lvm, "vg0", volname, size)
	status, err := r.Check(fr)
	assert.NoError(t, err)
	assert.True(t, status.HasChanges())
	comparison.AssertDiff(t, status.Diffs(), "data", "<not exists>", "created /dev/mapper/vg0-data")

	status, err = r.Apply()
	assert.NoError(t, err)
	me.AssertCalled(t, "Run", "lvcreate", []string{"-n", volname, "-L", "100G", "vg0"})
}

func simpleSize(t *testing.T, sizeStr string) *lowlevel.LvmSize {
	size, err := lowlevel.ParseSize(sizeStr)
	require.NoError(t, err)
	return size
}

func simpleCheckSuccess(t *testing.T, lvm lowlevel.LVM, group string, name string, size *lowlevel.LvmSize) (resource.TaskStatus, resource.Task) {
	fr := fakerenderer.New()
	res := lv.NewResourceLV(lvm, group, name, size)
	status, err := res.Check(fr)
	assert.NoError(t, err)
	assert.NotNil(t, status)
	return status, res
}

func simpleCheckFailure(t *testing.T, lvm lowlevel.LVM, group string, name string, size *lowlevel.LvmSize) resource.TaskStatus {
	fr := fakerenderer.New()
	res := lv.NewResourceLV(lvm, group, name, size)
	status, err := res.Check(fr)
	assert.Error(t, err)
	return status
}

func simpleApplySuccess(t *testing.T, lvm lowlevel.LVM, group string, name string, size *lowlevel.LvmSize) resource.TaskStatus {
	checkStatus, res := simpleCheckSuccess(t, lvm, group, name, size)
	require.True(t, checkStatus.HasChanges())
	status, err := res.Apply()
	require.NoError(t, err)
	return status
}

func simpleApplyFailure(t *testing.T, lvm lowlevel.LVM, group string, name string, size *lowlevel.LvmSize) resource.TaskStatus {
	checkStatus, res := simpleCheckSuccess(t, lvm, group, name, size)
	require.True(t, checkStatus.HasChanges())
	status, err := res.Apply()
	assert.Error(t, err)
	return status
}
