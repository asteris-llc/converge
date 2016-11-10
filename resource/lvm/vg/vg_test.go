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

package vg_test

import (
	"github.com/asteris-llc/converge/helpers/comparison"
	"github.com/asteris-llc/converge/helpers/fakerenderer"
	"github.com/asteris-llc/converge/resource"
	"github.com/asteris-llc/converge/resource/lvm/lowlevel"
	"github.com/asteris-llc/converge/resource/lvm/testdata"
	"github.com/asteris-llc/converge/resource/lvm/testhelpers"
	"github.com/asteris-llc/converge/resource/lvm/vg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"

	"fmt"
	"testing"
)

// TestVGCheck is test for VG.Check
func TestVGCheck(t *testing.T) {
	t.Run("check prerequisites failure", func(t *testing.T) {
		lvm, m := testhelpers.MakeFakeLvm()
		m.On("Check").Return(fmt.Errorf("failed"))
		_ = simpleCheckFailure(t, lvm, "vg0", []string{"/dev/sda1"}, false, false)
	})

	t.Run("QueryVolumeGroups failure", func(t *testing.T) {
		lvm, m := testhelpers.MakeFakeLvm()
		m.On("Check").Return(nil)
		m.On("QueryVolumeGroups").Return(make(map[string]*lowlevel.VolumeGroup), nil)
		m.On("QueryPhysicalVolumes").Return(make(map[string]*lowlevel.PhysicalVolume), fmt.Errorf("failed"))
		_ = simpleCheckFailure(t, lvm, "vg0", []string{"/dev/sda1"}, false, false)
	})

	t.Run("QueryPhysicalVolumes failure", func(t *testing.T) {
		lvm, m := testhelpers.MakeFakeLvm()
		m.On("Check").Return(nil)
		m.On("QueryVolumeGroups").Return(make(map[string]*lowlevel.VolumeGroup), fmt.Errorf("failed"))
		m.On("QueryPhysicalVolumes").Return(make(map[string]*lowlevel.PhysicalVolume), nil)
		_ = simpleCheckFailure(t, lvm, "vg0", []string{"/dev/sda1"}, false, false)
	})

	t.Run("single device", func(t *testing.T) {
		lvm, _ := testhelpers.MakeFakeLvmEmpty()

		status, _ := simpleCheckSuccess(t, lvm, "vg0", []string{"/dev/sda1"}, false, false)
		assert.True(t, status.HasChanges())
		comparison.AssertDiff(t, status.Diffs(), "vg0", "<not exists>", "/dev/sda1")
	})

	t.Run("multiple devices", func(t *testing.T) {
		lvm, _ := testhelpers.MakeFakeLvmEmpty()

		status, _ := simpleCheckSuccess(t, lvm, "vg0", []string{"/dev/sda1", "/dev/sdb1", "/dev/sdc1"}, false, false)
		assert.True(t, status.HasChanges())
		comparison.AssertDiff(t, status.Diffs(), "vg0", "<not exists>", "/dev/sda1, /dev/sdb1, /dev/sdc1")
	})

	t.Run("device from another VG", func(t *testing.T) {
		lvm, _ := testhelpers.MakeFakeLvmNonEmpty()
		_ = simpleCheckFailure(t, lvm, "vg0", []string{"/dev/sda1", "/dev/sdb1", "/dev/sdc1"}, false, false)
	})

	t.Run("device remove", func(t *testing.T) {
		lvm, _ := testhelpers.MakeFakeLvmNonEmpty()
		status, _ := simpleCheckSuccess(t, lvm, "vg1", []string{"/dev/sdc1"}, true, false)
		assert.True(t, status.HasChanges())
		comparison.AssertDiff(t, status.Diffs(), "/dev/sdd1", "<group vg1>", "<removed>")

		// `/dev/sdc1` should remain intact
		assert.NotContains(t, "/dev/sdc1", status.Diffs())
	})

	t.Run("device force remove", func(t *testing.T) {
		lvm, _ := testhelpers.MakeFakeLvmNonEmpty()
		status, _ := simpleCheckSuccess(t, lvm, "vg1", []string{"/dev/sdc1"}, true, true)
		assert.True(t, status.HasChanges())
		comparison.AssertDiff(t, status.Diffs(), "/dev/sdd1", "<group vg1>", "<destructed>")

		// `/dev/sdc1` should remain intact
		assert.NotContains(t, "/dev/sdc1", status.Diffs())
	})

	t.Run("one device add", func(t *testing.T) {
		lvm, _ := testhelpers.MakeFakeLvmNonEmpty()
		status, _ := simpleCheckSuccess(t, lvm, "vg1", []string{"/dev/sdc1", "/dev/sdd1", "/dev/sde1"}, false, false)
		assert.True(t, status.HasChanges())
		comparison.AssertDiff(t, status.Diffs(), "/dev/sde1", "<no group>", "<group vg1>")

		// `/dev/sdc1` and `/dev/sdd1` should remain intact
		assert.NotContains(t, "/dev/sdc1", status.Diffs())
		assert.NotContains(t, "/dev/sdd1", status.Diffs())
	})

	t.Run("once device add, one remove", func(t *testing.T) {
		lvm, _ := testhelpers.MakeFakeLvmNonEmpty()
		status, _ := simpleCheckSuccess(t, lvm, "vg1", []string{"/dev/sdc1", "/dev/sde1"}, true, false)
		assert.True(t, status.HasChanges())
		comparison.AssertDiff(t, status.Diffs(), "/dev/sde1", "<no group>", "<group vg1>")
		comparison.AssertDiff(t, status.Diffs(), "/dev/sdd1", "<group vg1>", "<removed>")

		// `/dev/sdc1` should remain intact
		assert.NotContains(t, "/dev/sdc1", status.Diffs())
	})
}

