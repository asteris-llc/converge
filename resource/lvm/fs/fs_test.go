package fs_test

import (
	//    "github.com/asteris-llc/converge/helpers/comparison"
	"github.com/asteris-llc/converge/helpers/fakerenderer"
	"github.com/asteris-llc/converge/resource/lvm/fs"
	"github.com/asteris-llc/converge/resource/lvm/testdata"
	"github.com/asteris-llc/converge/resource/lvm/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"testing"
)

func TestCreateFilesystem(t *testing.T) {
	lvm, me := testhelpers.MakeLvmWithMockExec()
	me.LvsFirstCall = true
	me.On("Getuid").Return(0)
	me.On("Lookup", "mkfs.xfs").Return(nil)
	me.On("Read", "pvs", mock.Anything).Return(testdata.TESTDATA_PVS, nil)
	me.On("Read", "vgs", mock.Anything).Return(testdata.TESTDATA_VGS, nil)
	me.On("Read", "lvs", mock.Anything).Return(testdata.TESTDATA_LVS, nil)
	me.On("ReadWithExitCode", "blkid", []string{"-c", "/dev/null", "-o", "value", "-s", "TYPE", "/dev/mapper/vg0-data"}).Return("", 0, nil)
	me.On("ReadFile", "/etc/systemd/system/mnt-data.mount").Return([]byte(""), nil)
	me.On("WriteFile", "/etc/systemd/system/mnt-data.mount", mock.Anything, mock.Anything).Return(nil)
	me.On("Run", "mkfs", []string{"-t", "xfs", "/dev/mapper/vg0-data"}).Return(nil)
	me.On("RunWithExitCode", "mountpoint", []string{"-q", "/mnt/data"}).Return(1, nil)
	me.On("Run", "systemctl", []string{"daemon-reload"}).Return(nil)
	me.On("Run", "systemctl", []string{"start", "mnt-data.mount"}).Return(nil)

	fr := fakerenderer.New()

	mount := &fs.Mount{
		What:  "/dev/mapper/vg0-data",
		Where: "/mnt/data",
		Type:  "xfs",
	}
	r, e := fs.NewResourceFS(lvm, mount)
	require.NoError(t, e)
	status, err := r.Check(fr)
	require.NoError(t, err)
	assert.True(t, status.HasChanges())
	// FIXME: proper diffs
	//    comparison.AssertDiff(t, status.Diffs(), "vg0", "<not exists>", "/dev/sda1")

	status, err = r.Apply()
	require.NoError(t, err)
	me.AssertCalled(t, "Run", "mkfs", []string{"-t", "xfs", "/dev/mapper/vg0-data"})
}
