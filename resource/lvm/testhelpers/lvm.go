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

// FakeLVM is mock object implementing lowlevel.LVM
type FakeLVM struct {
	mock.Mock
}

// MakeFakeLvm create fake LVM for test injections
func MakeFakeLvm() (lowlevel.LVM, *FakeLVM) {
	lvm := &FakeLVM{}
	return lvm, lvm
}

// Check is mock for LVM.Check()
func (f *FakeLVM) Check() error {
	return f.Called().Error(0)
}

// CheckFilesystemTools is mock for LVM.CheckFilesystemTools()
func (f *FakeLVM) CheckFilesystemTools(fstype string) error {
	return f.Called(fstype).Error(0)
}

// QueryLogicalVolumes is mock for LVM.QueryLogicalVolumes
func (f *FakeLVM) QueryLogicalVolumes(vg string) (map[string]*lowlevel.LogicalVolume, error) {
	c := f.Called(vg)
	return c.Get(0).(map[string]*lowlevel.LogicalVolume), c.Error(1)
}

// QueryPhysicalVolumes is mock for LVM.QueryPhysicalVolumes()
func (f *FakeLVM) QueryPhysicalVolumes() (map[string]*lowlevel.PhysicalVolume, error) {
	c := f.Called()
	return c.Get(0).(map[string]*lowlevel.PhysicalVolume), c.Error(1)
}

// QueryVolumeGroups is mock for LVM.QueryVolumeGroups()
func (f *FakeLVM) QueryVolumeGroups() (map[string]*lowlevel.VolumeGroup, error) {
	c := f.Called()
	return c.Get(0).(map[string]*lowlevel.VolumeGroup), c.Error(1)
}

// CreateVolumeGroup is mock for LVM.CreateVolumeGroup()
func (f *FakeLVM) CreateVolumeGroup(vg string, devs []string) error {
	return f.Called(vg, devs).Error(0)
}

// ExtendVolumeGroup is mock for LVM.ExtendVolumeGroup()
func (f *FakeLVM) ExtendVolumeGroup(vg string, dev string) error {
	return f.Called(vg, dev).Error(0)
}

// ReduceVolumeGroup is mock for LVM.ReduceVolumeGroup()
func (f *FakeLVM) ReduceVolumeGroup(vg string, dev string) error {
	return f.Called(vg, dev).Error(0)
}

// CreatePhysicalVolume is mock for LVM.CreatePhysicalVolume()
func (f *FakeLVM) CreatePhysicalVolume(dev string) error {
	return f.Called(dev).Error(0)
}

// CreateLogicalVolume is mock for LVM.CreateLogicalVolume()
func (f *FakeLVM) CreateLogicalVolume(group string, volume string, size *lowlevel.LvmSize) error {
	return f.Called(group, volume, size).Error(0)
}

// RemovePhysicalVolume is mock for LVM.RemovePhysicalVolume()
func (f *FakeLVM) RemovePhysicalVolume(dev string, force bool) error {
	return f.Called(dev, force).Error(0)
}

// Mkfs is mock for LVM.Mkfs()
func (f *FakeLVM) Mkfs(dev string, fstype string) error {
	return f.Called(dev, fstype).Error(0)
}

// Mountpoint is mock for LVM.Mountpoint()
func (f *FakeLVM) Mountpoint(path string) (bool, error) {
	c := f.Called(path)
	return c.Bool(0), c.Error(1)
}

// Blkid is mock for LVM.Blkid()
func (f *FakeLVM) Blkid(dev string) (string, error) {
	c := f.Called(dev)
	return c.String(0), c.Error(1)
}

// WaitForDevice is mock for LVM.WaitForDevice()
func (f *FakeLVM) WaitForDevice(path string) error {
	return f.Called(path).Error(0)
}

// systemd units

// CheckUnit is mock for LVM.CheckUnit()
func (f *FakeLVM) CheckUnit(filename string, content string) (bool, error) {
	c := f.Called(filename, content)
	return c.Bool(0), c.Error(1)
}

// UpdateUnit is mock for LVM.UpdateUnit()
func (f *FakeLVM) UpdateUnit(filename string, content string) error {
	c := f.Called(filename, content)
	return c.Error(0)
}

// StartUnit is mock for LVM.StartUnit()
func (f *FakeLVM) StartUnit(unitname string) error {
	return f.Called(unitname).Error(0)
}
