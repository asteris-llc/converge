package lv_test

import (
	//    "github.com/asteris-llc/converge/helpers/comparsion"
	"github.com/asteris-llc/converge/helpers/fakerenderer"
	"github.com/asteris-llc/converge/resource/lvm/lv"
	"github.com/asteris-llc/converge/resource/lvm/testdata"
	"github.com/asteris-llc/converge/resource/lvm/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"testing"
)

func TestCreateLogicalVolume(t *testing.T) {
	lvm, me := testhelpers.MakeLvmWithMockExec()
	me.On("Read", "pvs", mock.Anything).Return(testdata.TESTDATA_PVS, nil)
	me.On("Read", "vgs", mock.Anything).Return(testdata.TESTDATA_VGS, nil)
	me.On("Read", "lvs", mock.Anything).Return(testdata.TESTDATA_VGS, nil)
	me.On("Run", "lvcreate", []string{"-n", "test", "-L", "100G", "vg0"}).Return(nil)

	fr := fakerenderer.New()

	r := &lv.ResourceLV{}
	r.Setup(lvm, "vg0", "test", "100G")
	status, err := r.Check(fr)
	assert.NoError(t, err)
	assert.True(t, status.HasChanges())
	assert.Equal(t, "test", status.Value())
	// FIXME: proper diffs
	//    comparsion.AssertDiff(t, status.Diffs(), "vg0", "<not exists>", "/dev/sda1")

	status, err = r.Apply(fr)
	assert.NoError(t, err)
	me.AssertCalled(t, "Run", "lvcreate", []string{"-n", "test", "-L", "100G", "vg0"})
}
