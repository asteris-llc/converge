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

package testhelpers

import (
	"github.com/asteris-llc/converge/resource/lvm/lowlevel"
	"github.com/stretchr/testify/mock"
)

// MakeFakeLvmEmpty create fake LVM for test injections
// This one created empty
func MakeFakeLvmEmpty() (lowlevel.LVM, *FakeLVM) {
	lvm, m := MakeFakeLvm()
	m.On("Check").Return(nil)
	m.On("QueryVolumeGroups").Return(make(map[string]*lowlevel.VolumeGroup), nil)
	m.On("QueryPhysicalVolumes").Return(make(map[string]*lowlevel.PhysicalVolume), nil)
	m.On("QueryLogicalVolumes", mock.Anything).Return(make(map[string]*lowlevel.LogicalVolume), nil)
	return lvm, m
}

// MakeFakeLvmEmpty create fake LVM for test injections
// This one created non-empty, it have group `vg`, consisting of
// two devices -- /dev/sdc1 and /dev/sdd1
func MakeFakeLvmNonEmpty() (lowlevel.LVM, *FakeLVM) {
	lvm, m := MakeFakeLvm()
	m.On("Check").Return(nil)
	m.On("QueryVolumeGroups").Return(map[string]*lowlevel.VolumeGroup{
		"vg1": &lowlevel.VolumeGroup{Name: "vg1"},
	}, nil)
	m.On("QueryPhysicalVolumes").Return(map[string]*lowlevel.PhysicalVolume{
		"/dev/sdc1": &lowlevel.PhysicalVolume{
			Name:  "/dev/sdc1",
			Group: "vg1",
		},
		"/dev/sdd1": &lowlevel.PhysicalVolume{
			Name:  "/dev/sdd1",
			Group: "vg1",
		},
	}, nil)
	return lvm, m
}
