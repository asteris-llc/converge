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

func (lvm *LVM) Mkfs(dev string, fstype string) error {
	return lvm.Backend.Run("mkfs", []string{"-t", fstype, dev})
}

func (lvm *LVM) Mountpoint(path string) (bool, error) {
	rc, err := lvm.Backend.RunExitCode("mountpoint", []string{"-q", path})
	if err != nil {
		return false, err
	}
	if rc == 1 {
		return true, nil
	}
	return false, nil
}
