package lowlevel_test

import (
	"github.com/asteris-llc/converge/resource/lvm/testdata"
	"github.com/asteris-llc/converge/resource/lvm/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestQueryPhysicalVolumes(t *testing.T) {
	lvm, e := testhelpers.MakeLvmWithMockExec()
	e.On("Read", "pvs", []string{"--nameprefix", "--noheadings", "--unquoted", "--units", "b", "-o", "pv_all,vg_name", "--separator", ";"}).Return(testdata.TESTDATA_PVS, nil)
	pvs, err := lvm.QueryPhysicalVolumes()
	require.NoError(t, err)
	require.Contains(t, pvs, "/dev/md127")
	pv := pvs["/dev/md127"]
	assert.Equal(t, "/dev/md127", pv.Name)
	assert.Equal(t, "vg0", pv.Group)
	assert.Equal(t, "/dev/md127", pv.Device)
}

func TestQueryVolumeGroups(t *testing.T) {
	lvm, e := testhelpers.MakeLvmWithMockExec()
	e.On("Read", "vgs", []string{"--nameprefix", "--noheadings", "--unquoted", "--units", "b", "-o", "all", "--separator", ";"}).Return(testdata.TESTDATA_PVS, nil)
	vgs, err := lvm.QueryVolumeGroups()
	require.NoError(t, err)
	require.Contains(t, vgs, "vg0")
}

func TestQueryLogicalVolume(t *testing.T) {
	lvm, e := testhelpers.MakeLvmWithMockExec()
	e.On("Read", "lvs", []string{"--nameprefix", "--noheadings", "--unquoted", "--units", "b", "-o", "all", "--separator", ";", "vg0"}).Return(testdata.TESTDATA_LVS, nil)
	lvs, err := lvm.QueryLogicalVolumes("vg0")
	require.NoError(t, err)
	require.Contains(t, lvs, "data")
}
