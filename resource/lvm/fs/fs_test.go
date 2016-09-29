package fs_test

import (
	//    "github.com/asteris-llc/converge/helpers/comparsion"
	"github.com/asteris-llc/converge/helpers/fakerenderer"
	"github.com/asteris-llc/converge/resource/lvm/fs"
	"github.com/asteris-llc/converge/resource/lvm/testdata"
	"github.com/asteris-llc/converge/resource/lvm/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"testing"
)

func TestCreateFilesystem(t *testing.T) {
	lvm, me := testhelpers.MakeLvmWithMockExec()
	me.On("Read", "pvs", mock.Anything).Return(testdata.TESTDATA_PVS, nil)
	me.On("Read", "vgs", mock.Anything).Return(testdata.TESTDATA_VGS, nil)
	me.On("Read", "lvs", mock.Anything).Return(testdata.TESTDATA_LVS, nil)
	me.On("Read", "blkid", []string{"-c", "/dev/null", "-o", "value", "-s", "TYPE", "/dev/mapper/vg0-data"}).Return("", nil)
	me.On("ReadFile", "/etc/systemd/system/mnt-data.mount").Return([]byte(""), nil)
	me.On("WriteFile", "/etc/systemd/system/mnt-data.mount", mock.Anything, mock.Anything).Return(nil)
	me.On("Run", "mkfs", []string{"-t", "xfs", "/dev/mapper/vg0-data"}).Return(nil)
	me.On("RunExitCode", "mountpoint", []string{"-q", "/mnt/data"}).Return(1, nil)
	me.On("Run", "systemctl", []string{"daemon-reload"}).Return(nil)
	me.On("Run", "systemctl", []string{"start", "mnt-data.mount"}).Return(nil)

	fr := fakerenderer.New()

	r := &fs.ResourceFS{}
	mount := &fs.Mount{
		What:  "/dev/mapper/vg0-data",
		Where: "/mnt/data",
		Type:  "xfs",
	}
	r.Setup(lvm, mount)
	status, err := r.Check(fr)
	assert.NoError(t, err)
	assert.True(t, status.HasChanges())
	// FIXME: proper diffs
	//    comparsion.AssertDiff(t, status.Diffs(), "vg0", "<not exists>", "/dev/sda1")

	status, err = r.Apply()
	assert.NoError(t, err)
	me.AssertCalled(t, "Run", "mkfs", []string{"-t", "xfs", "/dev/mapper/vg0-data"})
}
