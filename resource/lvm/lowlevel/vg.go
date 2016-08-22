package lowlevel

type VG struct {
	Name string `mapstructure:"LVM2_VG_NAME"`
}

func QueryVG() (map[string]*VG, error) {
	result := map[string]*VG{}
	vgs, err := queryLVM("vgs", "all", []string{})
	if err != nil {
		return nil, err
	}
	for _, values := range vgs {
		vg := &VG{}
		if err = parseLVM(&vg, values); err != nil {
			return nil, err
		}
		result[vg.Name] = vg
	}
	return result, nil
}
