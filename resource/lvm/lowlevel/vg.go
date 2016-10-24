package lowlevel

// VolumeGroup is parsed record for LVM Volume Groups (from `vgs` output)
// Add more fields, if required
// (at the moment we need only LVM2_VG_NAME to get list all existing groups)
type VolumeGroup struct {
	Name string `mapstructure:"LVM2_VG_NAME"`
}

func (lvm *realLVM) QueryVolumeGroups() (map[string]*VolumeGroup, error) {
	result := map[string]*VolumeGroup{}
	vgs, err := lvm.Query("vgs", "all", []string{})
	if err != nil {
		return nil, err
	}
	for _, values := range vgs {
		vg := &VolumeGroup{}
		if err = lvm.parse(values, vg); err != nil {
			return nil, err
		}
		result[vg.Name] = vg
	}
	return result, nil
}
