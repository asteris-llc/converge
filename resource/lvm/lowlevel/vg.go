package lowlevel

type VolumeGroup struct {
	Name string `mapstructure:"LVM2_VG_NAME"`
}

func QueryVolumeGroups() (map[string]*VolumeGroup, error) {
	result := map[string]*VolumeGroup{}
	vgs, err := queryLVM("vgs", "all", []string{})
	if err != nil {
		return nil, err
	}
	for _, values := range vgs {
		vg := &VolumeGroup{}
		if err = parseLVM(&vg, values); err != nil {
			return nil, err
		}
		result[vg.Name] = vg
	}
	return result, nil
}
