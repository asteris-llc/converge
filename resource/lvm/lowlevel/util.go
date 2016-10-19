package lowlevel

import (
	"fmt"

	"github.com/pkg/errors"
)

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
	CreateLogicalVolume(group string, volume string, size int64, sizeOption string, sizeUnit string) error
	Mkfs(dev string, fstype string) error
	Mountpoint(path string) (bool, error)
	Blkid(dev string) (string, error)

	// systemd units
	CheckUnit(filename string, content string) (bool, error)
	UpdateUnit(filename string, content string) error

	// FIXME: possible unneeded
	QueryDeviceMapperName(dmName string) (string, error)

	GetBackend() Exec // FIXME: abstraction leak
}

type RealLVM struct {
	Backend Exec
}

func MakeLvmBackend() LVM {
	backend := MakeOsExec()
	return &RealLVM{Backend: backend}
}

// MakeRealLVM is actually kludge for DI (intended for creating test-backed RealLVM, and unpublish type inself)
func MakeRealLVM(backend Exec) LVM {
	return &RealLVM{Backend: backend}
}

// FIXME: remove when no more used
func (lvm *RealLVM) GetBackend() Exec {
	return lvm.Backend
}

func (lvm *RealLVM) CreateVolumeGroup(vg string, devs []string) error {
	args := []string{vg}
	args = append(args, devs...)
	return lvm.Backend.Run("vgcreate", args)
}

func (lvm *RealLVM) ExtendVolumeGroup(vg string, dev string) error {
	return lvm.Backend.Run("vgextend", []string{vg, dev})
}

func (lvm *RealLVM) ReduceVolumeGroup(vg string, dev string) error {
	return lvm.Backend.Run("vgreduce", []string{vg, dev})
}

func (lvm *RealLVM) CreatePhysicalVolume(dev string) error {
	return lvm.Backend.Run("pvcreate", []string{dev})
}

func (lvm *RealLVM) CreateLogicalVolume(group string, volume string, size int64, sizeOption string, sizeUnit string) error {
	sizeStr := fmt.Sprintf("%d%s", size, sizeUnit)
	option := fmt.Sprintf("-%s", sizeOption)
	return lvm.Backend.Run("lvcreate", []string{"-n", volume, option, sizeStr, group})
}

func (lvm *RealLVM) Mkfs(dev string, fstype string) error {
	return lvm.Backend.Run("mkfs", []string{"-t", fstype, dev})
}

func (lvm *RealLVM) Mountpoint(path string) (bool, error) {
	rc, err := lvm.Backend.RunExitCode("mountpoint", []string{"-q", path})
	if err != nil {
		return false, err
	}
	if rc == 0 {
		return true, nil
	}
	return false, nil
}

func (lvm *RealLVM) Check() error {
	if uid := lvm.Backend.Getuid(); uid != 0 {
		return fmt.Errorf("lvm require root permissions (uid == 0), but converge run from user id (uid == %d)", uid)
	}
	// FIXME: extend list to all used tools or wrap all calls via `lvm $subcommand` and check for lvm only
	//        second way need careful check, if `lvm $subcommand` and just `$subcommand`  accepot exact same parameters
	for _, tool := range []string{"lvs", "vgs", "pvs", "lvcreate", "lvreduce", "lvremove", "vgcreate", "vgreduce", "pvcreate"} {
		if err := lvm.Backend.Lookup(tool); err != nil {
			return errors.Wrapf(err, "lvm: can't find required tool %s in $PATH", tool)
		}
	}
	return nil
}

func (lvm *RealLVM) CheckFilesystemTools(fstype string) error {
	// Root check just copied from .Check() because lvm.fs can be used w/o lvm utils,  but require root and mkfs.*
	if uid := lvm.Backend.Getuid(); uid != 0 {
		return fmt.Errorf("lvm require root permissions (uid == 0), but converge run from user id (uid == %d)", uid)
	}

	tool := fmt.Sprintf("mkfs.%s", fstype)
	if err := lvm.Backend.Lookup(tool); err != nil {
		return errors.Wrapf(err, "lvm: can't find required tool %s in $PATH", tool)
	}
	return nil
}