// TestVGApply is test for VG.Apply()
func TestVGApply(t *testing.T) {
	t.Run("single device", func(t *testing.T) {
		lvm, m := testhelpers.MakeFakeLvmEmpty()
		m.On("CreateVolumeGroup", mock.Anything, mock.Anything).Return(nil)

		_ = simpleApplySuccess(t, lvm, "vg0", []string{"/dev/sda1"}, false, false)
		m.AssertCalled(t, "CreateVolumeGroup", "vg0", []string{"/dev/sda1"})
	})

	t.Run("multiple devices", func(t *testing.T) {
		lvm, m := testhelpers.MakeFakeLvmEmpty()
		m.On("CreateVolumeGroup", "vg0", []string{"/dev/sda1", "/dev/sdb1", "/dev/sdc1"}).Return(nil)

		_ = simpleApplySuccess(t, lvm, "vg0", []string{"/dev/sda1", "/dev/sdb1", "/dev/sdc1"}, false, false)
		m.AssertCalled(t, "CreateVolumeGroup", "vg0", []string{"/dev/sda1", "/dev/sdb1", "/dev/sdc1"})
	})

	t.Run("one device add", func(t *testing.T) {
		lvm, m := testhelpers.MakeFakeLvmNonEmpty()
		m.On("ExtendVolumeGroup", "vg1", mock.Anything).Return(nil)
		_ = simpleApplySuccess(t, lvm, "vg1", []string{"/dev/sdc1", "/dev/sdd1", "/dev/sde1", "/dev/sda"}, false, false)
		m.AssertCalled(t, "ExtendVolumeGroup", "vg1", "/dev/sde1")
		m.AssertCalled(t, "ExtendVolumeGroup", "vg1", "/dev/sda")
	})

	t.Run("device remove", func(t *testing.T) {
		lvm, m := testhelpers.MakeFakeLvmNonEmpty()
		m.On("ReduceVolumeGroup", "vg1", mock.Anything).Return(nil)
		m.On("RemovePhysicalVolume", mock.Anything, mock.Anything).Return(nil)
		_ = simpleApplySuccess(t, lvm, "vg1", []string{"/dev/sdc1"}, true, false)
		m.AssertCalled(t, "ReduceVolumeGroup", "vg1", "/dev/sdd1")
		m.AssertNotCalled(t, "RemovePhysicalVolume", mock.Anything)
	})

	t.Run("device force remove", func(t *testing.T) {
		lvm, m := testhelpers.MakeFakeLvmNonEmpty()
		m.On("ReduceVolumeGroup", "vg1", mock.Anything).Return(nil)
		m.On("RemovePhysicalVolume", mock.Anything, mock.Anything).Return(nil)
		_ = simpleApplySuccess(t, lvm, "vg1", []string{"/dev/sdc1"}, true, true)
		m.AssertCalled(t, "ReduceVolumeGroup", "vg1", "/dev/sdd1")
		m.AssertCalled(t, "RemovePhysicalVolume", "/dev/sdd1", mock.Anything)
	})

	t.Run("CreatePhysicalVolume failure", func(t *testing.T) {
		lvm, m := testhelpers.MakeFakeLvmEmpty()
		m.On("CreateVolumeGroup", mock.Anything, mock.Anything).Return(fmt.Errorf("failed"))

		_ = simpleApplyFailure(t, lvm, "vg0", []string{"/dev/sda1"}, false, false)
		m.AssertCalled(t, "CreateVolumeGroup", "vg0", []string{"/dev/sda1"})
	})
	t.Run("ExtendVolumeGroup failure", func(t *testing.T) {
		lvm, m := testhelpers.MakeFakeLvmNonEmpty()
		m.On("ExtendVolumeGroup", "vg1", mock.Anything).Return(fmt.Errorf("failed"))
		_ = simpleApplyFailure(t, lvm, "vg1", []string{"/dev/sdc1", "/dev/sdd1", "/dev/sde1", "/dev/sda"}, false, false)
		m.AssertCalled(t, "ExtendVolumeGroup", "vg1", mock.Anything)
	})
	t.Run("ReduceVolumeGroup failure", func(t *testing.T) {
		lvm, m := testhelpers.MakeFakeLvmNonEmpty()
		m.On("ReduceVolumeGroup", "vg1", mock.Anything).Return(fmt.Errorf("failure"))
		m.On("RemovePhysicalVolume", mock.Anything, mock.Anything).Return(nil)
		_ = simpleApplyFailure(t, lvm, "vg1", []string{"/dev/sdc1"}, true, true)
		// also check that "RemovePhysicalVolume" not called, if ReduceVolumeGroup failed
		m.AssertNotCalled(t, "RemovePhysicalVolume", mock.Anything)
	})
	t.Run("RemovePhysicalVolume failure", func(t *testing.T) {
		lvm, m := testhelpers.MakeFakeLvmNonEmpty()
		m.On("ReduceVolumeGroup", "vg1", mock.Anything).Return(nil)
		m.On("RemovePhysicalVolume", mock.Anything, mock.Anything).Return(fmt.Errorf("failure"))
		_ = simpleApplyFailure(t, lvm, "vg1", []string{"/dev/sdc1"}, true, true)
		m.AssertCalled(t, "RemovePhysicalVolume", "/dev/sdd1", mock.Anything)
	})
}

