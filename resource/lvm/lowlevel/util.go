package lowlevel

type LVM struct {
	Backend Exec
}

func MakeLvmBackend() *LVM {
	return &LVM{Backend: &OsExec{}}
}

func (lvm *LVM) CreateVolumeGroup(vg string, devs []string) error {
	args := []string{vg}
	args = append(args, devs...)
	return lvm.Backend.Run("vgcreate", args)
}

func (lvm *LVM) ExtendVolumeGroup(vg string, dev string) error {
	return lvm.Backend.Run("vgextend", []string{vg, dev})
}

func (lvm *LVM) ReduceVolumeGroup(vg string, dev string) error {
	return lvm.Backend.Run("vgreduce", []string{vg, dev})
}

func (lvm *LVM) CreatePhysicalVolume(dev string) error {
	return lvm.Backend.Run("pvcreate", []string{dev})
}
