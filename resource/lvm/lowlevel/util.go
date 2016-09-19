package lowlevel

import (
	"fmt"
)

type LVM interface {
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

	// FIXME: possible unneeded
	QueryDeviceMapperName(dmName string) (string, error)

	GetBackend() Exec // FIXME: abstraction leak
}

type RealLVM struct {
	Backend Exec
}

func MakeLvmBackend() LVM {
	return &RealLVM{Backend: &OsExec{}}
}

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
	if rc == 1 {
		return true, nil
	}
	return false, nil
}