// TestCreateVolume is a full-blown test, using fake engine to trace from high-level
// graph node vg.resourceVG, to commands passed to LVM tools. It cover only straighforward
// cases. Use mock-LVM for real tests of highlevel stuff.
func TestCreateVolume(t *testing.T) {
	t.Run("single device", func(t *testing.T) {
		lvm, me := testhelpers.MakeLvmWithMockExec()

		me.On("Getuid").Return(0)                  // assume, that we have root
		me.On("Lookup", mock.Anything).Return(nil) // assume, that all tools are here

		me.On("Read", "pvs", mock.Anything).Return("", nil)
		me.On("Read", "vgs", mock.Anything).Return("", nil)

		me.On("Run", "vgcreate", []string{"vg0", "/dev/sda1"}).Return(nil)

		fr := fakerenderer.New()

		r := vg.NewResourceVG(lvm, "vg0", []string{"/dev/sda1"}, false, false)
		status, err := r.Check(context.Background(), fr)
		require.NoError(t, err)
		assert.True(t, status.HasChanges())
		comparison.AssertDiff(t, status.Diffs(), "vg0", "<not exists>", "/dev/sda1")

		status, err = r.Apply(context.Background())
		assert.NoError(t, err)
		me.AssertCalled(t, "Run", "vgcreate", []string{"vg0", "/dev/sda1"})
	})

	t.Run("multiple devices", func(t *testing.T) {
		lvm, me := testhelpers.MakeLvmWithMockExec()

		me.On("Getuid").Return(0)                  // assume, that we have root
		me.On("Lookup", mock.Anything).Return(nil) // assume, that all tools are here

		me.On("Read", "pvs", mock.Anything).Return(testdata.Pvs, nil)
		me.On("Read", "vgs", mock.Anything).Return(testdata.Vgs, nil)

		fr := fakerenderer.New()

		r := vg.NewResourceVG(lvm, "vg1", []string{"/dev/md127"}, false, false)
		_, err := r.Check(context.Background(), fr)
		assert.Error(t, err)
	})

	t.Run("volume which already exists", func(t *testing.T) {
		lvm, me := testhelpers.MakeLvmWithMockExec()

		me.On("Getuid").Return(0)                  // assume, that we have root
		me.On("Lookup", mock.Anything).Return(nil) // assume, that all tools are here

		me.On("Read", "pvs", mock.Anything).Return(testdata.Pvs, nil)
		me.On("Read", "vgs", mock.Anything).Return(testdata.Vgs, nil)

		fr := fakerenderer.New()

		r := vg.NewResourceVG(lvm, "vg0", []string{"/dev/md127"}, false, false)
		status, err := r.Check(context.Background(), fr)
		assert.NoError(t, err)
		assert.False(t, status.HasChanges())
	})
}

