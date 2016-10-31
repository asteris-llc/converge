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
	//    "github.com/asteris-llc/converge/helpers/comparsion"
	"github.com/asteris-llc/converge/helpers/fakerenderer"
	"github.com/asteris-llc/converge/resource/lvm/lowlevel"
	"github.com/asteris-llc/converge/resource/lvm/lv"
	"github.com/asteris-llc/converge/resource/lvm/testdata"
	"github.com/asteris-llc/converge/resource/lvm/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"testing"
)

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
	// FIXME: proper diffs
	//    comparsion.AssertDiff(t, status.Diffs(), "vg0", "<not exists>", "/dev/sda1")

	status, err = r.Apply()
	assert.NoError(t, err)
	me.AssertCalled(t, "Run", "lvcreate", []string{"-n", volname, "-L", "100G", "vg0"})
}
