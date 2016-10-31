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
	"github.com/asteris-llc/converge/resource/lvm/fs"
	"github.com/asteris-llc/converge/resource/lvm/testdata"
	"github.com/asteris-llc/converge/resource/lvm/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"testing"
)

// TestCreateFilesystem is a full-blown test, using fake execution engine, to look
// which commands should be executed from given node.
//
// It covers only basic case, for detailed testing, tests with mock-LVM should be used
func TestCreateFilesystem(t *testing.T) {
	lvm, me := testhelpers.MakeLvmWithMockExec()
	me.LvsFirstCall = true
	me.On("Getuid").Return(0)
	me.On("Lookup", "mkfs.xfs").Return(nil)
	me.On("Read", "pvs", mock.Anything).Return(testdata.Pvs, nil)
	me.On("Read", "vgs", mock.Anything).Return(testdata.Vgs, nil)
	me.On("Read", "lvs", mock.Anything).Return(testdata.Lvs, nil)
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
	comparison.AssertDiff(t, status.Diffs(), "format", "<unformatted>", "xfs")

	status, err = r.Apply()
	require.NoError(t, err)
	me.AssertCalled(t, "Run", "mkfs", []string{"-t", "xfs", "/dev/mapper/vg0-data"})
}