func simpleCheckFailure(t *testing.T, lvm lowlevel.LVM, group string, devs []string, remove bool, forceRemove bool) resource.TaskStatus {
	r := vg.NewResourceVG(lvm, group, devs, remove, forceRemove)
	status, err := r.Check(context.Background(), fakerenderer.New())
	assert.Error(t, err)
	return status
}

func simpleCheckSuccess(t *testing.T, lvm lowlevel.LVM, group string, devs []string, remove bool, forceRemove bool) (resource.TaskStatus, resource.Task) {
	r := vg.NewResourceVG(lvm, group, devs, remove, forceRemove)
	status, err := r.Check(context.Background(), fakerenderer.New())
	require.NoError(t, err)
	require.NotNil(t, status)
	return status, r
}

func simpleApplySuccess(t *testing.T, lvm lowlevel.LVM, group string, devs []string, remove bool, forceRemove bool) resource.TaskStatus {
	checkStatus, vg := simpleCheckSuccess(t, lvm, group, devs, remove, forceRemove)
	require.True(t, checkStatus.HasChanges())

	status, err := vg.Apply(context.Background())
	require.NoError(t, err)
	return status
}

func simpleApplyFailure(t *testing.T, lvm lowlevel.LVM, group string, devs []string, remove bool, forceRemove bool) resource.TaskStatus {
	checkStatus, vg := simpleCheckSuccess(t, lvm, group, devs, remove, forceRemove)
	require.True(t, checkStatus.HasChanges())

	status, err := vg.Apply(context.Background())
	require.Error(t, err)
	return status
}
