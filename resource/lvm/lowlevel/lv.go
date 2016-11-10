package lowlevel

type LogicalVolume struct {
	Name       string `mapstructure:"LVM2_LV_NAME"`
	DevicePath string `mapstructure:"LVM2_LV_DM_PATH"`
}

func (lvm *realLVM) QueryLogicalVolumes(vg string) (map[string]*LogicalVolume, error) {
	result := map[string]*LogicalVolume{}
	lvs, err := lvm.Query("lvs", "all", []string{vg})
	if err != nil {
		return nil, err
	}
	for _, values := range lvs {
		lv := &LogicalVolume{}
		if err = lvm.parse(values, lv); err != nil {
			return nil, err
		}
		result[lv.Name] = lv
	}
	return result, nil
}
