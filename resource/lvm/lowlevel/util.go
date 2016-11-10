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

package lowlevel

import (
	"fmt"
	"path/filepath"

	"github.com/asteris-llc/converge/resource/wait"
	"github.com/pkg/errors"
)

// LVM is a public interface to LVM guts for converge highlevel modules
type LVM interface {
	// Check for LVM tools installed and useable
	Check() error

	// Check for mkfs.* tools installed and useable
	CheckFilesystemTools(fstype string) error

	QueryLogicalVolumes(vg string) (map[string]*LogicalVolume, error)
	QueryPhysicalVolumes() (map[string]*PhysicalVolume, error)
	QueryVolumeGroups() (map[string]*VolumeGroup, error)
	CreateVolumeGroup(vg string, devs []string) error
	ExtendVolumeGroup(vg string, dev string) error
	ReduceVolumeGroup(vg string, dev string) error
	CreatePhysicalVolume(dev string) error
	RemovePhysicalVolume(dev string, force bool) error
	CreateLogicalVolume(group string, volume string, size *LvmSize) error
	Mkfs(dev string, fstype string) error
	Mountpoint(path string) (bool, error)
	Blkid(dev string) (string, error)
	WaitForDevice(path string) error

	// systemd units
	CheckUnit(filename string, content string) (bool, error)
	UpdateUnit(filename string, content string) error
	StartUnit(filename string) error
}

type realLVM struct {
	backend Exec
}

// MakeLvmBackend creates default LVM backend
func MakeLvmBackend() LVM {
	backend := MakeOsExec()
	return &realLVM{backend: backend}
}

// MakeRealLVM is actually kludge for DI (intended for creating test-backed RealLVM, and unpublish type inself)
func MakeRealLVM(backend Exec) LVM {
	return &realLVM{backend: backend}
}

func (lvm *realLVM) CreateVolumeGroup(vg string, devs []string) error {
	args := []string{vg}
	var canonicalDevs []string
	for _, dev := range devs {
		canonicalDev, err := lvm.backend.EvalSymlinks(dev)
		if err != nil {
			return errors.Wrap(err, "resolving symlink for "+dev)
		}
		canonicalDevs = append(canonicalDevs, canonicalDev)
	}
	args = append(args, canonicalDevs...)
	return lvm.backend.Run("vgcreate", args)
}

func (lvm *realLVM) ExtendVolumeGroup(vg string, dev string) error {
	canonicalDev, err := lvm.backend.EvalSymlinks(dev)
	if err != nil {
		return err
	}
	return lvm.backend.Run("vgextend", []string{vg, canonicalDev})
}

func (lvm *realLVM) ReduceVolumeGroup(vg string, dev string) error {
	canonicalDev, err := lvm.backend.EvalSymlinks(dev)
	if err != nil {
		return err
	}
	return lvm.backend.Run("vgreduce", []string{vg, canonicalDev})
}

func (lvm *realLVM) CreatePhysicalVolume(dev string) error {
	canonicalDev, err := lvm.backend.EvalSymlinks(dev)
	if err != nil {
		return err
	}
	return lvm.backend.Run("pvcreate", []string{canonicalDev})
}

func (lvm *realLVM) RemovePhysicalVolume(dev string, force bool) error {
	canonicalDev, err := lvm.backend.EvalSymlinks(dev)
	if err != nil {
		return err
	}
	args := []string{}
	if force {
		args = append(args, "--force", "--force", "--yes")
	}
	args = append(args, canonicalDev)
	return lvm.backend.Run("pvremove", args)
}

func (lvm *realLVM) CreateLogicalVolume(group string, volume string, size *LvmSize) error {
	sizeStr := size.String()
	option := size.Option()
	return lvm.backend.Run("lvcreate", []string{"-n", volume, option, sizeStr, group})
}

func (lvm *realLVM) Mkfs(dev string, fstype string) error {
	canonicalDev, err := lvm.backend.EvalSymlinks(dev)
	if err != nil {
		return err
	}
	return lvm.backend.Run("mkfs", []string{"-t", fstype, canonicalDev})
}

func (lvm *realLVM) Mountpoint(path string) (bool, error) {
	rc, err := lvm.backend.RunWithExitCode("mountpoint", []string{"-q", path})
	if err != nil {
		return false, err
	}
	if rc == 0 {
		return true, nil
	}
	return false, nil
}

func (lvm *realLVM) Check() error {
	if uid := lvm.backend.Getuid(); uid != 0 {
		return fmt.Errorf("lvm require root permissions (uid == 0), but converge run from user id (uid == %d)", uid)
	}
	// NB: extend list to all used tools or wrap all calls via `lvm $subcommand` and check for lvm only
	//     second way need careful check, if `lvm $subcommand` and just `$subcommand`  accepot exact same parameters
	// Related issue: https://github.com/asteris-llc/converge/issues/457
	for _, tool := range []string{"lvs", "vgs", "pvs", "lvcreate", "lvreduce", "lvremove", "vgcreate", "vgreduce", "pvcreate"} {
		if err := lvm.backend.Lookup(tool); err != nil {
			return errors.Wrapf(err, "lvm: can't find required tool %s in $PATH", tool)
		}
	}
	return nil
}

func (lvm *realLVM) CheckFilesystemTools(fstype string) error {
	// Root check just copied from .Check() because lvm.fs can be used w/o lvm utils,  but require root and mkfs.*
	if uid := lvm.backend.Getuid(); uid != 0 {
		return fmt.Errorf("lvm require root permissions (uid == 0), but converge run from user id (uid == %d)", uid)
	}

	tool := fmt.Sprintf("mkfs.%s", fstype)
	if err := lvm.backend.Lookup(tool); err != nil {
		return errors.Wrapf(err, "lvm: can't find required tool %s in $PATH", tool)
	}
	return nil
}

func (lvm *realLVM) WaitForDevice(path string) error {
	retrier := wait.PrepareRetrier(nil, nil, nil)
	ok, err := retrier.RetryUntil(func() (bool, error) {
		return lvm.backend.Exists(path)
	})
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("device path %s not appeared after %s seconds", path, retrier.Duration.String())
	}
	return nil
}

// evalDeviceSymlinks returns the real path of deach device (otherwise it breaks
// on GCE)
func evalDeviceSymlinks(devices []string) ([]string, error) {
	realpaths := make([]string, len(devices))
	for idx, dev := range devices {
		realpath, err := filepath.EvalSymlinks(dev)
		if err != nil {
			return realpaths, errors.Wrap(err, "unable to resolve path: "+dev)
		}
		realpaths[idx] = realpath
	}
	return realpaths, nil
}
