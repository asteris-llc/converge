package vg_test

import (
	"github.com/asteris-llc/converge/helpers/comparsion"
	"github.com/asteris-llc/converge/helpers/fakerenderer"
	"github.com/asteris-llc/converge/resource/lvm/testdata"
	"github.com/asteris-llc/converge/resource/lvm/testhelpers"
	"github.com/asteris-llc/converge/resource/lvm/vg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"testing"
)

func TestCreateVolumeFromSingleDevice(t *testing.T) {
	lvm, me := testhelpers.MakeLvmWithMockExec()
	me.On("Read", "pvs", mock.Anything).Return("", nil)
	me.On("Read", "vgs", mock.Anything).Return("", nil)
	me.On("Run", "vgcreate", []string{"vg0", "/dev/sda1"}).Return(nil)

	fr := fakerenderer.New()

	r := &vg.ResourceVG{Name: "vg0"}
	r.Setup(lvm, []string{"/dev/sda1"})
	status, err := r.Check(fr)
	assert.NoError(t, err)
	assert.True(t, status.HasChanges())
	comparsion.AssertDiff(t, status.Diffs(), "vg0", "<not exists>", "/dev/sda1")

	status, err = r.Apply()
	assert.NoError(t, err)
	me.AssertCalled(t, "Run", "vgcreate", []string{"vg0", "/dev/sda1"})
}

func TestCreateVolumeWhichIsInAnotherGroup(t *testing.T) {
	lvm, me := testhelpers.MakeLvmWithMockExec()
	me.On("Read", "pvs", mock.Anything).Return(testdata.TESTDATA_PVS, nil)
	me.On("Read", "vgs", mock.Anything).Return(testdata.TESTDATA_VGS, nil)

	fr := fakerenderer.New()

	r := &vg.ResourceVG{Name: "vg1"}
	r.Setup(lvm, []string{"/dev/md127"})
	_, err := r.Check(fr)
	assert.Error(t, err)
}

func TestCreateVolumeWhichAlreadyExists(t *testing.T) {
	lvm, me := testhelpers.MakeLvmWithMockExec()
	me.On("Read", "pvs", mock.Anything).Return(testdata.TESTDATA_PVS, nil)
	me.On("Read", "vgs", mock.Anything).Return(testdata.TESTDATA_VGS, nil)

	fr := fakerenderer.New()

	r := &vg.ResourceVG{Name: "vg0"}
	r.Setup(lvm, []string{"/dev/md127"})
	status, err := r.Check(fr)
	assert.NoError(t, err)
	assert.False(t, status.HasChanges())
}
